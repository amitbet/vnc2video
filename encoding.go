package vnc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"sync"
)

// EncodingType represents a known VNC encoding type.
type EncodingType int32

//go:generate stringer -type=EncodingType

const (
	EncRaw                           EncodingType = 0
	EncCopyRect                      EncodingType = 1
	EncRRE                           EncodingType = 2
	EncCoRRE                         EncodingType = 4
	EncHextile                       EncodingType = 5
	EncZlib                          EncodingType = 6
	EncTight                         EncodingType = 7
	EncZlibHex                       EncodingType = 8
	EncUltra1                        EncodingType = 9
	EncUltra2                        EncodingType = 10
	EncJPEG                          EncodingType = 21
	EncJRLE                          EncodingType = 22
	EncTRLE                          EncodingType = 15
	EncZRLE                          EncodingType = 16
	EncJPEGQualityLevelPseudo10      EncodingType = -23
	EncJPEGQualityLevelPseudo9       EncodingType = -24
	EncJPEGQualityLevelPseudo8       EncodingType = -25
	EncJPEGQualityLevelPseudo7       EncodingType = -26
	EncJPEGQualityLevelPseudo6       EncodingType = -27
	EncJPEGQualityLevelPseudo5       EncodingType = -28
	EncJPEGQualityLevelPseudo4       EncodingType = -29
	EncJPEGQualityLevelPseudo3       EncodingType = -30
	EncJPEGQualityLevelPseudo2       EncodingType = -31
	EncJPEGQualityLevelPseudo1       EncodingType = -32
	EncColorPseudo                   EncodingType = -239
	EncDesktopSizePseudo             EncodingType = -223
	EncLastRectPseudo                EncodingType = -224
	EncCompressionLevel10            EncodingType = -247
	EncCompressionLevel9             EncodingType = -248
	EncCompressionLevel8             EncodingType = -249
	EncCompressionLevel7             EncodingType = -250
	EncCompressionLevel6             EncodingType = -251
	EncCompressionLevel5             EncodingType = -252
	EncCompressionLevel4             EncodingType = -253
	EncCompressionLevel3             EncodingType = -254
	EncCompressionLevel2             EncodingType = -255
	EncCompressionLevel1             EncodingType = -256
	EncQEMUPointerMotionChangePseudo EncodingType = -257
	EncQEMUExtendedKeyEventPseudo    EncodingType = -258
	EncTightPng                      EncodingType = -260
	EncExtendedDesktopSizePseudo     EncodingType = -308
	EncXvpPseudo                     EncodingType = -309
	EncFencePseudo                   EncodingType = -312
	EncContinuousUpdatesPseudo       EncodingType = -313
	EncClientRedirect                EncodingType = -311
)

//go:generate stringer -type=TightCompression

type TightCompression uint8

const (
	TightCompressionBasic TightCompression = 0
	TightCompressionFill  TightCompression = 8
	TightCompressionJPEG  TightCompression = 9
	TightCompressionPNG   TightCompression = 10
)

//go:generate stringer -type=TightFilter

type TightFilter uint8

const (
	TightFilterCopy     TightFilter = 0
	TightFilterPalette  TightFilter = 1
	TightFilterGradient TightFilter = 2
)

var bPool = sync.Pool{
	New: func() interface{} {
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		return new(bytes.Buffer)
	},
}

type Encoding interface {
	Type() EncodingType
	Read(Conn, *Rectangle) error
	Write(Conn, *Rectangle) error
}

type CopyRectEncoding struct {
	SX, SY uint16
}

func (CopyRectEncoding) Type() EncodingType { return EncCopyRect }

