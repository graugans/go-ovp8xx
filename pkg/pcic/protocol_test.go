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
	r := strings.NewReader("Hello, Reader!")
	p := pcic.PCIC{}
	err := p.Receive(r, testHandler)
	assert.Error(t, err, "We expect an error while receiving malformed data")

	// Test the minimal possible PCIC message
	r = strings.NewReader("0000L000000014\r\n0000starstop\r\n")
	err = p.Receive(r, testHandler)
	assert.NoError(t, err, "We expect no error while receiving data")
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
	p := pcic.PCIC{}
	buffer := fmt.Sprintf(
		"0000L%09d\r\n0000star%sstop\r\n",
		miniMalContentLength+len(chunkData),
		string(chunkData),
	)
	// Test the PCIC message with single chunk
	r := strings.NewReader(buffer)
	err := p.Receive(r, testHandler)
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
	r = strings.NewReader(buffer)
	err = p.Receive(r, testHandler)
	assert.Error(t, err, "We expect an error while receiving malformed data")

	// test with invalid ticket after the chunk
	buffer = fmt.Sprintf(
		"0002L%09d\r\n0002star%sstop\r\n",
		miniMalContentLength+len(chunkData),
		string(chunkData),
	)
	r = strings.NewReader(buffer)
	err = p.Receive(r, testHandler)
	assert.Error(
		t,
		err,
		"We expect an error while receiving data with an invalid ticket",
	)

}

func TestWithRealChunkData(t *testing.T) {
	file, err := tfs.Open("testdata/pcic-test-data.blob.bz2")
	assert.NoError(t, err, "No error expected while reading the input")
	defer file.Close()
	buf := bufio.NewReader(file)
	cr := bzip2.NewReader(buf)
	p := pcic.PCIC{}
	for {
		err := p.Receive(cr, testHandler)
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

func TestWithRealErrorData(t *testing.T) {
	file, err := tfs.Open("testdata/pcic-diagnostic.blob.bz2")
	assert.NoError(t, err, "No error expected while reading the input")
	defer file.Close()
	buf := bufio.NewReader(file)
	cr := bzip2.NewReader(buf)
	p := pcic.PCIC{}
	for {
		err := p.Receive(cr, testHandler)
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
