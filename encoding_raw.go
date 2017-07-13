package vnc

type RawEncoding struct {
	Colors []Color
}

func (*RawEncoding) Supported(Conn) bool {
	return true
}

func (enc *RawEncoding) Write(c Conn, rect *Rectangle) error {
	var err error

	for _, clr := range enc.Colors {
		if err = clr.Write(c); err != nil {
			return err
		}
	}

	return err
}

// Read implements the Encoding interface.
func (enc *RawEncoding) Read(c Conn, rect *Rectangle) error {
	pf := c.PixelFormat()
	cm := c.ColorMap()
	colors := make([]Color, rect.Area())

	for y := uint16(0); y < rect.Height; y++ {
		for x := uint16(0); x < rect.Width; x++ {
			color := NewColor(&pf, &cm)
			if err := color.Read(c); err != nil {
				return err
			}
			colors[int(y)*int(rect.Width)+int(x)] = *color
		}
	}

	enc.Colors = colors
	return nil
}

func (*RawEncoding) Type() EncodingType { return EncRaw }
