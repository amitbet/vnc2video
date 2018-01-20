package vnc2video

import (
	"image/draw"
)

type RawEncoding struct {
	Image draw.Image
	//Colors []Color
}

func (*RawEncoding) Supported(Conn) bool {
	return true
}

func (*RawEncoding) Reset() error {
	return nil
}

func (enc *RawEncoding) Write(c Conn, rect *Rectangle) error {
	var err error

	return err
}
func (enc *RawEncoding) SetTargetImage(img draw.Image) {
	enc.Image = img
}

// Read implements the Encoding interface.
func (enc *RawEncoding) Read(c Conn, rect *Rectangle) error {
	pf := c.PixelFormat()

	DecodeRaw(c, &pf, rect, enc.Image)

	return nil
}

func (*RawEncoding) Type() EncodingType { return EncRaw }
