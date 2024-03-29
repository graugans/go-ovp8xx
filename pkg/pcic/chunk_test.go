package pcic_test

import (
	"bytes"
	crand "crypto/rand"
	"math/rand"
	"testing"
	"time"

	"github.com/graugans/go-ovp8xx/pkg/pcic"
	"github.com/stretchr/testify/assert"
)

func TestChunkType(t *testing.T) {
	c := pcic.NewChunk(pcic.WithChunkType(pcic.RADIAL_DISTANCE_NOISE))
	assert.Equal(t,
		pcic.RADIAL_DISTANCE_NOISE,
		c.Type(),
		"There is a chunk type mismatch detected",
	)
}

func TestUnmarshalBinary(t *testing.T) {
	c := pcic.Chunk{}
	assert.Error(t,
		c.UnmarshalBinary([]byte{}),
		"An error is expected when sending an empty byte slice",
	)
	assert.NoError(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x01, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x01, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"A successful parse expected",
	)
	assert.NoError(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x01, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x01, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"A successful parse expected",
	)
	assert.Equal(t,
		pcic.RADIAL_DISTANCE_NOISE,
		c.Type(),
		"Type mismatch detected",
	)
	assert.Equal(t,
		0x30,
		c.Size(),
		"Size mismatch detected",
	)
	assert.Error(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x01, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x01, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x01, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x01, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"A error due to invalid width and height expected",
	)
	assert.Error(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x01, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x01, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x01, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"A error due to an invalid data format expected",
	)

	assert.NoError(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x01, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x00, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"A successful parse expected",
	)
	assert.Equal(t,
		uint32(0x100),
		c.FrameCount(),
		"A frame count mismatch occurred",
	)
	assert.Equal(t, uint32(0x00), c.Status(), "A status code mismatch occurred")
	assert.NoError(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x00, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x00, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x00, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x01, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"A successful parse expected",
	)
	assert.NotEqual(t,
		time.Unix(0, 0),
		c.TimeStamp(),
		"A timestamp mismatch expected",
	)
	assert.Equal(t,
		time.Unix(int64(0x100), int64(0x101)),
		c.TimeStamp(),
		"A timestamp mismatch occur",
	)
	assert.NoError(t,
		c.UnmarshalBinary([]byte{
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
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE of second frame*/
		}),
		"A successful parse expected",
	)
	assert.Equal(t,
		4,
		len(c.Bytes()),
		"A data size mismatch occurred",
	)
	assert.Equal(t,
		[]byte{0xFF, 0xFF, 0xFF, 0xBB},
		c.Bytes(),
		"A data size mismatch occurred",
	)

	assert.Error(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x34, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x04, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x01, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x04, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x00, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x00, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x00, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x01, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
			0xFF, 0xFF, 0xFF, 0xBB, /* DATA */
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE of second frame*/
		}),
		"The (width * height * data format) does not match the data size",
	)

	assert.Error(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x28, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x04, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x01, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x04, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x00, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x00, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x00, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x01, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
			0xFF, 0xFF, 0xFF, 0xBB, /* DATA */
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE of second frame*/
		}),
		"The Chunk size is smaller than the expected size",
	)

	assert.Error(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x00, 0x01, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x04, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x01, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x04, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x00, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x00, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x00, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x01, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
			0xFF, 0xFF, 0xFF, 0xBB, /* DATA */
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE of second frame*/
		}),
		"The Chunk size is bigger than the data size",
	)

}

func TestRoundtrip(t *testing.T) {
	chunk := pcic.NewChunk(pcic.WithChunkType(pcic.RADIAL_DISTANCE_NOISE))
	chunkData, err := chunk.MarshalBinary()
	assert.NoError(t, err, "No error expected when marshalling to binary")
	roundTripChunk := pcic.NewChunk()
	assert.NoError(t,
		roundTripChunk.UnmarshalBinary(chunkData),
		"No Error expected when unmarshalling from binary",
	)
	assert.Equal(t,
		chunk.Type(),
		roundTripChunk.Type(),
		"We expect the type to be the same after the round trip",
	)
}

func TestHeaderSizeTooSmall(t *testing.T) {
	c := pcic.NewChunk()
	assert.Error(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x28, 0x00, 0x00, 0x00, /* HEADER_SIZE set too small */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x01, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x01, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"An error expected, due to small header size",
	)
}

func TestHeaderSizeTooBig(t *testing.T) {
	c := pcic.NewChunk()
	assert.Error(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x32, 0x00, 0x00, 0x00, /* HEADER_SIZE set too small */
			0x02, 0x00, 0x00, 0x00, /* HEADER_VERSION */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x01, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x01, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"An error expected, due to big header size",
	)
}

