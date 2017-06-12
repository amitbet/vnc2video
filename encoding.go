package vnc

// EncodingType represents a known VNC encoding type.
type EncodingType int32

//go:generate stringer -type=EncodingType

const (
	EncRaw               EncodingType = 0
	EncCopyRect          EncodingType = 1
	EncRRE               EncodingType = 2
	EncHextile           EncodingType = 5
	EncTRLE              EncodingType = 15
	EncZRLE              EncodingType = 16
	EncColorPseudo       EncodingType = -239
	EncDesktopSizePseudo EncodingType = -223
)

type Encoding interface {
	Type() EncodingType
	Read(Conn, *Rectangle) error
	Write(Conn, *Rectangle) error
}

type RawEncoding struct {
	Colors []Color
}

func (enc *RawEncoding) Write(c Conn, rect *Rectangle) error {
	/*
		for _, cl := range enc.Colors {
			bytes, err := cl.Marshal()
			if err != nil {
				return err
			}
			if err := binary.Write(c, binary.BigEndian, bytes); err != nil {
				return err
			}
		}
	*/
	return nil

}

// Read implements the Encoding interface.
func (enc *RawEncoding) Read(c Conn, rect *Rectangle) error {
	/*
		var buf bytes.Buffer
		pf := c.PixelFormat()
		cm := c.ColorMap()
		bytesPerPixel := int(pf.BPP / 8)
		n := rect.Area() * bytesPerPixel
		if err := c.receiveN(&buf, n); err != nil {
			return fmt.Errorf("unable to read rectangle with raw encoding: %s", err)
		}

		colors := make([]Color, rect.Area())
		for y := uint16(0); y < rect.Height; y++ {
			for x := uint16(0); x < rect.Width; x++ {
				color := NewColor(pf, cm)
				if err := color.Unmarshal(buf.Next(bytesPerPixel)); err != nil {
					return nil, err
				}
				colors[int(y)*int(rect.Width)+int(x)] = *color
			}
		}

		return &RawEncoding{colors}, nil
	*/
	return nil
}

func (*RawEncoding) Type() EncodingType { return EncRaw }

// DesktopSizePseudoEncoding represents a desktop size message from the server.
type DesktopSizePseudoEncoding struct{}

// Read implements the Encoding interface.
func (*DesktopSizePseudoEncoding) Read(c Conn, rect *Rectangle) error {
	c.SetWidth(rect.Width)
	c.SetHeight(rect.Height)

	return nil
}

func (enc *DesktopSizePseudoEncoding) Write(c *ServerConn, rect *Rectangle) error {
	return nil
}
