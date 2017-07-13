package vnc

import (
	"encoding/binary"
	"fmt"
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

type TightEncoding struct{}

func (*TightEncoding) Supported(Conn) bool {
	return true
}

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

type TightPixel struct {
	R uint8
	G uint8
	B uint8
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
