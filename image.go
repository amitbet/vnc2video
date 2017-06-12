package vnc

import (
	"image"
)

//var _ draw.Drawer = (*ServerConn)(nil)
//var _ draw.Image = (*ServerConn)(nil)

// Color represents a single color in a color map.
type Color struct {
	pf      *PixelFormat
	cm      *ColorMap
	cmIndex uint32 // Only valid if pf.TrueColor is false.
	R, G, B uint16
}

type ColorMap [256]Color

// NewColor returns a new Color object.
func NewColor(pf *PixelFormat, cm *ColorMap) *Color {
	return &Color{
		pf: pf,
		cm: cm,
	}
}

// Rectangle represents a rectangle of pixel data.
type Rectangle struct {
	X, Y          uint16
	Width, Height uint16
	Enc           Encoding
}

// Marshal implements the Marshaler interface.
func (c *Color) Marshal() ([]byte, error) {
	order := c.pf.order()
	pixel := c.cmIndex
	if c.pf.TrueColor == 1 {
		pixel = uint32(c.R) << c.pf.RedShift
		pixel |= uint32(c.G) << c.pf.GreenShift
		pixel |= uint32(c.B) << c.pf.BlueShift
	}

	var bytes []byte
	switch c.pf.BPP {
	case 8:
		bytes = make([]byte, 1)
		bytes[0] = byte(pixel)
	case 16:
		bytes = make([]byte, 2)
		order.PutUint16(bytes, uint16(pixel))
	case 32:
		bytes = make([]byte, 4)
		order.PutUint32(bytes, pixel)
	}

	return bytes, nil
}

// Unmarshal implements the Unmarshaler interface.
func (c *Color) Unmarshal(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	order := c.pf.order()

	var pixel uint32
	switch c.pf.BPP {
	case 8:
		pixel = uint32(data[0])
	case 16:
		pixel = uint32(order.Uint16(data))
	case 32:
		pixel = order.Uint32(data)
	}

	if c.pf.TrueColor == 1 {
		c.R = uint16((pixel >> c.pf.RedShift) & uint32(c.pf.RedMax))
		c.G = uint16((pixel >> c.pf.GreenShift) & uint32(c.pf.GreenMax))
		c.B = uint16((pixel >> c.pf.BlueShift) & uint32(c.pf.BlueMax))
	} else {
		*c = c.cm[pixel]
		c.cmIndex = pixel
	}

	return nil
}

func colorsToImage(x, y, width, height uint16, colors []Color) *image.RGBA64 {
	rect := image.Rect(int(x), int(y), int(x+width), int(y+height))
	rgba := image.NewRGBA64(rect)
	a := uint16(1)
	for i, color := range colors {
		rgba.Pix[4*i+0] = uint8(color.R >> 8)
		rgba.Pix[4*i+1] = uint8(color.R)
		rgba.Pix[4*i+2] = uint8(color.G >> 8)
		rgba.Pix[4*i+3] = uint8(color.G)
		rgba.Pix[4*i+4] = uint8(color.B >> 8)
		rgba.Pix[4*i+5] = uint8(color.B)
		rgba.Pix[4*i+6] = uint8(a >> 8)
		rgba.Pix[4*i+7] = uint8(a)
	}
	return rgba
}

// Marshal implements the Marshaler interface.
func (r *Rectangle) Marshal() ([]byte, error) {
	/*
		buf := bytes.NewBuffer(nil)

		var msg Rectangle
		msg.X, msg.Y, msg.W, msg.H = r.X, r.Y, r.Width, r.Height
		msg.E = r.Enc.Type()
		if err := binary.Write(buf, binary.BigEndian, msg); err != nil {
			return nil, err
		}

		bytes, err := r.Enc.Marshal()
		if err != nil {
			return nil, err
		}
		if err := binary.Write(buf, binary.BigEndian, bytes); err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	*/
	return nil, nil
}

// Unmarshal implements the Unmarshaler interface.
func (r *Rectangle) Unmarshal(data []byte) error {
	/*
		buf := bytes.NewBuffer(data)

		var msg Rectangle
		if err := binary.Read(buf, binary.BigEndian, &msg); err != nil {
			return err
		}
		r.X, r.Y, r.Width, r.Height = msg.X, msg.Y, msg.W, msg.H

		switch msg.E {
		case encodings.Raw:
			r.Enc = &RawEncoding{}
		default:
			return fmt.Errorf("unable to unmarshal encoding %v", msg.E)
		}
		return nil
	*/
	return nil
}

// Area returns the total area in pixels of the Rectangle.
func (r *Rectangle) Area() int { return int(r.Width) * int(r.Height) }
