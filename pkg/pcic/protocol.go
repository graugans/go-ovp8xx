package pcic

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"sync"
)

type (
	PCICClient struct {
		reader        *bufio.Reader
		writer        *bufio.Writer
		closer        io.Closer
		responseChans map[string]chan Response
		mu            sync.Mutex // protects responseChans
		writeMu       sync.Mutex // serializes requests on the writer
		maxFrameSize  int
	}
	PCICClientOption func(c *PCICClient) error
)

const (
	headerSize           int = 20
	minimumContentLength int = 6
	ticketFieldLength    int = 4
	lengthFieldLength    int = 10
	delimiterFieldLength int = 2
)

const (
	firstTicketOffset  int = 0
	lengthOffset       int = 4
	secondTicketOffset int = 16
	delimiterOffset    int = 14
	dataOffset         int = 20
)

const (
	startMarker string = "star"
	endMarker   string = "stop"
)

const (
	resultTicket       string = "0000"
	errorTicket        string = "0001"
	notificationTicket string = "0010"
	// minResponseTicket is the lowest ticket number used for command responses
	minResponseTicket int = 100
)

// DefaultMaxFrameSize is the default upper limit for the size of a single
// PCIC message. It protects against huge allocations caused by a corrupted
// length field.
const DefaultMaxFrameSize int = 64 << 20

type MessageHandler interface {
	Result(Frame)
	Error(ErrorMessage)
	Notification(NotificationMessage)
}

type NotificationMessage struct {
	ID      int
	Message string
}

type ErrorMessage struct {
	ID      int
	Message string
}

type Response struct {
	Ticket string
	Data   []byte
}

func NewPCICClient(options ...PCICClientOption) (*PCICClient, error) {
	pcic := &PCICClient{
		responseChans: make(map[string]chan Response),
		maxFrameSize:  DefaultMaxFrameSize,
	}
	// Apply options
	for _, opt := range options {
		if err := opt(pcic); err != nil {
			return nil, err
		}
	}
	return pcic, nil
}

// WithMaxFrameSize sets the maximum accepted size in bytes of a single PCIC
// message. Messages announcing a bigger length are rejected.
func WithMaxFrameSize(size int) PCICClientOption {
	return func(c *PCICClient) error {
		if size < minimumContentLength {
			return fmt.Errorf("the maximum frame size must be at least %d", minimumContentLength)
		}
		c.maxFrameSize = size
		return nil
	}
}

// Close closes the underlying connection if the client owns one, e.g. when
// created with WithTCPClient.
func (p *PCICClient) Close() error {
	if p.closer == nil {
		return nil
	}
	return p.closer.Close()
}

func WithBufioReaderWriter(com *bufio.ReadWriter) PCICClientOption {
	return func(c *PCICClient) error {
		c.reader = com.Reader
		c.writer = com.Writer
		return nil
	}
}

// WithTCPClient is a PCICClientOption that sets up a TCP client connection to the specified URI.
// It establishes a connection using the net.Dial function and initializes the reader and writer for the PCICClient.
// If an error occurs during the connection establishment, it will be handled and returned.
func WithTCPClient(hostname string, port uint16) PCICClientOption {
	return func(c *PCICClient) error {
		conn, err := net.Dial("tcp", net.JoinHostPort(hostname, strconv.Itoa(int(port))))
		if err != nil {
			return err
		}
		c.reader = bufio.NewReader(conn)
		c.writer = bufio.NewWriter(conn)
		c.closer = conn
		return nil
	}
}

// generateTicket generates a unique ticket and associates it with the provided reply channel.
// It uses a random number generator to generate a 4-digit ticket number.
// If the generated ticket already exists in the responseChans map, it continues generating a new ticket until a unique one is found.
// Once a unique ticket is found, it adds the ticket and reply channel to the responseChans map and returns the ticket.
func (p *PCICClient) generateTicket(replyChan chan Response) string {
	for {
		ticket := fmt.Sprintf("%04d", rand.Intn(8999)+1000)
		p.mu.Lock()
		if _, exists := p.responseChans[ticket]; !exists {
			p.responseChans[ticket] = replyChan
			p.mu.Unlock()
			return ticket
		}
		p.mu.Unlock()
	}
}

