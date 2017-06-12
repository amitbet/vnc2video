// Implementation of RFC 6143 §7.4 Pixel Format Data Structure.

package vnc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

var (
	PixelFormat8bit  *PixelFormat = NewPixelFormat(8)
	PixelFormat16bit *PixelFormat = NewPixelFormat(16)
	PixelFormat32bit *PixelFormat = NewPixelFormat(32)
)

// PixelFormat describes the way a pixel is formatted for a VNC connection.
type PixelFormat struct {
	BPP                             uint8   // bits-per-pixel
	Depth                           uint8   // depth
	BigEndian                       uint8   // big-endian-flag
	TrueColor                       uint8   // true-color-flag
	RedMax, GreenMax, BlueMax       uint16  // red-, green-, blue-max (2^BPP-1)
	RedShift, GreenShift, BlueShift uint8   // red-, green-, blue-shift
	_                               [3]byte // padding
}

const pixelFormatLen = 16

// NewPixelFormat returns a populated PixelFormat structure.
func NewPixelFormat(bpp uint8) *PixelFormat {
	bigEndian := uint8(1)
	rgbMax := uint16(math.Exp2(float64(bpp))) - 1
	var (
		tc         = uint8(1)
		rs, gs, bs uint8
	)
	switch bpp {
	case 8:
		tc = 0
		rs, gs, bs = 0, 0, 0
	case 16:
		rs, gs, bs = 0, 4, 8
	case 32:
		rs, gs, bs = 0, 8, 16
	}
	return &PixelFormat{bpp, bpp, bigEndian, tc, rgbMax, rgbMax, rgbMax, rs, gs, bs, [3]byte{}}
}

// Marshal implements the Marshaler interface.
func (pf *PixelFormat) Marshal() ([]byte, error) {
	// Validation checks.
	switch pf.BPP {
	case 8, 16, 32:
	default:
		return nil, fmt.Errorf("Invalid BPP value %v; must be 8, 16, or 32.", pf.BPP)
	}

	if pf.Depth < pf.BPP {
		return nil, fmt.Errorf("Invalid Depth value %v; cannot be < BPP", pf.Depth)
	}
	switch pf.Depth {
	case 8, 16, 32:
	default:
		return nil, fmt.Errorf("Invalid Depth value %v; must be 8, 16, or 32.", pf.Depth)
	}

	// Create the slice of bytes
	buf := bytes.NewBuffer(nil)
	if err := binary.Write(buf, binary.BigEndian, &pf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Read reads from an io.Reader, and populates the PixelFormat.
func (pf *PixelFormat) Read(r io.Reader) error {
	buf := make([]byte, pixelFormatLen)
	if _, err := io.ReadAtLeast(r, buf, pixelFormatLen); err != nil {
		return err
	}
	return pf.Unmarshal(buf)
}

// Unmarshal implements the Unmarshaler interface.
func (pf *PixelFormat) Unmarshal(data []byte) error {
	buf := bytes.NewBuffer(data)

	var msg PixelFormat
	if err := binary.Read(buf, binary.BigEndian, &msg); err != nil {
		return err
	}
	*pf = msg

	return nil
}

// String implements the fmt.Stringer interface.
func (pf *PixelFormat) String() string {
	return fmt.Sprintf("{ bpp: %d depth: %d big-endian: %s true-color: %s red-max: %d green-max: %d blue-max: %d red-shift: %d green-shift: %d blue-shift: %d }",
		pf.BPP, pf.Depth, pf.BigEndian, pf.TrueColor, pf.RedMax, pf.GreenMax, pf.BlueMax, pf.RedShift, pf.GreenShift, pf.BlueShift)
}

func (pf *PixelFormat) order() binary.ByteOrder {
	if pf.BigEndian == 1 {
		return binary.BigEndian
	}
	return binary.LittleEndian
}
