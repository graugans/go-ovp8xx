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

	"github.com/graugans/go-ovp8xx/pkg/chunk"
	"github.com/graugans/go-ovp8xx/pkg/pcic"
	"github.com/stretchr/testify/assert"
)

const miniMalContentLength int = 14

//go:embed testdata/*.bz2
var tfs embed.FS

func TestMinimalReceive(t *testing.T) {
	r := strings.NewReader("Hello, Reader!")
	p := pcic.PCIC{}
	_, err := p.Receive(r)
	assert.Error(t, err, "We expect an error while receiving malformed data")

	// Test the minimal possible PCIC message
	r = strings.NewReader("0001L000000014\r\n0001starstop\r\n")
	_, err = p.Receive(r)
	assert.NoError(t, err, "We expect no error while receiving data")
}

func TestReceiveWithChunk(t *testing.T) {
	c := chunk.ChunkData{}
	chunkData := []byte{
		0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
		0x34, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
		0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
		0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
		0x04, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
		0x01, 0x00, 0x00, 0x00, /* IMAGE_HEIGTH */
		0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
		0x00, 0x00, 0x00, 0x00, /* TIME_STAMP */
		0x00, 0x00, 0x00, 0x00, /* FRAME_COUNT */
		0x00, 0x00, 0x00, 0x00, /* STATUS_CODE */
		0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
		0x01, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		0xFF, 0xFF, 0xFF, 0xBB, /* DATA */
	}
	assert.NoError(t,
		c.Parse(chunkData),
		"A successful parse expected",
	)
	p := pcic.PCIC{}
	buffer := fmt.Sprintf(
		"0001L%09d\r\n0001star%sstop\r\n",
		miniMalContentLength+len(chunkData),
		string(chunkData),
	)
	// Test the PCIC message with single chunk
	r := strings.NewReader(buffer)
	f, err := p.Receive(r)
	assert.NoError(t, err, "We expect no error while receiving data")

	assert.Equal(t, chunk.RADIAL_DISTANCE_NOISE, f.Chunks[0].Type())

	// test with trailing XX after the chunk
	buffer = fmt.Sprintf(
		"0001L%09d\r\n0001star%sXXstop\r\n",
		miniMalContentLength+len(chunkData)+2,
		string(chunkData),
	)
	// Test the PCIC message with single chunk
	r = strings.NewReader(buffer)
	_, err = p.Receive(r)
	assert.Error(t, err, "We expect an error while receiving malformed data")

}

func TestWithRealData(t *testing.T) {
	file, err := tfs.Open("testdata/pcic-test-data.blob.bz2")
	assert.NoError(t, err, "No error expected while reading the input")
	defer file.Close()
	buf := bufio.NewReader(file)
	cr := bzip2.NewReader(buf)
	p := pcic.PCIC{}
	for {
		f, err := p.Receive(cr)
		if errors.Is(err, io.EOF) {
			break
		}
		assert.NoError(t, err, "No error expected while reading the compressed input")
		fmt.Print("Chunks: [ ")
		for _, c := range f.Chunks {
			fmt.Printf("%d, ", c.Type())
		}
		fmt.Println("]")
	}

}