// ProcessIncomming reads and dispatches a single PCIC message.
//
// Asynchronous results, errors and notifications are passed to the handler,
// responses to commands are forwarded to the pending Send call. The handler is
// only called for messages which could be parsed successfully.
//
// The frames passed to the handler own their memory, the chunk data is not
// reused by subsequent calls.
func (p *PCICClient) ProcessIncomming(handler MessageHandler) error {
	reader := p.reader
	if reader == nil {
		return errors.New("no bufio.Reader provided, please instantiate the object")
	}
	var header [headerSize]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return err
	}
	firstTicket := header[:ticketFieldLength]
	secondTicket := header[secondTicketOffset:dataOffset]
	if !bytes.Equal(firstTicket, secondTicket) {
		return fmt.Errorf("mismatch in the tickets %s != %s ", firstTicket, secondTicket)
	}
	lengthField := header[lengthOffset:delimiterOffset]
	if lengthField[0] != 'L' {
		return fmt.Errorf("the length field does not start with 'L': %q", lengthField)
	}
	length, ok := parseDecimal(lengthField[1:])
	if !ok {
		return fmt.Errorf("the length field is not a decimal number: %q", lengthField)
	}
	if header[delimiterOffset] != '\r' || header[delimiterOffset+1] != '\n' {
		return errors.New("the length field is not terminated by CRLF")
	}
	if length < minimumContentLength {
		return errors.New("the length information is too short")
	}
	if length > p.maxFrameSize {
		return fmt.Errorf("the length %d exceeds the maximum frame size of %d", length, p.maxFrameSize)
	}
	data := make([]byte, length-ticketFieldLength)
	if _, err := io.ReadFull(reader, data); err != nil {
		return err
	}
	if !bytes.HasSuffix(data, []byte{'\r', '\n'}) {
		return errors.New("invalid trailer detected")
	}
	content := data[:len(data)-delimiterFieldLength]

	ticket := string(firstTicket)
	switch ticket {
	case resultTicket:
		frame, err := asyncResultParser(content)
		if err != nil {
			return err
		}
		handler.Result(frame)
		return nil
	case errorTicket:
		id, msg, err := parseStatusMessage(content)
		if err != nil {
			return fmt.Errorf("unable to parse the error message: %w", err)
		}
		handler.Error(ErrorMessage{ID: id, Message: msg})
		return nil
	case notificationTicket:
		id, msg, err := parseStatusMessage(content)
		if err != nil {
			return fmt.Errorf("unable to parse the notification message: %w", err)
		}
		handler.Notification(NotificationMessage{ID: id, Message: msg})
		return nil
	}
	if ticketNum, ok := parseDecimal(firstTicket); ok && ticketNum >= minResponseTicket {
		p.dispatchResponse(ticket, content)
		return nil
	}
	return fmt.Errorf("unknown ticket received: %q", ticket)
}

