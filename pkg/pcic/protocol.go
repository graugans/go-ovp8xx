package pcic

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/graugans/go-ovp8xx/pkg/chunk"
)

type PCIC struct {
}

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

func (p *PCIC) Receive(reader io.Reader) (Frame, error) {
	frame := Frame{}
	header := make([]byte, headerSize)
	n, err := io.ReadFull(reader, header)
	if err != nil {
		return frame, err
	}
	if n < headerSize {
		return frame, fmt.Errorf("not enough data received: %d", n)
	}
	firstTicket := header[:ticketFieldLength]
	secondTicket := header[secondTicketOffset:dataOffset]
	if !bytes.Equal(firstTicket, secondTicket) {
		return frame, fmt.Errorf("mismatch in the tickets %s != %s ",
			string(firstTicket),
			string(secondTicket),
		)
	}
	lengthBuffer := string(header[lengthOffset:secondTicketOffset])
	if lengthBuffer[0] != 'L' {
		return frame, fmt.Errorf("the length field does not start with 'L': %v", lengthBuffer)
	}
	length := 0
	n, err = fmt.Sscanf(lengthBuffer, "L%09d\r\n", &length)
	if err != nil {
		return frame, err
	}
	if n != 1 {
		return frame, errors.New("no length in the length field detected")
	}
	if length < minimumContentLength {
		return frame, errors.New("the length information is too short")
	}
	data := make([]byte, length-ticketFieldLength)
	if _, err = io.ReadFull(reader, data); err != nil {
		return frame, err
	}
	trailer := data[len(data)-delimiterFieldLength:]
	if !bytes.Equal(trailer, []byte{'\r', '\n'}) {
		return frame, errors.New("invalid trailer detected")
	}
	contentDecorated := data[:len(data)-delimiterFieldLength]
	if len(startMarker)+len(endMarker) > len(contentDecorated) {
		return frame, fmt.Errorf("missing start (%s) and end markers (%s) buffer length: %d",
			startMarker,
			endMarker,
			len(contentDecorated),
		)
	}
	content := contentDecorated[len(endMarker) : len(contentDecorated)-len(endMarker)]
	if len(content) == 0 {
		// no content is available
		return frame, nil
	}
	remainingBytes := len(content)
	offset := 0
	for remainingBytes > 0 {
		c := chunk.ChunkData{}
		if err := c.Parse(content[offset:]); err != nil {
			return frame, err
		}
		frame.Chunks = append(frame.Chunks, c)
		offset += c.Size()
		remainingBytes -= c.Size()
	}
	return frame, err
}
