package pcic_test

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"runtime"
	"testing"

	"github.com/graugans/go-ovp8xx/v2/pkg/pcic"
	"github.com/stretchr/testify/assert"
)

// chunkHeader describes the fields of a chunk header used to craft test input.
type chunkHeader struct {
	chunkType, chunkSize, headerSize, headerVersion uint32
	width, height, format                           uint32
}

// buildChunk returns the binary representation of a chunk with the given
// header. The bytes between offset 0x30 and headerSize are filled with meta,
// followed by payload.
func buildChunk(h chunkHeader, meta, payload []byte) []byte {
	buf := make([]byte, 0x30)
	binary.LittleEndian.PutUint32(buf[0x00:], h.chunkType)
	binary.LittleEndian.PutUint32(buf[0x04:], h.chunkSize)
	binary.LittleEndian.PutUint32(buf[0x08:], h.headerSize)
	binary.LittleEndian.PutUint32(buf[0x0C:], h.headerVersion)
	binary.LittleEndian.PutUint32(buf[0x10:], h.width)
	binary.LittleEndian.PutUint32(buf[0x14:], h.height)
	binary.LittleEndian.PutUint32(buf[0x18:], h.format)
	buf = append(buf, meta...)
	return append(buf, payload...)
}

// pcicMessage wraps content into a PCIC message with the given ticket.
func pcicMessage(ticket string, content []byte) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%sL%09d\r\n%s", ticket, len(ticket)+len(content)+2, ticket)
	b.Write(content)
	b.WriteString("\r\n")
	return b.Bytes()
}

func newTestClient(t *testing.T, input []byte) *pcic.PCICClient {
	t.Helper()
	rw := bufio.NewReadWriter(bufio.NewReader(bytes.NewReader(input)), nil)
	p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(rw))
	assert.NoError(t, err)
	return p
}

func TestUnmarshalDataFormatOutOfRange(t *testing.T) {
	// FORMAT_MAX itself is not a valid data format and must not be used as an
	// index into the byte size lookup table.
	data := buildChunk(chunkHeader{
		chunkType: 1, chunkSize: 0x31, headerSize: 0x30, headerVersion: 2,
		width: 1, height: 1, format: uint32(pcic.FORMAT_MAX),
	}, nil, []byte{0})
	c := pcic.Chunk{}
	assert.NotPanics(t, func() {
		assert.Error(t, c.UnmarshalBinary(data))
	})
}

func TestUnmarshalHeaderSizeBiggerThanChunkSize(t *testing.T) {
	data := buildChunk(chunkHeader{
		chunkType: 1, chunkSize: 0x30, headerSize: 0x34, headerVersion: 3,
	}, []byte("{}  "), nil)
	c := pcic.Chunk{}
	assert.NotPanics(t, func() {
		assert.Error(t, c.UnmarshalBinary(data))
	})
}

func TestUnmarshalHeaderV3WithMetadata(t *testing.T) {
	// A version 3 header may be bigger than 0x30 bytes, the additional bytes
	// carry JSON meta data. The payload starts right after the header.
	meta := []byte(`{"a":1}`)
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	headerSize := uint32(0x30 + len(meta))
	data := buildChunk(chunkHeader{
		chunkType:     1,
		chunkSize:     headerSize + uint32(len(payload)),
		headerSize:    headerSize,
		headerVersion: 3,
		width:         4, height: 1, format: uint32(pcic.FORMAT_8U),
	}, meta, payload)
	c := pcic.Chunk{}
	assert.NoError(t, c.UnmarshalBinary(data))
	assert.Equal(t, payload, c.Bytes(), "the payload must start after the header")
	assert.Equal(t, string(meta), c.Metadata())
}

func TestUnmarshalDimensionOverflow(t *testing.T) {
	// width * height overflows uint32 to exactly 0, which matches the empty
	// payload if the multiplication is done in 32 bit.
	data := buildChunk(chunkHeader{
		chunkType: 1, chunkSize: 0x30, headerSize: 0x30, headerVersion: 2,
		width: 0x10000, height: 0x10000, format: uint32(pcic.FORMAT_8U),
	}, nil, nil)
	c := pcic.Chunk{}
	assert.Error(t, c.UnmarshalBinary(data))
}

func TestReceiveTooShortFrame(t *testing.T) {
	// "stars" is long enough for the start marker but not for start and stop
	p := newTestClient(t, pcicMessage("0000", []byte("stars")))
	assert.NotPanics(t, func() {
		assert.Error(t, p.ProcessIncomming(testHandler))
	})
}

func TestReceiveInvalidMarkers(t *testing.T) {
	p := newTestClient(t, pcicMessage("0000", []byte("xxxxyyyy")))
	assert.Error(t, p.ProcessIncomming(testHandler))
}

func TestReceiveExcessiveLength(t *testing.T) {
	// A corrupted length field must not lead to a huge allocation.
	p := newTestClient(t, []byte("0000L999999999\r\n0000star"))
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	assert.Error(t, p.ProcessIncomming(testHandler))
	runtime.ReadMemStats(&after)
	assert.Less(t, after.TotalAlloc-before.TotalAlloc, uint64(1<<20), "no huge allocation expected")
}

type recordingHandler struct {
	results       int
	errors        []pcic.ErrorMessage
	notifications []pcic.NotificationMessage
}