func (enc *CopyRectEncoding) Read(c Conn, rect *Rectangle) error {
	if err := binary.Read(c, binary.BigEndian, &enc.SX); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &enc.SY); err != nil {
		return err
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

type RawEncoding struct {
	Colors []Color
}

type TightEncoding struct{}

func (*TightEncoding) Type() EncodingType { return EncTight }

func (enc *TightEncoding) Write(c Conn, rect *Rectangle) error {
	return nil
}

func (enc *TightEncoding) Read(c Conn, rect *Rectangle) error {
	return nil
}

type TightCC struct {
	Compression TightCompression
	Filter      TightFilter
}

func readTightCC(c Conn) (*TightCC, error) {
	var ccb uint8 // compression control byte
	if err := binary.Read(c, binary.BigEndian, &ccb); err != nil {
		return nil, err
	}
	cmp := TightCompression(ccb >> 4)
	switch cmp {
	case TightCompressionBasic:
		return &TightCC{TightCompressionBasic, TightFilterCopy}, nil
	case TightCompressionFill:
		return &TightCC{TightCompressionFill, TightFilterCopy}, nil
	case TightCompressionPNG:
		return &TightCC{TightCompressionPNG, TightFilterCopy}, nil
	}
	return nil, fmt.Errorf("unknown tight compression %d", cmp)
}

func setBit(n uint8, pos uint8) uint8 {
	n |= (1 << pos)
	return n
}

func clrBit(n uint8, pos uint8) uint8 {
	n = n &^ (1 << pos)
	return n
}

func hasBit(n uint8, pos uint8) bool {
	v := n & (1 << pos)
	return (v > 0)
}

func getBit(n uint8, pos uint8) uint8 {
	n = n & (1 << pos)
	return n
}

func writeTightCC(c Conn, tcc *TightCC) error {
	var ccb uint8 // compression control byte
	switch tcc.Compression {
	case TightCompressionFill:
		ccb = setBit(ccb, 7)
	case TightCompressionJPEG:
		ccb = setBit(ccb, 7)
		ccb = setBit(ccb, 4)
	case TightCompressionPNG:
		ccb = setBit(ccb, 7)
		ccb = setBit(ccb, 5)
	}
	return binary.Write(c, binary.BigEndian, ccb)
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

type TightPixel struct {
	R uint8
	G uint8
	B uint8
}

type TightPngEncoding struct {
	TightCC *TightCC
	Image   image.Image
}

func (*TightPngEncoding) Type() EncodingType { return EncTightPng }

func (enc *TightPngEncoding) Read(c Conn, rect *Rectangle) error {
	tcc, err := readTightCC(c)
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
		enc.Image, err = png.Decode(io.LimitReader(c, int64(l)))
		if err != nil {
			return err
		}
	case TightCompressionFill:
		var tpx TightPixel
		if err := binary.Read(c, binary.BigEndian, &tpx); err != nil {
			return err
		}
		enc.Image = image.NewRGBA(image.Rect(0, 0, 1, 1))
		enc.Image.(draw.Image).Set(0, 0, color.RGBA{R: tpx.R, G: tpx.G, B: tpx.B, A: 1})
	default:
		return fmt.Errorf("unknown compression %d", cmp)
	}
	return nil
}

func writeTightLength(c Conn, l int) error {
	var buf []uint8

	buf = append(buf, uint8(l&0x7F))
	if l > 0x7F {
		buf[0] |= 0x80
		buf = append(buf, uint8((l>>7)&0x7F))
		if l > 0x3FFF {
			buf[1] |= 0x80
			buf = append(buf, uint8((l>>14)&0xFF))
		}
	}
	return binary.Write(c, binary.BigEndian, buf)
}

func readTightLength(c Conn) (int, error) {
	var length int
	var err error
	var b uint8

	if err = binary.Read(c, binary.BigEndian, &b); err != nil {
		return 0, err
	}

	length = int(b) & 0x7F
	if (b & 0x80) == 0 {
		return length, nil
	}

	if err = binary.Read(c, binary.BigEndian, &b); err != nil {
		return 0, err
	}
	length |= (int(b) & 0x7F) << 7
	if (b & 0x80) == 0 {
		return length, nil
	}

	if err = binary.Read(c, binary.BigEndian, &b); err != nil {
		return 0, err
	}
	length |= (int(b) & 0xFF) << 14

	return length, nil
}

func (enc *RawEncoding) Write(c Conn, rect *Rectangle) error {
	var err error
	for _, clr := range enc.Colors {
		if err = clr.Write(c); err != nil {
			break
		}
	}
	return err
}

// Read implements the Encoding interface.
func (enc *RawEncoding) Read(c Conn, rect *Rectangle) error {
	var err error
	pf := c.PixelFormat()
	cm := c.ColorMap()
	colors := make([]Color, rect.Area())
Loop:
	for y := uint16(0); y < rect.Height; y++ {
		for x := uint16(0); x < rect.Width; x++ {
			color := NewColor(pf, cm)
			if err = color.Read(c); err != nil {
				break Loop
			}
			colors[int(y)*int(rect.Width)+int(x)] = *color
		}
	}

	enc.Colors = colors
	return err
}

func (*RawEncoding) Type() EncodingType { return EncRaw }

// DesktopSizePseudoEncoding represents a desktop size message from the server.
type DesktopSizePseudoEncoding struct {
}

func (*DesktopSizePseudoEncoding) Type() EncodingType { return EncDesktopSizePseudo }

// Read implements the Encoding interface.
func (*DesktopSizePseudoEncoding) Read(c Conn, rect *Rectangle) error {
	c.SetWidth(rect.Width)
	c.SetHeight(rect.Height)
	return nil
}

func (enc *DesktopSizePseudoEncoding) Write(c Conn, rect *Rectangle) error {
	return nil
}
