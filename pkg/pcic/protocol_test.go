package pcic_test

import (
	"bufio"
	"compress/bzip2"
	"embed"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/graugans/go-ovp8xx/pkg/pcic"
	"github.com/stretchr/testify/assert"
)

const miniMalContentLength int = 14

//go:embed testdata/*.bz2
var tfs embed.FS

type PCICAsyncReceiver struct {
	frame           pcic.Frame
	notificationMsg pcic.NotificationMessage
	errorMsg        pcic.ErrorMessage
}

func (r *PCICAsyncReceiver) Result(frame pcic.Frame) {
	r.frame = frame
}

func (r *PCICAsyncReceiver) Error(msg pcic.ErrorMessage) {
	r.errorMsg = msg
}

func (r *PCICAsyncReceiver) Notification(msg pcic.NotificationMessage) {
	r.notificationMsg = msg
}

var testHandler *PCICAsyncReceiver = &PCICAsyncReceiver{}

func TestMinimalReceive(t *testing.T) {
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader("Hello, Reader!")),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(t, err, "We expect an error while receiving malformed data")

	// Test the minimal possible PCIC message
	readerWriter = bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader("0000L000000014\r\n0000starstop\r\n")),
		nil,
	)
	p, err = pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.NoError(t, err, "We expect no error while receiving data")
}

func TestNotMatchingTickets(t *testing.T) {
	// Test the minimal possible PCIC message
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader("0001L000000014\r\n0000starstop\r\n")),
		nil,
	)
	client, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = client.ProcessIncomming(testHandler)
	assert.Error(t,
		err,
		"We expect an error because the tickets do not match",
	)
}

func TestMalformedLength(t *testing.T) {
	// Test the minimal possible PCIC message
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader("0000l000000014\r\n0000starstop\r\n")),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(t,
		err,
		"We expect an error because the length field does not start with `L`",
	)
}

func TestMalformedLengthField(t *testing.T) {
	// Test the minimal possible PCIC message
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader("0000L00000014X\r\n0000starstop\r\n")),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(t,
		err,
		"We expect an error because the length field is no well formed",
	)
}

func TestMinimumLengthField(t *testing.T) {
	// Test the minimal possible PCIC message
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader("0000L000000005\r\n0000starstop\r\n")),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(t,
		err,
		"We expect an error because the length is too short",
	)
}

func TestBiggerLengthField(t *testing.T) {
	// Test the minimal possible PCIC message
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader("0000L000000015\r\n0000starstop\r\n")),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(t,
		err,
		"We expect an error because the length too big",
	)
}

func TestInvalidTrailer(t *testing.T) {
	// Test the minimal possible PCIC message
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader("0000L000000014\r\n0000starstop\r\r")),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(t,
		err,
		"We expect an error because the trailer is invalid",
	)
}
func TestWithNilReader(t *testing.T) {
	readerWriter := bufio.NewReadWriter(
		nil,
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(t,
		err,
		"We expect an error when reader is nil",
	)
}

func TestReceiveWithChunk(t *testing.T) {
	c := pcic.Chunk{}
	chunkData := []byte{
		0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
		0x34, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
		0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
		0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
		0x04, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
		0x01, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
		0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
		0x00, 0x00, 0x00, 0x00, /* TIME_STAMP */
		0x00, 0x00, 0x00, 0x00, /* FRAME_COUNT */
		0x00, 0x00, 0x00, 0x00, /* STATUS_CODE */
		0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
		0x01, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		0xFF, 0xFF, 0xFF, 0xBB, /* DATA */
	}
	assert.NoError(t,
		c.UnmarshalBinary(chunkData),
		"A successful parse expected",
	)

	buffer := fmt.Sprintf(
		"0000L%09d\r\n0000star%sstop\r\n",
		miniMalContentLength+len(chunkData),
		string(chunkData),
	)
	// Test the PCIC message with single chunk
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader(buffer)),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.NoError(t, err, "We expect no error while receiving data")

	assert.Equal(t,
		pcic.RADIAL_DISTANCE_NOISE,
		testHandler.frame.Chunks[0].Type(),
	)

	// test with trailing XX after the chunk
	buffer = fmt.Sprintf(
		"0000L%09d\r\n0000star%sXXstop\r\n",
		miniMalContentLength+len(chunkData)+2,
		string(chunkData),
	)
	// Test the PCIC message with single chunk
	readerWriter = bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader(buffer)),
		nil,
	)
	p, err = pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(t, err, "We expect an error while receiving malformed data")

	// test with invalid ticket after the chunk
	buffer = fmt.Sprintf(
		"0002L%09d\r\n0002star%sstop\r\n",
		miniMalContentLength+len(chunkData),
		string(chunkData),
	)
	readerWriter = bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader(buffer)),
		nil,
	)
	p, err = pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(
		t,
		err,
		"We expect an error while receiving data with an invalid ticket",
	)

}

