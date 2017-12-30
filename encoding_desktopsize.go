package vnc2webm

// DesktopSizePseudoEncoding represents a desktop size message from the server.
type DesktopSizePseudoEncoding struct{}

func (*DesktopSizePseudoEncoding) Supported(Conn) bool {
	return true
}

func (*DesktopSizePseudoEncoding) Type() EncodingType { return EncDesktopSizePseudo }

// Read implements the Encoding interface.
func (*DesktopSizePseudoEncoding) Read(c Conn, rect *Rectangle) error {
	return nil
}

func (enc *DesktopSizePseudoEncoding) Write(c Conn, rect *Rectangle) error {
	return nil
}
