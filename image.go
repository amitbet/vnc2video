package vnc2video

import (
	"encoding/binary"
	"fmt"
	"image"
	"github.com/amitbet/vnc2video/logger"
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

// ColorMap represent color map
type ColorMap [256]Color

// NewColor returns a new Color object
func NewColor(pf *PixelFormat, cm *ColorMap) *Color {
	return &Color{
		pf: pf,
		cm: cm,
	}
}

// Rectangle represents a rectangle of pixel data
type Rectangle struct {
	X, Y          uint16
	Width, Height uint16
	EncType       EncodingType
	Enc           Encoding
}

// String return string representation
func (rect *Rectangle) String() string {
	return fmt.Sprintf("rect x: %d, y: %d, width: %d, height: %d, enc: %s", rect.X, rect.Y, rect.Width, rect.Height, rect.EncType)
}

// NewRectangle returns new rectangle
func NewRectangle() *Rectangle {
	return &Rectangle{}
}

// Write marshal color to conn
func (clr *Color) Write(c Conn) error {
	var err error
	pf := c.PixelFormat()
	order := pf.order()
	pixel := clr.cmIndex
	if clr.pf.TrueColor != 0 {
		pixel = uint32(clr.R) << pf.RedShift
		pixel |= uint32(clr.G) << pf.GreenShift
		pixel |= uint32(clr.B) << pf.BlueShift
	}

	switch pf.BPP {
	case 8:
		err = binary.Write(c, order, byte(pixel))
	case 16:
		err = binary.Write(c, order, uint16(pixel))
	case 32:
		err = binary.Write(c, order, uint32(pixel))
	}

	return err
}

// Read unmarshal color from conn
func (clr *Color) Read(c Conn) error {
	order := clr.pf.order()
	var pixel uint32

	switch clr.pf.BPP {
	case 8:
		var px uint8
		if err := binary.Read(c, order, &px); err != nil {
			return err
		}
		pixel = uint32(px)
	case 16:
		var px uint16
		if err := binary.Read(c, order, &px); err != nil {
			return err
		}
		pixel = uint32(px)
	case 32:
		var px uint32
		if err := binary.Read(c, order, &px); err != nil {
			return err
		}
		pixel = uint32(px)
	}

	if clr.pf.TrueColor != 0 {
		clr.R = uint16((pixel >> clr.pf.RedShift) & uint32(clr.pf.RedMax))
		clr.G = uint16((pixel >> clr.pf.GreenShift) & uint32(clr.pf.GreenMax))
		clr.B = uint16((pixel >> clr.pf.BlueShift) & uint32(clr.pf.BlueMax))
	} else {
		*clr = clr.cm[pixel]
		clr.cmIndex = pixel
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

// Write marshal rectangle to conn
func (rect *Rectangle) Write(c Conn) error {
	var err error

	if err = binary.Write(c, binary.BigEndian, rect.X); err != nil {
		return err
	}
	if err = binary.Write(c, binary.BigEndian, rect.Y); err != nil {
		return err
	}
	if err = binary.Write(c, binary.BigEndian, rect.Width); err != nil {
		return err
	}
	if err = binary.Write(c, binary.BigEndian, rect.Height); err != nil {
		return err
	}
	if err = binary.Write(c, binary.BigEndian, rect.EncType); err != nil {
		return err
	}

	return rect.Enc.Write(c, rect)
}

// Read unmarshal rectangle from conn
func (rect *Rectangle) Read(c Conn) error {
	var err error

	if err = binary.Read(c, binary.BigEndian, &rect.X); err != nil {
		return err
	}
	if err = binary.Read(c, binary.BigEndian, &rect.Y); err != nil {
		return err
	}
	if err = binary.Read(c, binary.BigEndian, &rect.Width); err != nil {
		return err
	}
	if err = binary.Read(c, binary.BigEndian, &rect.Height); err != nil {
		return err
	}
	if err = binary.Read(c, binary.BigEndian, &rect.EncType); err != nil {
		return err
	}
	logger.Debug(rect)
	switch rect.EncType {
	// case EncCopyRect:
	// 	rect.Enc = &CopyRectEncoding{}
	// case EncTight:
	// 	rect.Enc = c.GetEncInstance(rect.EncType)
	// case EncTightPng:
	// 	rect.Enc = &TightPngEncoding{}
	// case EncRaw:
	// 	if strings.HasPrefix(c.Protocol(), "aten") {
	// 		rect.Enc = &AtenHermon{}
	// 	} else {
	// 		rect.Enc = &RawEncoding{}
	// 	}
	case EncDesktopSizePseudo:
		rect.Enc = &DesktopSizePseudoEncoding{}
	case EncDesktopNamePseudo:
		rect.Enc = &DesktopNamePseudoEncoding{}
	// case EncXCursorPseudo:
	// 	rect.Enc = &XCursorPseudoEncoding{}
	// case EncAtenHermon:
	// 	rect.Enc = &AtenHermon{}
	default:
		rect.Enc = c.GetEncInstance(rect.EncType)
		if rect.Enc == nil {
			return fmt.Errorf("unsupported encoding %s", rect.EncType)
		}
	}

	return rect.Enc.Read(c, rect)
}

// Area returns the total area in pixels of the Rectangle
func (rect *Rectangle) Area() int { return int(rect.Width) * int(rect.Height) }
