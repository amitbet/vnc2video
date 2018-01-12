package vnc2video

import (
	"bytes"
	"encoding/binary"
	"io"
)

type ZLibEncoding struct {
	bytes []byte
}

func (z *ZLibEncoding) Type() int32 {
	return 6
}
func (z *ZLibEncoding) WriteTo(w io.Writer) (n int, err error) {
	return w.Write(z.bytes)
}
func (z *ZLibEncoding) Read(r Conn, rect *Rectangle) error {
	//func (z *ZLibEncoding) Read(pixelFmt *PixelFormat, rect *Rectangle, r io.Reader) (Encoding, error) {
	//conn := RfbReadHelper{Reader:r}
	//conn := &DataSource{conn: conn.c, PixelFormat: conn.PixelFormat}
	//bytesPerPixel := c.PixelFormat.BPP / 8
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