func TestInvalidHeaderVersion(t *testing.T) {
	c := pcic.NewChunk()
	assert.Error(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x00, 0x00, 0x00, 0x00, /* HEADER_VERSION == 0 */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x01, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x01, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"An error expected, due to wrong header version",
	)
	assert.Error(t,
		c.UnmarshalBinary([]byte{
			0x69, 0x00, 0x00, 0x00, /* CHUNK_TYPE */
			0x30, 0x00, 0x00, 0x00, /* CHUNK_SIZE */
			0x30, 0x00, 0x00, 0x00, /* HEADER_SIZE */
			0x04, 0x00, 0x00, 0x00, /* HEADER_VERSION == 4 */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_WIDTH */
			0x00, 0x00, 0x00, 0x00, /* IMAGE_HEIGHT */
			0x00, 0x00, 0x00, 0x00, /* DATA_FORMAT */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP */
			0x00, 0x01, 0x00, 0x00, /* FRAME_COUNT */
			0x00, 0x01, 0x00, 0x00, /* STATUS_CODE */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_SEC */
			0x00, 0x01, 0x00, 0x00, /* TIME_STAMP_NSEC */
		}),
		"An error expected, due to wrong header version",
	)
}

func TestWithNil(t *testing.T) {
	vector := pcic.NewChunk()
	assert.Error(
		t,
		vector.UnmarshalBinary(nil),
		"An error is expected when providing nil as input",
	)
}
func TestWithDimension(t *testing.T) {
	vector := make([]*pcic.Chunk, 2)
	vector[0] = pcic.NewChunk(pcic.WithDimension(640, 480, pcic.FORMAT_32F))
	vector[1] = pcic.NewChunk()
	data, err := vector[0].MarshalBinary()
	assert.NoError(t, err, "No Error expected during MarshalBinary")
	assert.NoError(
		t,
		vector[1].UnmarshalBinary(data),
		"No error expected while UnmarshalBinary",
	)
}

func TestFrameCount(t *testing.T) {
	count := rand.Uint32()
	chunk := pcic.NewChunk()
	chunk.SetFrameCount(count)
	assert.Equal(
		t,
		count,
		chunk.FrameCount(),
		"No error is expected when getting the frame count",
	)
	clone := pcic.NewChunk()
	data, err := chunk.MarshalBinary()
	assert.NoError(t,
		err,
		"No error expected when creating a binary clone",
	)
	assert.NoError(t,
		clone.UnmarshalBinary(data),
		"No error expected when creating a clone from the bytes",
	)
	assert.Equal(
		t,
		chunk.FrameCount(),
		clone.FrameCount(),
		"The frame count of the original and the clone do not match",
	)
}

func TestFrameStatus(t *testing.T) {
	status := rand.Uint32()
	chunk := pcic.NewChunk()
	chunk.SetStatus(status)
	assert.Equal(
		t,
		status,
		chunk.Status(),
		"No error expected when setting the staus",
	)
	clone := pcic.NewChunk()
	data, err := chunk.MarshalBinary()
	assert.NoError(t,
		err,
		"No error expected when creating a binary clone",
	)
	assert.NoError(t,
		clone.UnmarshalBinary(data),
		"No error expected when creating a clone from the bytes",
	)
	assert.Equal(
		t,
		chunk.Status(),
		clone.Status(),
		"The status of the original and the clone do not match",
	)
}

func TestChunkTimeStamp(t *testing.T) {
	now := time.Now()
	chunk := pcic.NewChunk()
	chunk.SetTimestamp(now)
	assert.Equal(
		t,
		now.UnixNano(),
		chunk.TimeStamp().UnixNano(),
		"No error expected when setting the time stamp",
	)
	clone := pcic.NewChunk()
	data, err := chunk.MarshalBinary()
	assert.NoError(t,
		err,
		"No error expected when creating a binary clone",
	)
	assert.NoError(t,
		clone.UnmarshalBinary(data),
		"No error expected when creating a clone from the bytes",
	)
	assert.Equal(
		t,
		chunk.TimeStamp(),
		clone.TimeStamp(),
		"The status of the original and the clone do not match",
	)
}

func TestCloneWithData(t *testing.T) {
	chunk := pcic.NewChunk(pcic.WithDimension(2, 1, pcic.FORMAT_16U))
	assert.Equal(t,
		4,
		len(chunk.Bytes()),
		"Size missmatch detected",
	)
	blob := chunk.Bytes()
	_, err := crand.Read(blob)
	assert.NoError(t,
		err,
		"We expect no error when getting random data",
	)
	clone := pcic.NewChunk()
	data, err := chunk.MarshalBinary()
	assert.NoError(t,
		err,
		"No error expected when creating a binary clone",
	)
	assert.NoError(t,
		clone.UnmarshalBinary(data),
		"No error expected when creating a clone from the bytes",
	)
	assert.True(t,
		bytes.Equal(blob, clone.Bytes()),
		"A data mismatch detected",
	)
}
