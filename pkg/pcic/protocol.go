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
	"strings"
	"sync"
)

type (
	PCICClient struct {
		reader        *bufio.Reader
		writer        *bufio.Writer
		responseChans map[string]chan Response
		mu            sync.Mutex
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

var (
	resultTicket []byte = []byte{'0', '0', '0', '0'}
	errorTicket  []byte = []byte{'0', '0', '0', '1'}
)

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
	var err error
	pcic := &PCICClient{}
	// Apply options
	for _, opt := range options {
		if err = opt(pcic); err != nil {
			return nil, err
		}
	}
	return pcic, err
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
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
		if err != nil {
			return err
		}
		c.reader = bufio.NewReader(conn)
		c.writer = bufio.NewWriter(conn)
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

func (p *PCICClient) ProcessIncomming(handler MessageHandler) error {
	reader := p.reader
	if reader == nil {
		return errors.New("no bufio.Reader provided, please instantiate the object")
	}
	header := make([]byte, headerSize)
	_, err := io.ReadFull(reader, header)
	if err != nil {
		return err
	}
	firstTicket := header[:ticketFieldLength]
	ticketStr := string(firstTicket)

	secondTicket := header[secondTicketOffset:dataOffset]
	if !bytes.Equal(firstTicket, secondTicket) {
		return fmt.Errorf("mismatch in the tickets %s != %s ",
			string(ticketStr),
			string(secondTicket),
		)
	}
	lengthBuffer := string(header[lengthOffset:secondTicketOffset])
	if lengthBuffer[0] != 'L' {
		return fmt.Errorf("the length field does not start with 'L': %v", lengthBuffer)
	}
	length := 0
	_, err = fmt.Sscanf(lengthBuffer, "L%09d\r\n", &length)
	if err != nil {
		return err
	}
	if length < minimumContentLength {
		return errors.New("the length information is too short")
	}
	data := make([]byte, length-ticketFieldLength)
	if _, err = io.ReadFull(reader, data); err != nil {
		return err
	}
	trailer := data[len(data)-delimiterFieldLength:]
	if !bytes.Equal(trailer, []byte{'\r', '\n'}) {
		return errors.New("invalid trailer detected")
	}
	var ticketNum = 0
	if ticketStr != "0000" {
		ticketNum, err = strconv.Atoi(strings.TrimLeft(ticketStr, "0"))
		if err != nil {
			return fmt.Errorf("unable to convert the ticket number %s to an integer", ticketStr)
		}
	}
	if ticketNum > 100 {
		err := p.responseParser(ticketStr, data)
		if err != nil {
			return fmt.Errorf("unable to parse the response: %w", err)
		}
		return nil
	} else if bytes.Equal(resultTicket, firstTicket) {
		frame, err := asyncResultParser(data)
		handler.Result(frame)
		return err
	} else if bytes.Equal(errorTicket, firstTicket) {
		errorStatus, err := errorParser(data)
		handler.Error(errorStatus)
		return err
	}
	return fmt.Errorf("unknown ticket received: %s", string(firstTicket))
}

func (p *PCICClient) Send(ctx context.Context, data []byte) ([]byte, error) {
	var res []byte
	if p.writer == nil {
		return res, errors.New("no bufio.Writer provided, please instantiate the object")
	}
	respChan := make(chan Response)
	// Let's generate a random ticket number
	ticket := p.generateTicket(respChan)

	defer func() {
		p.mu.Lock()
		delete(p.responseChans, ticket)
		p.mu.Unlock()
	}()

	// Create a new buffer to aggregate the message
	var buf bytes.Buffer
	var delimter = []byte("\r\n")
	// Convert ticket to a 4-digit string and then to bytes
	ticketBytes := []byte(ticket)
	buf.Write(ticketBytes)
	// A Command message is composed like this
	// <ticket><length>CRLF<ticket><content>CRLF
	// <ticket> is a 4-digit number in the range 100-9999
	// <length> is a character string starting with an 'L' followed by 9 digits
	// interpreted as a decimal value. The number is the length of data that follows
	// <content> is the actual data that is being sent
	length := len(ticketBytes) /*<ticket>*/ + len(data) + 2 /*CRLF*/
	lengthStr := fmt.Sprintf("L%09d", length)
	lengthBytes := []byte(lengthStr)
	buf.Write(lengthBytes)
	buf.Write(delimter)
	buf.Write(ticketBytes)
	buf.Write(data)
	buf.Write(delimter)

	// Write the buffer to the underlying writer
	_, err := p.writer.Write(buf.Bytes())
	if err != nil {
		return res, fmt.Errorf("unable to write to the buffer: %w", err)
	}
	// This is necessary to flush the buffer to the underlying writer
	// Otherwise, the data will not be sent over the network
	err = p.writer.Flush()
	if err != nil {
		return res, fmt.Errorf("unable to flush to the buffer: %w", err)
	}

	// Wait for the response or timeout
	select {
	case resp := <-respChan:
		res = resp.Data
	case <-ctx.Done():
		return res, fmt.Errorf("request with ticket: %s timed out or canceled", ticket)
	}

	return res, nil
}

func (p *PCICClient) responseParser(ticket string, data []byte) error {
	var err error
	res := Response{}
	if len(data) <= delimiterFieldLength {
		return fmt.Errorf("the data is too short to be a valid frame: %d", len(data))
	}
	res.Ticket = ticket
	res.Data = data[:len(data)-delimiterFieldLength]

	p.mu.Lock()
	if ch, ok := p.responseChans[ticket]; ok {
		ch <- res
	}
	p.mu.Unlock()
	return err
}

func errorParser(data []byte) (ErrorMessage, error) {
	var err error
	errorStatus := ErrorMessage{}
	n, err := fmt.Sscanf(
		string(data),
		"%09d:%s",
		&errorStatus.ID,
		&errorStatus.Message,
	)
	if n != 2 {
		return ErrorMessage{}, errors.New("unable to parse the error message")
	}
	return errorStatus, err
}

func asyncResultParser(data []byte) (Frame, error) {
	frame := Frame{}
	var err error
	if len(data) <= delimiterFieldLength {
		return frame, fmt.Errorf("the data is too short to be a valid frame: %d", len(data))
	}
	contentDecorated := data[:len(data)-delimiterFieldLength]
	if len(contentDecorated)-len(endMarker) < 0 {
		return frame, fmt.Errorf("the data is too short to be a valid frame: %d: content: %s", len(data), string(data))
	}
	content := contentDecorated[len(endMarker) : len(contentDecorated)-len(endMarker)]
	if len(content) == 0 {
		// no content is available
		return frame, nil
	}
	remainingBytes := len(content)
	offset := 0
	for remainingBytes > 0 {
		c := Chunk{}
		if err := c.UnmarshalBinary(content[offset:]); err != nil {
			return frame, err
		}
		frame.Chunks = append(frame.Chunks, c)
		offset += c.Size()
		remainingBytes -= c.Size()

	}
	return frame, err

}
