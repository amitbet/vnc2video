package vnc

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// EncodingType represents a known VNC encoding type.
type EncodingType int32

//go:generate stringer -type=EncodingType

const (
	EncRaw      EncodingType = 0
	EncCopyRect EncodingType = 1
	EncRRE      EncodingType = 2
	EncCoRRE    EncodingType = 4
	EncHextile  EncodingType = 5
	EncZlib     EncodingType = 6
	EncTight    EncodingType = 7
	EncZlibHex  EncodingType = 8
	EncUltra1   EncodingType = 9
	EncUltra2   EncodingType = 10
	EncJPEG     EncodingType = 21
	EncJRLE     EncodingType = 22
	//EncRichCursor        EncodingType = 0xFFFFFF11
	//EncPointerPos        EncodingType = 0xFFFFFF18
	//EncLastRec           EncodingType = 0xFFFFFF20
	EncTRLE              EncodingType = 15
	EncZRLE              EncodingType = 16
	EncColorPseudo       EncodingType = -239
	EncDesktopSizePseudo EncodingType = -223
	EncClientRedirect    EncodingType = -311
)

type Encoding interface {
	Type() EncodingType
	Read(Conn, *Rectangle) error
	Write(Conn, *Rectangle) error
}

type RawEncoding struct {
	Colors []Color
}

func (enc *RawEncoding) Write(c Conn, rect *Rectangle) error {
	/*
		for _, cl := range enc.Colors {
			bytes, err := cl.Marshal()
			if err != nil {
				return err
			}
			if err := binary.Write(c, binary.BigEndian, bytes); err != nil {
				return err
			}
		}
	*/
	return nil

}

// Read implements the Encoding interface.
func (enc *RawEncoding) Read(c Conn, rect *Rectangle) error {
	buf := bytes.NewBuffer(nil)
	pf := c.PixelFormat()
	cm := c.ColorMap()
	bytesPerPixel := int(pf.BPP / 8)
	n := rect.Area() * bytesPerPixel
	data := make([]byte, n)
	fmt.Printf("eeee\n")
	if err := binary.Read(c, binary.BigEndian, &data); err != nil {
		return err
	}
	buf.Write(data)
	defer buf.Reset()
	colors := make([]Color, rect.Area())
	for y := uint16(0); y < rect.Height; y++ {
		for x := uint16(0); x < rect.Width; x++ {
			color := NewColor(pf, cm)
			if err := color.Unmarshal(buf.Next(bytesPerPixel)); err != nil {
				return err
			}
			colors[int(y)*int(rect.Width)+int(x)] = *color
		}
	}

	enc.Colors = colors
	return nil
}

func (*RawEncoding) Type() EncodingType { return EncRaw }

// DesktopSizePseudoEncoding represents a desktop size message from the server.
type DesktopSizePseudoEncoding struct{}

// Read implements the Encoding interface.
func (*DesktopSizePseudoEncoding) Read(c Conn, rect *Rectangle) error {
	c.SetWidth(rect.Width)
	c.SetHeight(rect.Height)

	return nil
}

func (enc *DesktopSizePseudoEncoding) Write(c *ServerConn, rect *Rectangle) error {
	return nil
}
