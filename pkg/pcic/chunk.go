package pcic

import (
	"encoding/binary"
	"errors"
	"fmt"
	"time"
)

type (
	ChunkType   uint32
	DataFormat  uint32
	ChunkOption func(c *Chunk)
)

// The known Chunk Types
const (
	RADIAL_DISTANCE_NOISE ChunkType = 105
)

// The known Data Formats
const (
	FORMAT_8U  DataFormat = 0 /* 8bit unsigned integer */
	FORMAT_8S  DataFormat = 1 /* 8bit signed integer */
	FORMAT_16U DataFormat = 2 /* 16bit unsigned integer*/
	FORMAT_16S DataFormat = 3 /* 16bit signed integer  */
	FORMAT_32U DataFormat = 4 /* 32bit unsigned integer*/
	FORMAT_32S DataFormat = 5 /* 32bit signed integer  */
	FORMAT_32F DataFormat = 6 /* 32bit floating point number */
	FORMAT_64U DataFormat = 7 /* 64bit unsigned integer*/
	FORMAT_64F DataFormat = 8 /* 64bit floating point number*/
	FORMAT_MAX DataFormat = 9 /* The maximum known data type*/
)

var (
	byteSizeLUT [FORMAT_MAX]uint32 = [FORMAT_MAX]uint32{1, 1, 2, 2, 4, 4, 4, 8, 8}
)

// A Chunk object contains all the information based on a PCIC chunk
type Chunk struct {
	chunkType     ChunkType  // The type of the Chunk, each chunk type requires a unique ID
	chunkSize     uint32     // The size of the complete chunk, including the header and the data
	headerSize    uint32     // The Size of the chunk header after this amount of bytes the data section starts
	headerVersion uint32     // The version of the header
	dataWidth     uint32     // The width of the data
	dataHeight    uint32     // The height of the data, for none image data this is set to 1
	dataFormat    DataFormat // The data format
	timeStamp     uint32     // The timestamp in micro seconds (deprecated)
	frameCount    uint32     // A frame count
	statusCode    uint32     // Conveys the status of the device default: 0
	timestampSec  uint32     // The timestamp seconds part
	timestampNSec uint32     // The timestamp nano seconds part
	metadata      string     // The JSON meta data is always {} for v2 chunks
	data          []byte     // The data the chunk describes
}

const (
	offsetOfType          = 0x0000
	offsetOfSize          = 0x0004
	offsetOfHeaderSize    = 0x0008
	offsetOfHeaderVersion = 0x000C
	offsetOfWidth         = 0x0010
	offsetOfHeight        = 0x0014
	offsetOfFormat        = 0x0018
	offsetOfTimeStamp     = 0x001C
	offsetOfFrameCount    = 0x0020
	offsetOfStatusCode    = 0x0024
	offsetOfTimeStampSec  = 0x0028
	offsetOfTimeStampNsec = 0x002C
	offsetOfData          = 0x0030
)

const (
	MaxSupportedChunkHeaderVersion = 3
)

func NewChunk(options ...ChunkOption) *Chunk {
	chunk := &Chunk{
		metadata: "{}",
		data:     []byte{},
	}
	// Apply options
	for _, opt := range options {
		opt(chunk)
	}
	return chunk
}

func WithChunkType(cType ChunkType) ChunkOption {
	return func(c *Chunk) {
		c.chunkType = cType
	}
}

func (c *Chunk) Type() ChunkType {
	return c.chunkType
}

func (c *Chunk) Size() int {
	return int(c.chunkSize)
}

func (c *Chunk) FrameCount() uint32 {
	return c.frameCount
}

func (c *Chunk) Status() uint32 {
	return c.statusCode
}

func (c *Chunk) TimeStamp() time.Time {
	return time.Unix(int64(c.timestampSec), int64(c.timestampNSec))
}

func (c *Chunk) Bytes() []byte {
	return c.data
}

// MarshalBinary creates a binary representation of the Chunk
//
// The binary representation is encoded in the byte slice
func (c *Chunk) MarshalBinary() (data []byte, err error) {
	return []byte{}, nil
}