func TestReceiveWithNewChunk(t *testing.T) {
	chunk := pcic.NewChunk(
		pcic.WithChunkType(pcic.RADIAL_DISTANCE_NOISE),
		pcic.WithDimension(100, 100, pcic.FORMAT_16U),
	)
	chunkData, err := chunk.MarshalBinary()
	assert.NoError(t,
		err,
		"We do not expect an error while marshalling the Chunk to binary",
	)

	buffer := fmt.Sprintf(
		"0000L%09d\r\n0000star%sstop\r\n",
		miniMalContentLength+len(chunkData),
		string(chunkData),
	)
	// Test the PCIC message with single chunk
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader(buffer)),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.NoError(t,
		err,
		"We expect no error while receiving chunk data",
	)
}

func TestWithRealChunkData(t *testing.T) {
	file, err := tfs.Open("testdata/pcic-test-data.blob.bz2")
	assert.NoError(t, err, "No error expected while reading the input")
	defer file.Close()
	buf := bufio.NewReader(file)
	cr := bzip2.NewReader(buf)
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(cr),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	for {
		err := p.ProcessIncomming(testHandler)
		if errors.Is(err, io.EOF) {
			break
		}
		assert.NoError(t, err, "No error expected while reading the compressed input")
		fmt.Print("Chunks: [ ")
		for _, c := range testHandler.frame.Chunks {
			fmt.Printf("%d, ", c.Type())
		}
		fmt.Println("]")
	}

}

func TestWithMalformedErrorData(t *testing.T) {
	buffer := fmt.Sprintf(
		"0001L%09d\r\n0001000000000:\r\n",
		16,
	)
	// Test the PCIC message with error message
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader(buffer)),
		nil,
	)
	p, err := pcic.NewPCICClient(
		pcic.WithBufioReaderWriter(readerWriter),
	)
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.Error(t,
		err,
		"We expect an error while receiving malformed data",
	)
}
func TestWithErrorData(t *testing.T) {
	buffer := fmt.Sprintf(
		"0001L%09d\r\n0001000000000:{}\r\n",
		18,
	)
	// Test the PCIC message with error message
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(strings.NewReader(buffer)),
		nil,
	)
	p, err := pcic.NewPCICClient(
		pcic.WithBufioReaderWriter(readerWriter),
	)
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	err = p.ProcessIncomming(testHandler)
	assert.NoError(t,
		err,
		"We expect no error while receiving data",
	)
}

func TestWithRealErrorData(t *testing.T) {
	file, err := tfs.Open("testdata/pcic-diagnostic.blob.bz2")
	assert.NoError(t, err, "No error expected while reading the input")
	defer file.Close()
	buf := bufio.NewReader(file)
	cr := bzip2.NewReader(buf)
	readerWriter := bufio.NewReadWriter(
		bufio.NewReader(cr),
		nil,
	)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(readerWriter))
	assert.NoError(t, err, "We expect no error while creating the PCICClient")
	for {
		err := p.ProcessIncomming(testHandler)
		if errors.Is(err, io.EOF) {
			break
		}
		assert.NoError(t, err, "No error expected while reading the compressed input")
		assert.NotEqual(
			t,
			0, /* 0 means no Error, what does not make sense */
			testHandler.errorMsg.ID,
			"An invalid error ID received",
		)
	}

}
