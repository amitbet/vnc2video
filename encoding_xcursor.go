package vnc2video

import (
	"encoding/binary"
	"math"
)

type XCursorPseudoEncoding struct {
	PrimaryR   uint8
	PrimaryG   uint8
	PrimaryB   uint8
	SecondaryR uint8
	SecondaryG uint8
	SecondaryB uint8
	Bitmap     []byte
	Bitmask    []byte
}

func (*XCursorPseudoEncoding) Supported(Conn) bool {
	return true
}
func (*XCursorPseudoEncoding) Reset() error {
	return nil
}

func (*XCursorPseudoEncoding) Type() EncodingType { return EncXCursorPseudo }

// Read implements the Encoding interface.
func (enc *XCursorPseudoEncoding) Read(c Conn, rect *Rectangle) error {
	if err := binary.Read(c, binary.BigEndian, &enc.PrimaryR); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &enc.PrimaryG); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &enc.PrimaryB); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &enc.SecondaryR); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &enc.SecondaryG); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &enc.SecondaryB); err != nil {
		return err
	}

	bitmapsize := int(math.Floor((float64(rect.Width)+7)/8) * float64(rect.Height))
	bitmasksize := int(math.Floor((float64(rect.Width)+7)/8) * float64(rect.Height))

	enc.Bitmap = make([]byte, bitmapsize)
	enc.Bitmask = make([]byte, bitmasksize)
	if err := binary.Read(c, binary.BigEndian, &enc.Bitmap); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &enc.Bitmask); err != nil {
		return err
	}

	return nil
}

func (enc *XCursorPseudoEncoding) Write(c Conn, rect *Rectangle) error {
	if err := binary.Write(c, binary.BigEndian, enc.PrimaryR); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, enc.PrimaryG); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, enc.PrimaryB); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, enc.SecondaryR); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, enc.SecondaryG); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, enc.SecondaryB); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, enc.Bitmap); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, enc.Bitmask); err != nil {
		return err
	}

	return nil
}
