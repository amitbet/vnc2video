package vnc2video

import "encoding/binary"

// DesktopNamePseudoEncoding represents a desktop size message from the server.
type DesktopNamePseudoEncoding struct {
	Name []byte
}

func (*DesktopNamePseudoEncoding) Supported(Conn) bool {
	return true
}
func (*DesktopNamePseudoEncoding) Reset() error {
	return nil
}
func (*DesktopNamePseudoEncoding) Type() EncodingType { return EncDesktopNamePseudo }

// Read implements the Encoding interface.
func (enc *DesktopNamePseudoEncoding) Read(c Conn, rect *Rectangle) error {
	var length uint32
	if err := binary.Read(c, binary.BigEndian, &length); err != nil {
		return err
	}
	name := make([]byte, length)
	if err := binary.Read(c, binary.BigEndian, &name); err != nil {
		return err
	}
	enc.Name = name
	return nil
}

func (enc *DesktopNamePseudoEncoding) Write(c Conn, rect *Rectangle) error {
	if err := binary.Write(c, binary.BigEndian, uint32(len(enc.Name))); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, enc.Name); err != nil {
		return err
	}

	return c.Flush()
}
