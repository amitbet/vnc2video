package vnc2video

import (
	"image"
	"image/draw"
	"github.com/amitbet/vnc2video/logger"
)

type CursorPosPseudoEncoding struct {
	prevPosBackup    draw.Image
	prevPositionRect image.Rectangle
	cursorImage      draw.Image
	Image            draw.Image
}

func (*CursorPosPseudoEncoding) Supported(Conn) bool {
	return true
}

func (enc *CursorPosPseudoEncoding) SetTargetImage(img draw.Image) {
	enc.Image = img
}

func (enc *CursorPosPseudoEncoding) Reset() error {
	return nil
}

func (*CursorPosPseudoEncoding) Type() EncodingType { return EncPointerPosPseudo }

func (enc *CursorPosPseudoEncoding) Read(c Conn, rect *Rectangle) error {
	logger.Tracef("CursorPosPseudoEncoding: got cursot pos update: %v", rect)
	canvas := enc.Image.(*VncCanvas)
	canvas.CursorLocation = &image.Point{X: int(rect.X), Y: int(rect.Y)}
	return nil
}

func (enc *CursorPosPseudoEncoding) Write(c Conn, rect *Rectangle) error {

	return nil
}
