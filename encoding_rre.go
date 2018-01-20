package vnc2video

import (
	"encoding/binary"
	"image/draw"
	"io"
	//"image/draw"
)

type RREEncoding struct {
	//Colors []Color
	numSubRects     uint32
	backgroundColor []byte
	subRectData     []byte
	Image           draw.Image
}

func (*RREEncoding) Supported(Conn) bool {
	return true
}

func (enc *RREEncoding) SetTargetImage(img draw.Image) {
	enc.Image = img
}

func (enc *RREEncoding) Reset() error {
	return nil
}

func (*RREEncoding) Type() EncodingType { return EncRRE }

func (enc *RREEncoding) Write(c Conn, rect *Rectangle) error {

	return nil
}

func (z *RREEncoding) WriteTo(w io.Writer) (n int, err error) {
	binary.Write(w, binary.BigEndian, z.numSubRects)
	if err != nil {
		return 0, err
	}

	w.Write(z.backgroundColor)
	if err != nil {
		return 0, err
	}

	w.Write(z.subRectData)

	if err != nil {
		return 0, err
	}
	b := len(z.backgroundColor) + len(z.subRectData) + 4
	return b, nil
}

func (enc *RREEncoding) Read(r Conn, rect *Rectangle) error {
	//func (z *RREEncoding) Read(pixelFmt *PixelFormat, rect *Rectangle, r io.Reader) (Encoding, error) {
	pf := r.PixelFormat()
	//bytesPerPixel := int(pf.BPP / 8)

	var numOfSubrectangles uint32
	if err := binary.Read(r, binary.BigEndian, &numOfSubrectangles); err != nil {
		return err
	}

	var err error
	enc.numSubRects = numOfSubrectangles

	//read whole-rect background color
	bgColor, err := ReadColor(r, &pf)
	if err != nil {
		return err
	}
	imgRect := MakeRectFromVncRect(rect)
	FillRect(enc.Image, &imgRect, bgColor)

	//read all individual rects (color=bytesPerPixel + x=16b + y=16b + w=16b + h=16b)

	for i := 0; i < int(numOfSubrectangles); i++ {
		color, err := ReadColor(r, &pf)
		if err != nil {
			return err
		}

		x, err := ReadUint16(r)
		if err != nil {
			return err
		}

		y, err := ReadUint16(r)
		if err != nil {
			return err
		}

		width, err := ReadUint16(r)
		if err != nil {
			return err
		}
		height, err := ReadUint16(r)
		if err != nil {
			return err
		}
		subRect := MakeRect(int(rect.X+x), int(rect.Y+y), int(width), int(height))
		FillRect(enc.Image, &subRect, color)
	}

	return nil
}
