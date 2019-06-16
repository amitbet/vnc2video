package vnc2video

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"github.com/amitbet/vnc2video/logger"
)

func (*TightPngEncoding) Supported(Conn) bool {
	return true
}
func (*TightPngEncoding) Reset() error {
	return nil
}

func (enc *TightPngEncoding) Write(c Conn, rect *Rectangle) error {
	if err := writeTightCC(c, enc.TightCC); err != nil {
		return err
	}
	cmp := enc.TightCC.Compression
	switch cmp {
	case TightCompressionPNG:
		buf := bPool.Get().(*bytes.Buffer)
		buf.Reset()
		defer bPool.Put(buf)
		pngEnc := &png.Encoder{CompressionLevel: png.BestSpeed}
		//pngEnc := &png.Encoder{CompressionLevel: png.NoCompression}
		if err := pngEnc.Encode(buf, enc.Image); err != nil {
			return err
		}
		if err := writeTightLength(c, buf.Len()); err != nil {
			return err
		}

		if _, err := buf.WriteTo(c); err != nil {
			return err
		}
	case TightCompressionFill:
		var tpx TightPixel
		r, g, b, _ := enc.Image.At(0, 0).RGBA()
		tpx.R = uint8(r)
		tpx.G = uint8(g)
		tpx.B = uint8(b)
		if err := binary.Write(c, binary.BigEndian, tpx); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown tight compression %d", cmp)
	}
	return nil
}

type TightPngEncoding struct {
	TightCC *TightCC
	Image   draw.Image
}

func (*TightPngEncoding) Type() EncodingType { return EncTightPng }

func (enc *TightPngEncoding) Read(c Conn, rect *Rectangle) error {
	tcc, err := readTightCC(c)
	logger.Trace("starting to read a tight rect: %v", rect)
	if err != nil {
		return err
	}
	enc.TightCC = tcc
	cmp := enc.TightCC.Compression
	switch cmp {
	case TightCompressionPNG:
		l, err := readTightLength(c)
		if err != nil {
			return err
		}
		img, err := png.Decode(io.LimitReader(c, int64(l)))
		if err != nil {
			return err
		}
		//draw.Draw(enc.Image, enc.Image.Bounds(), img, image.Point{X: int(rect.X), Y: int(rect.Y)}, draw.Src)
		DrawImage(enc.Image, img, image.Point{X: int(rect.X), Y: int(rect.Y)})
	case TightCompressionFill:
		var tpx TightPixel
		if err := binary.Read(c, binary.BigEndian, &tpx); err != nil {
			return err
		}
		//enc.Image = image.NewRGBA(image.Rect(0, 0, 1, 1))
		col := color.RGBA{R: tpx.R, G: tpx.G, B: tpx.B, A: 1}
		myRect := MakeRectFromVncRect(rect)
		FillRect(enc.Image, &myRect, col)
		//enc.Image.(draw.Image).Set(0, 0, color.RGBA{R: tpx.R, G: tpx.G, B: tpx.B, A: 1})
	default:
		return fmt.Errorf("unknown compression %d", cmp)
	}
	return nil
}

func (enc *TightPngEncoding) SetTargetImage(img draw.Image) {
	enc.Image = img
}
