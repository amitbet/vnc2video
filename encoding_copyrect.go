package vnc

import "encoding/binary"

type CopyRectEncoding struct {
	SX, SY uint16
}

func (*CopyRectEncoding) Supported(Conn) bool {
	return true
}

func (*CopyRectEncoding) Type() EncodingType { return EncCopyRect }

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
