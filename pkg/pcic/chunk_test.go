package pcic_test

import (
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
