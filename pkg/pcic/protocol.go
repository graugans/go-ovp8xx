package pcic

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

type (
	PCICClient struct {
		reader *bufio.Reader
		writer *bufio.Writer
	}
	PCICClientOption func(c *PCICClient)
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
	resultTicket       []byte = []byte{'0', '0', '0', '0'}
	errorTicket        []byte = []byte{'0', '0', '0', '1'}
	notificationTicket []byte = []byte{'0', '0', '1', '0'}
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

func NewPCICClient(options ...PCICClientOption) *PCICClient {
	pcic := &PCICClient{}
	// Apply options
	for _, opt := range options {
		opt(pcic)
	}
	return pcic
}

func WithBufioReaderWriter(com *bufio.ReadWriter) PCICClientOption {
	return func(c *PCICClient) {
		c.reader = com.Reader
		c.writer = com.Writer
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
	secondTicket := header[secondTicketOffset:dataOffset]
	if !bytes.Equal(firstTicket, secondTicket) {
		return fmt.Errorf("mismatch in the tickets %s != %s ",
			string(firstTicket),
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
	if bytes.Equal(resultTicket, firstTicket) {
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
	fmt.Printf("Async Data received\n")
	frame := Frame{}
	var err error
	contentDecorated := data[:len(data)-delimiterFieldLength]
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
