package vnc

import "io"

type Conn interface {
	io.ReadWriteCloser
	Protocol() string
	PixelFormat() *PixelFormat
	ColorMap() *ColorMap
	SetColorMap(*ColorMap)
	Encodings() []Encoding
	Width() uint16
	Height() uint16
	SetWidth(uint16)
	SetHeight(uint16)
	DesktopName() string
	Flush() error
}
