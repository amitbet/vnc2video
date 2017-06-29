package vnc

import (
	"io"
	"net"
)

type Conn interface {
	io.ReadWriteCloser
	Conn() net.Conn
	Config() interface{}
	Protocol() string
	PixelFormat() PixelFormat
	SetPixelFormat(PixelFormat) error
	ColorMap() ColorMap
	SetColorMap(ColorMap)
	Encodings() []Encoding
	SetEncodings([]EncodingType) error
	Width() uint16
	Height() uint16
	SetWidth(uint16)
	SetHeight(uint16)
	DesktopName() []byte
	SetDesktopName([]byte)
	Flush() error
	Wait()
	SetProtoVersion(string)
}
