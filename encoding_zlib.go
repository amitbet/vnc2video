package vnc2video

import (
	"bytes"
	"compress/zlib"
	"image/draw"
	"io"
)

type ZLibEncoding struct {
	Image      draw.Image
	unzipper   io.Reader
	zippedBuff *bytes.Buffer
}

func (*ZLibEncoding) Type() EncodingType {
	return EncZlib
}

func (enc *ZLibEncoding) WriteTo(w io.Writer) (n int, err error) {
	return 0, nil
}

func (enc *ZLibEncoding) Write(c Conn, rect *Rectangle) error {
	return nil
}

func (enc *ZLibEncoding) SetTargetImage(img draw.Image) {
	enc.Image = img
}

func (*ZLibEncoding) Supported(Conn) bool {
	return true
}
func (enc *ZLibEncoding) Reset() error {
	enc.unzipper = nil
	return nil
}

func (enc *ZLibEncoding) Read(r Conn, rect *Rectangle) error {
	//func (z *ZLibEncoding) Read(pixelFmt *PixelFormat, rect *Rectangle, r io.Reader) (Encoding, error) {
	//conn := RfbReadHelper{Reader:r}
	//conn := &DataSource{conn: conn.c, PixelFormat: conn.PixelFormat}
	pf := r.PixelFormat()
	//bytesPerPixel := r.PixelFormat().BPP / 8
	//bytesBuff := &bytes.Buffer{}
	zippedLen, err := ReadUint32(r)
	if err != nil {
		return err
	}

	b, err := ReadBytes(int(zippedLen), r)
	if err != nil {
		return err
	}
	bytesBuff := bytes.NewBuffer(b)

	if enc.unzipper == nil {
		enc.unzipper, err = zlib.NewReader(bytesBuff)
		enc.zippedBuff = bytesBuff
		if err != nil {
			return err
		}
	} else {
		enc.zippedBuff.Write(b)
	}
	DecodeRaw(enc.unzipper, &pf, rect, enc.Image)

	return nil
}
