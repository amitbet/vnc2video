package vnc2video

import (
	"encoding/binary"
	"image"
	"image/draw"
	"github.com/amitbet/vnc2video/logger"
)

type CopyRectEncoding struct {
	SX, SY uint16
	Image  draw.Image
}

func (*CopyRectEncoding) Supported(Conn) bool {
	return true
}
func (*CopyRectEncoding) Reset() error {
	return nil
}
func (*CopyRectEncoding) Type() EncodingType { return EncCopyRect }

func (enc *CopyRectEncoding) SetTargetImage(img draw.Image) {
	//logger.Tracef("!!!!!!!!!!!!!setting image: %v", img.Bounds())
	enc.Image = img
}

func (enc *CopyRectEncoding) Read(c Conn, rect *Rectangle) error {
	logger.Tracef("Reading: CopyRect %v", rect)
	if err := binary.Read(c, binary.BigEndian, &enc.SX); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &enc.SY); err != nil {
		return err
	}
	cpyIm := image.NewRGBA(image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{int(rect.Width), int(rect.Height)}})
	for x := 0; x < int(rect.Width); x++ {
		for y := 0; y < int(rect.Height); y++ {
			col := enc.Image.At(x+int(enc.SX), y+int(enc.SY))
			cpyIm.Set(x, y, col)
		}
	}

	for x := 0; x < int(rect.Width); x++ {
		for y := 0; y < int(rect.Height); y++ {
			col := cpyIm.At(x, y)
			enc.Image.Set(int(rect.X)+x, int(rect.Y)+y, col)
		}
	}

	return nil
}

func (enc *CopyRectEncoding) Write(c Conn, rect *Rectangle) error {
	if err := binary.Write(c, binary.BigEndian, enc.SX); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, enc.SY); err != nil {
		return err
	}
	return nil
}
