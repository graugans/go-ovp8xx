package pcic

import (
	"bytes"
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
		chunkSize:     offsetOfData,
		headerSize:    offsetOfData,
		headerVersion: 2,
		metadata:      "{}",
		data:          []byte{},
	}
	// Apply options
	for _, opt := range options {
		opt(chunk)
	}
	return chunk
}

// WithChunkType sets the chunk type of the new Chunk object
func WithChunkType(cType ChunkType) ChunkOption {
	return func(c *Chunk) {
		c.chunkType = cType
	}
}

// WithDimension set the chunk dimension and the given data format
func WithDimension(width, height int, format DataFormat) ChunkOption {
	return func(c *Chunk) {
		c.dataWidth = uint32(width)
		c.dataHeight = uint32(height)
		c.dataFormat = format
		c.data = make([]byte, c.dataWidth*c.dataHeight*byteSizeLUT[format])
		c.chunkSize = c.headerSize + uint32(len(c.data))
	}
}

// Type returns the given ChunkType
func (c *Chunk) Type() ChunkType {
	return c.chunkType
}

// Size returns the Size of the whole Chunk
//
// This is the size the Chunk is marshalled to.
func (c *Chunk) Size() int {
	return int(c.chunkSize)
}

// FrameCount returns the frame count of the Chunk
func (c *Chunk) FrameCount() uint32 {
	return c.frameCount
}

// SetFrameCount sets the frame count of the Chunk
func (c *Chunk) SetFrameCount(num uint32) {
	c.frameCount = num
}

// Status returns the status of the given Chunk
func (c *Chunk) Status() uint32 {
	return c.statusCode
}

// SetStatus sets the status of the Chunk
func (c *Chunk) SetStatus(status uint32) {
	c.statusCode = status
}

// TimeStamp returns the time stamp of the given Chunk
func (c *Chunk) TimeStamp() time.Time {
	return time.Unix(int64(c.timestampSec), int64(c.timestampNSec))
}

func (c *Chunk) SetTimestamp(value time.Time) {
	c.timestampSec = uint32(value.Unix())
	c.timestampNSec = uint32(value.Nanosecond())
}

// Bytes return the data the current Chunk is holding
func (c *Chunk) Bytes() []byte {
	return c.data
}

// Metadata returns the JSON meta data of the Chunk.
//
// Only version 3 headers carry meta data, for all other chunks "{}" is returned.
func (c *Chunk) Metadata() string {
	return c.metadata
}

// MarshalBinary creates a binary representation of the Chunk
//
// The binary representation is encoded in the byte slice
func (c *Chunk) MarshalBinary() (data []byte, err error) {
	headerSize := c.headerSize
	if headerSize < offsetOfData {
		headerSize = offsetOfData
	}
	blob := make([]byte, int(headerSize)+len(c.data))
	binary.LittleEndian.PutUint32(
		blob,
		uint32(c.chunkType),
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfSize:offsetOfHeaderSize],
		c.chunkSize,
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfHeaderSize:offsetOfHeaderVersion],
		c.headerSize,
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfHeaderVersion:offsetOfWidth],
		c.headerVersion,
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfWidth:offsetOfHeight],
		c.dataWidth,
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfHeight:offsetOfFormat],
		c.dataHeight,
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfFormat:offsetOfTimeStamp],
		uint32(c.dataFormat),
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfTimeStamp:offsetOfFrameCount],
		c.timeStamp,
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfFrameCount:offsetOfStatusCode],
		c.frameCount,
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfStatusCode:offsetOfTimeStampSec],
		c.statusCode,
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfTimeStampSec:offsetOfTimeStampNsec],
		c.timestampSec,
	)
	binary.LittleEndian.PutUint32(
		blob[offsetOfTimeStampNsec:offsetOfData],
		c.timestampNSec,
	)

	// The meta data of a version 3 header is stored between the fixed header
	// fields and the data
	if headerSize > offsetOfData {
		copy(blob[offsetOfData:headerSize], c.metadata)
	}
	copy(blob[headerSize:], c.data)
	return blob, nil
}

// UnmarshalBinary creates a ChunkData from a byte slice
//
// It copies the data from the input slice to comply with the BinaryUnmarshaler
// interface.
func (c *Chunk) UnmarshalBinary(data []byte) error {
	return c.unmarshal(data, true)
}

// unmarshal parses a chunk from data. If copyData is false the chunk data
// references the input slice instead of copying it. This is used when parsing
// a frame, where the input buffer is owned by the frame anyway.
func (c *Chunk) unmarshal(data []byte, copyData bool) error {
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
	if c.headerSize > c.chunkSize {
		return fmt.Errorf(
			"the chunk header size %d exceeds the chunk size %d",
			c.headerSize,
			c.chunkSize,
		)
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
	c.dataFormat = DataFormat(binary.LittleEndian.Uint32(
		data[offsetOfFormat : offsetOfFormat+4],
	))
	if c.dataFormat >= FORMAT_MAX {
		return fmt.Errorf(
			"the data format %d does not match the range of valid data formats [0,%d]",
			c.dataFormat,
			FORMAT_MAX-1,
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

	// A version 3 header carries JSON meta data after the fixed header fields
	c.metadata = "{}"
	if c.headerSize > offsetOfData {
		c.metadata = string(bytes.TrimRight(data[offsetOfData:c.headerSize], "\x00"))
	}

	// The data starts right after the header. Use 64 bit arithmetic to
	// prevent an overflow of width * height * byte size.
	src := data[c.headerSize:c.chunkSize:c.chunkSize]
	expected := uint64(c.dataWidth) * uint64(c.dataHeight) * uint64(byteSizeLUT[c.dataFormat])
	if expected != uint64(len(src)) {
		return fmt.Errorf(
			"a size mismatch detected, width (%d) times height (%d) does not equal the data size (%d) format: %d",
			c.dataWidth,
			c.dataHeight,
			len(src),
			c.dataFormat,
		)
	}
	if copyData {
		c.data = make([]byte, len(src))
		copy(c.data, src)
	} else {
		c.data = src
	}
	return nil
}