// parseDecimal parses an unsigned decimal number consisting only of the digits 0-9.
func parseDecimal(b []byte) (int, bool) {
	if len(b) == 0 {
		return 0, false
	}
	n := 0
	for _, c := range b {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}

// Send sends a command to the device and waits for the response or until the
// context is done. It is safe to call Send from multiple goroutines, while
// ProcessIncomming is running in another goroutine.
func (p *PCICClient) Send(ctx context.Context, data []byte) ([]byte, error) {
	if p.writer == nil {
		return nil, errors.New("no bufio.Writer provided, please instantiate the object")
	}
	// A Command message is composed like this
	// <ticket><length>CRLF<ticket><content>CRLF
	// <ticket> is a 4-digit number in the range 100-9999
	// <length> is a character string starting with an 'L' followed by 9 digits
	// interpreted as a decimal value. The number is the length of data that follows
	// <content> is the actual data that is being sent
	length := ticketFieldLength + len(data) + delimiterFieldLength
	if length > 999999999 {
		return nil, fmt.Errorf("the command is too long: %d bytes", len(data))
	}

	// The channel is buffered, so the receiver never blocks on a response
	// which arrives after the request has timed out.
	respChan := make(chan Response, 1)
	ticket := p.generateTicket(respChan)
	defer func() {
		p.mu.Lock()
		delete(p.responseChans, ticket)
		p.mu.Unlock()
	}()

	if err := p.writeRequest(ticket, length, data); err != nil {
		return nil, err
	}

	// Wait for the response or timeout
	select {
	case resp := <-respChan:
		return resp.Data, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("request with ticket: %s timed out or canceled: %w", ticket, ctx.Err())
	}
}

func (p *PCICClient) writeRequest(ticket string, length int, data []byte) error {
	header := make([]byte, 0, headerSize)
	header = append(header, ticket...)
	header = fmt.Appendf(header, "L%09d\r\n", length)
	header = append(header, ticket...)

	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	if _, err := p.writer.Write(header); err != nil {
		return fmt.Errorf("unable to write to the buffer: %w", err)
	}
	if _, err := p.writer.Write(data); err != nil {
		return fmt.Errorf("unable to write to the buffer: %w", err)
	}
	if _, err := p.writer.WriteString("\r\n"); err != nil {
		return fmt.Errorf("unable to write to the buffer: %w", err)
	}
	// This is necessary to flush the buffer to the underlying writer
	// Otherwise, the data will not be sent over the network
	if err := p.writer.Flush(); err != nil {
		return fmt.Errorf("unable to flush to the buffer: %w", err)
	}
	return nil
}

// dispatchResponse forwards a command response to the waiting Send call.
// Responses for unknown tickets, e.g. for requests which already timed out,
// are dropped.
func (p *PCICClient) dispatchResponse(ticket string, content []byte) {
	p.mu.Lock()
	ch, ok := p.responseChans[ticket]
	p.mu.Unlock()
	if !ok {
		return
	}
	select {
	case ch <- Response{Ticket: ticket, Data: content}:
	default:
		// A response for this ticket is already pending, drop the duplicate
	}
}

// parseStatusMessage parses error and notification messages. Both are
// composed of a 9 digit ID followed by a colon and a non-empty message,
// e.g. "000103012:{...}".
func parseStatusMessage(content []byte) (int, string, error) {
	const idLength = 9
	if len(content) < idLength {
		return 0, "", fmt.Errorf("the message is too short: %q", content)
	}
	id, ok := parseDecimal(content[:idLength])
	if !ok {
		return 0, "", fmt.Errorf("the message does not start with a 9 digit ID: %q", content)
	}
	rest := content[idLength:]
	if len(rest) < 2 || rest[0] != ':' {
		return id, "", fmt.Errorf("a malformed message was received: %q", content)
	}
	return id, string(rest[1:]), nil
}

// asyncResultParser parses the content of an asynchronous result message,
// which is enclosed by the start and stop markers. The chunks reference the
// content slice, which must not be reused by the caller.
func asyncResultParser(content []byte) (Frame, error) {
	frame := Frame{}
	if len(content) < len(startMarker)+len(endMarker) {
		return frame, fmt.Errorf("the data is too short to be a valid frame: %d bytes", len(content))
	}
	if !bytes.HasPrefix(content, []byte(startMarker)) || !bytes.HasSuffix(content, []byte(endMarker)) {
		return frame, errors.New("the frame is not enclosed by the start and stop markers")
	}
	content = content[len(startMarker) : len(content)-len(endMarker)]
	for len(content) > 0 {
		c := Chunk{}
		if err := c.unmarshal(content, false); err != nil {
			return frame, err
		}
		frame.Chunks = append(frame.Chunks, c)
		content = content[c.Size():]
	}
	return frame, nil
}
