package vnc2video

import (
	"encoding/binary"
	"io"
	//"image/draw"
)

type RREEncoding struct {
	//Colors []Color
	numSubRects     uint32
	backgroundColor []byte
	subRectData     []byte
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

func (z *RREEncoding) Type() int32 {
	return 2
}
func (z *RREEncoding) Read(r Conn, rect *Rectangle) error {
	//func (z *RREEncoding) Read(pixelFmt *PixelFormat, rect *Rectangle, r io.Reader) (Encoding, error) {
	bytesPerPixel := int(r.PixelFormat().BPP / 8)

	var numOfSubrectangles uint32
	if err := binary.Read(r, binary.BigEndian, &numOfSubrectangles); err != nil {
		return err
	}

	var err error
	z.numSubRects = numOfSubrectangles

	//read whole-rect background color
	z.backgroundColor, err = ReadBytes(bytesPerPixel, r)
	if err != nil {
		return err
	}

	//read all individual rects (color=bytesPerPixel + x=16b + y=16b + w=16b + h=16b)
	z.subRectData, err = ReadBytes(int(numOfSubrectangles)*(bytesPerPixel+8), r) // x+y+w+h=8 bytes
	if err != nil {
		return err
	}

	return nil
}