func (r *recordingHandler) Result(pcic.Frame)         { r.results++ }
func (r *recordingHandler) Error(m pcic.ErrorMessage) { r.errors = append(r.errors, m) }
func (r *recordingHandler) Notification(m pcic.NotificationMessage) {
	r.notifications = append(r.notifications, m)
}

func TestReceiveNotification(t *testing.T) {
	p := newTestClient(t, pcicMessage("0010", []byte(`000500000:{"state":"RUN"}`)))
	h := &recordingHandler{}
	assert.NoError(t, p.ProcessIncomming(h))
	if assert.Len(t, h.notifications, 1) {
		assert.Equal(t, 500000, h.notifications[0].ID)
		assert.Equal(t, `{"state":"RUN"}`, h.notifications[0].Message)
	}
}

func TestResultHandlerNotCalledOnError(t *testing.T) {
	// The chunk claims to be bigger than the frame
	chunk := buildChunk(chunkHeader{
		chunkType: 1, chunkSize: 0x40, headerSize: 0x30, headerVersion: 2,
	}, nil, nil)
	content := append([]byte("star"), chunk...)
	content = append(content, []byte("stop")...)
	p := newTestClient(t, pcicMessage("0000", content))
	h := &recordingHandler{}
	assert.Error(t, p.ProcessIncomming(h))
	assert.Equal(t, 0, h.results, "an invalid frame must not be passed to the handler")
}

func TestReceivedChunksDoNotAlias(t *testing.T) {
	// Chunks of consecutive frames must not share memory.
	payload1 := []byte{1, 1, 1, 1}
	payload2 := []byte{2, 2, 2, 2}
	frame := func(payload []byte) []byte {
		c := buildChunk(chunkHeader{
			chunkType: 1, chunkSize: 0x34, headerSize: 0x30, headerVersion: 2,
			width: 4, height: 1, format: uint32(pcic.FORMAT_8U),
		}, nil, payload)
		return pcicMessage("0000", append(append([]byte("star"), c...), []byte("stop")...))
	}
	input := append(frame(payload1), frame(payload2)...)
	p := newTestClient(t, input)
	var frames []pcic.Frame
	h := &frameCollector{frames: &frames}
	assert.NoError(t, p.ProcessIncomming(h))
	assert.NoError(t, p.ProcessIncomming(h))
	assert.Equal(t, payload1, frames[0].Chunks[0].Bytes())
	assert.Equal(t, payload2, frames[1].Chunks[0].Bytes())
}

type frameCollector struct {
	recordingHandler
	frames *[]pcic.Frame
}

func (f *frameCollector) Result(fr pcic.Frame) { *f.frames = append(*f.frames, fr) }

func FuzzChunkUnmarshal(f *testing.F) {
	f.Add(buildChunk(chunkHeader{1, 0x30, 0x30, 2, 0, 0, 0}, nil, nil))
	f.Add(buildChunk(chunkHeader{1, 0x34, 0x30, 2, 4, 1, 0}, nil, []byte{1, 2, 3, 4}))
	f.Add(buildChunk(chunkHeader{1, 0x36, 0x32, 3, 4, 1, 0}, []byte("{}"), []byte{1, 2, 3, 4}))
	f.Fuzz(func(t *testing.T, data []byte) {
		c := pcic.Chunk{}
		if err := c.UnmarshalBinary(data); err != nil {
			return
		}
		if c.Size() > len(data) {
			t.Fatalf("chunk size %d exceeds input %d", c.Size(), len(data))
		}
	})
}

func FuzzProcessIncoming(f *testing.F) {
	f.Add([]byte("0000L000000014\r\n0000starstop\r\n"))
	f.Add(pcicMessage("0001", []byte("000000000:{}")))
	f.Add(pcicMessage("0010", []byte("000500000:{}")))
	f.Add(pcicMessage("0000", append(append([]byte("star"),
		buildChunk(chunkHeader{1, 0x34, 0x30, 2, 4, 1, 0}, nil, []byte{1, 2, 3, 4})...),
		[]byte("stop")...)))
	f.Fuzz(func(t *testing.T, data []byte) {
		rw := bufio.NewReadWriter(bufio.NewReader(bytes.NewReader(data)), nil)
		p, err := pcic.NewPCICClient(pcic.WithBufioReaderWriter(rw))
		if err != nil {
			t.Fatal(err)
		}
		for i := 0; i < 4; i++ {
			if p.ProcessIncomming(&recordingHandler{}) != nil {
				return
			}
		}
	})
}

func TestRoundtripHeaderV3WithMetadata(t *testing.T) {
	meta := []byte(`{"a":1}`)
	payload := []byte{1, 2, 3, 4}
	headerSize := uint32(0x30 + len(meta))
	data := buildChunk(chunkHeader{
		chunkType:     1,
		chunkSize:     headerSize + uint32(len(payload)),
		headerSize:    headerSize,
		headerVersion: 3,
		width:         4, height: 1, format: uint32(pcic.FORMAT_8U),
	}, meta, payload)
	c := pcic.Chunk{}
	assert.NoError(t, c.UnmarshalBinary(data))
	out, err := c.MarshalBinary()
	assert.NoError(t, err)
	assert.Equal(t, data, out, "marshalling must reproduce the input")
}
