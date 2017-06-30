package vnc

import (
	"encoding/binary"
	"fmt"
)

type AtenHermon struct {
	_             [4]byte
	AtenLength    uint32
	AtenType      uint8
	_             [1]byte
	AtenSubrects  uint32
	AtenRawLength uint32
	Encodings     []Encoding
}

func (*AtenHermon) Type() EncodingType { return EncAtenHermon }

func (enc *AtenHermon) Read(c Conn, rect *Rectangle) error {
	var pad4 [4]byte

	if err := binary.Read(c, binary.BigEndian, &pad4); err != nil {
		return err
	}

	var aten_length uint32
	if err := binary.Read(c, binary.BigEndian, &aten_length); err != nil {
		return err
	}
	enc.AtenLength = aten_length

	if rect.Width == 64896 && rect.Height == 65056 {
		if aten_length != 10 && aten_length != 0 {
			return fmt.Errorf("screen is off and length is invalid")
		}
		aten_length = 0
	}

	if c.Width() != rect.Width && c.Height() != rect.Height {
		c.SetWidth(rect.Width)
		c.SetHeight(rect.Height)
	}

	var aten_type uint8
	if err := binary.Read(c, binary.BigEndian, &aten_type); err != nil {
		return err
	}
	enc.AtenType = aten_type

	var pad1 [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad1); err != nil {
		return err
	}

	var subrects uint32
	if err := binary.Read(c, binary.BigEndian, &subrects); err != nil {
		return err
	}
	enc.AtenSubrects = subrects

	var raw_length uint32
	if err := binary.Read(c, binary.BigEndian, &raw_length); err != nil {
		return err
	}
	enc.AtenRawLength = raw_length

	if aten_length != raw_length {
		return fmt.Errorf("aten_length != raw_length, %d != %d", aten_length, raw_length)
	}

	aten_length -= 10 // skip

	for aten_length > 0 {
		switch aten_type {
		case 0: //subrects
			panic("unknown subrect")
			var a uint16
			var b uint16
			var x uint8
			var y uint8
			if err := binary.Read(c, binary.BigEndian, &a); err != nil {
				return err
			}
			if err := binary.Read(c, binary.BigEndian, &b); err != nil {
				return err
			}
			if err := binary.Read(c, binary.BigEndian, &y); err != nil {
				return err
			}
			if err := binary.Read(c, binary.BigEndian, &x); err != nil {
				return err
			}
			data := make([]byte, 16*16*uint32(c.PixelFormat().BPP))
			if err := binary.Read(c, binary.BigEndian, &data); err != nil {
				return err
			}

			aten_length -= 6 + (16 * 16 * uint32(c.PixelFormat().BPP))
		case 1: //raw
			fmt.Printf("raw reader %d %s\n", aten_length, rect)
			encRaw := &RawEncoding{}
			if err := encRaw.Read(c, rect); err != nil {
				return err
			}
			enc.Encodings = append(enc.Encodings, encRaw)
			aten_length = 0
			//aten_length -= uint32(rect.Area()) * uint32(c.PixelFormat().BPP)
		default:
			return fmt.Errorf("unknown aten hermon type %d", aten_type)

		}
	}

	if aten_length < 0 {
		return fmt.Errorf("aten_len dropped below zero")
	}
	fmt.Printf("aten hermon readed\n")
	return nil
}

func (enc *AtenHermon) Write(c Conn, rect *Rectangle) error {
	var pad4 [4]byte

	if err := binary.Write(c, binary.BigEndian, pad4); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, enc.AtenLength); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, enc.AtenType); err != nil {
		return err
	}

	var pad1 [1]byte
	if err := binary.Write(c, binary.BigEndian, pad1); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, enc.AtenSubrects); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, enc.AtenRawLength); err != nil {
		return err
	}

	for _, ew := range enc.Encodings {
		if err := ew.Write(c, rect); err != nil {
			return err
		}
	}

	fmt.Printf("aten hermon writed\n")
	return nil
}