// UnmarshalBinary creates a ChunkData from a byte slice
//
// It copies the data from the input slice to comply with the BinaryUnmarshaler
// interface.
func (c *Chunk) UnmarshalBinary(data []byte) error {
	dataLen := uint32(len(data))
	if dataLen < offsetOfData {
		return errors.New("unable to parse an empty input")
	}
	c.chunkType = ChunkType(
		binary.LittleEndian.Uint32(data[offsetOfType : offsetOfType+4]),
	)
	c.chunkSize = binary.LittleEndian.Uint32(
		data[offsetOfSize : offsetOfSize+4],
	)
	if c.chunkSize < offsetOfData {
		return fmt.Errorf("the chunk size needs to be at minimum: %d", offsetOfData)
	}
	if c.chunkSize > dataLen {
		return fmt.Errorf(
			"the chunk size expected is: %d but the data is only: %d",
			c.chunkSize,
			dataLen,
		)
	}
	c.headerSize = binary.LittleEndian.Uint32(
		data[offsetOfHeaderSize : offsetOfHeaderSize+4],
	)
	if c.headerSize < offsetOfData {
		return fmt.Errorf("the chunk header size needs to be at minimum: %d", offsetOfData)
	}
	c.headerVersion = binary.LittleEndian.Uint32(
		data[offsetOfHeaderVersion : offsetOfHeaderVersion+4],
	)
	if c.headerVersion == 2 && c.headerSize > offsetOfData {
		return fmt.Errorf(
			"the chunk header size expected is: %d but the expected maximum is only: %d",
			c.headerSize,
			offsetOfData,
		)
	}

	if c.headerVersion == 0 || c.headerVersion > MaxSupportedChunkHeaderVersion {
		return fmt.Errorf("invalid chunk header version given: %d maximum supported version: %d",
			c.headerVersion,
			MaxSupportedChunkHeaderVersion,
		)
	}
	c.dataWidth = binary.LittleEndian.Uint32(
		data[offsetOfWidth : offsetOfWidth+4],
	)
	c.dataHeight = binary.LittleEndian.Uint32(
		data[offsetOfHeight : offsetOfHeight+4],
	)
	if (c.dataHeight * c.dataWidth) > (dataLen - c.headerSize) {
		return fmt.Errorf(
			"the length of the given data can not be smaller than the given data width and height multiplied",
		)
	}
	c.dataFormat = DataFormat(binary.LittleEndian.Uint32(
		data[offsetOfFormat : offsetOfFormat+4],
	))
	if c.dataFormat > FORMAT_MAX {
		return fmt.Errorf(
			"the the data format does not match the range of valid data formats [0,%d]",
			FORMAT_MAX,
		)
	}

	c.timeStamp = binary.LittleEndian.Uint32(
		data[offsetOfTimeStamp : offsetOfTimeStamp+4],
	)

	c.frameCount = binary.LittleEndian.Uint32(
		data[offsetOfFrameCount : offsetOfFrameCount+4],
	)

	c.statusCode = binary.LittleEndian.Uint32(
		data[offsetOfStatusCode : offsetOfStatusCode+4],
	)

	c.timestampSec = binary.LittleEndian.Uint32(
		data[offsetOfTimeStampSec : offsetOfTimeStampSec+4],
	)

	c.timestampNSec = binary.LittleEndian.Uint32(
		data[offsetOfTimeStampNsec : offsetOfTimeStampNsec+4],
	)

	// Copy the data to this chunk
	src := data[offsetOfData : offsetOfData+(c.chunkSize-c.headerSize)]
	c.data = make([]byte, len(src))
	copy(c.data, src)
	if src == nil {
		c.data = nil
	}

	if (c.dataWidth * c.dataHeight * byteSizeLUT[c.dataFormat]) != uint32(len(c.data)) {
		return fmt.Errorf(
			"a size mismatch detected, width (%d) times height (%d) does not equal the data size (%d) format: %d",
			c.dataWidth,
			c.dataHeight,
			len(c.data),
			c.dataFormat,
		)

	}

	return nil
}
