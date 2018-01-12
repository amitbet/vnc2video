package vnc2video

import (
	"bytes"
	"encoding/binary"
	"io"
)

type ZRLEEncoding struct {
	bytes []byte
}

func (z *ZRLEEncoding) Type() int32 {
	return 16
}

func (z *ZRLEEncoding) WriteTo(w io.Writer) (n int, err error) {
	return w.Write(z.bytes)
}
func (z *ZRLEEncoding) Read(r Conn, rect *Rectangle) error {
	//func (z *ZRLEEncoding) Read(pixelFmt *PixelFormat, rect *Rectangle, r io.Reader) (Encoding, error) {

	bytes := &bytes.Buffer{}
	len, err := ReadUint32(r)
	if err != nil {
		return err
	}

	binary.Write(bytes, binary.BigEndian, len)
	_, err = ReadBytes(int(len), r)
	if err != nil {
		return err
	}
	//StoreBytes(bytes, bts)
	z.bytes = bytes.Bytes()
	return nil
}
