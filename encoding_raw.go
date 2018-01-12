package vnc2video

import "image"
import "image/draw"
import "image/color"

type RawEncoding struct {
	Image image.Image
	//Colors []Color
}

func (*RawEncoding) Supported(Conn) bool {
	return true
}

func (*RawEncoding) Reset() error {
	return nil
}

func (enc *RawEncoding) Write(c Conn, rect *Rectangle) error {
	var err error
	//
	//for _, clr := range enc.Colors {
	//	if err = clr.Write(c); err != nil {
	//		return err
	//	}
	//}

	return err
}
func (enc *RawEncoding) SetTargetImage(img draw.Image) {
	enc.Image = img
}

// Read implements the Encoding interface.
func (enc *RawEncoding) Read(c Conn, rect *Rectangle) error {
	pf := c.PixelFormat()
	cm := c.ColorMap()
	//colors := make([]Color, rect.Area())

	for y := 0; y < int(rect.Height); y++ {
		for x := 0; x < int(rect.Width); x++ {
			c1 := NewColor(&pf, &cm)
			if err := c1.Read(c); err != nil {
				return err
			}

			c2:=color.RGBA{R:uint8(c1.R),G:uint8(c1.G),B:uint8(c1.B),A:1}
			//c3 := color.RGBAModel.Convert(c2)

			enc.Image.(draw.Image).Set(int(rect.X)+x,int(rect.Y)+y,c2)
			//colors[int(y)*int(rect.Width)+int(x)] = *color
		}
	}

	//enc.Colors = colors
	return nil
}

func (*RawEncoding) Type() EncodingType { return EncRaw }
