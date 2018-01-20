package vnc2video

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"io"
)

func FillRect(img draw.Image, rect *image.Rectangle, c color.Color) {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			img.Set(x, y, c)
		}
	}
}

// Read unmarshal color from conn
func ReadColor(c io.Reader, pf *PixelFormat) (*color.RGBA, error) {
	if pf.TrueColor == 0 {
		return nil, errors.New("support for non true color formats was not implemented")
	}
	order := pf.order()
	var pixel uint32

	switch pf.BPP {
	case 8:
		var px uint8
		if err := binary.Read(c, order, &px); err != nil {
			return nil, err
		}
		pixel = uint32(px)
	case 16:
		var px uint16
		if err := binary.Read(c, order, &px); err != nil {
			return nil, err
		}
		pixel = uint32(px)
	case 32:
		var px uint32
		if err := binary.Read(c, order, &px); err != nil {
			return nil, err
		}
		pixel = uint32(px)
	}

	rgb := color.RGBA{
		R: uint8((pixel >> pf.RedShift) & uint32(pf.RedMax)),
		G: uint8((pixel >> pf.GreenShift) & uint32(pf.GreenMax)),
		B: uint8((pixel >> pf.BlueShift) & uint32(pf.BlueMax)),
		A: 1,
	}

	return &rgb, nil
}

func DecodeRaw(reader io.Reader, pf *PixelFormat, rect *Rectangle, targetImage draw.Image) error {
	for y := 0; y < int(rect.Height); y++ {
		for x := 0; x < int(rect.Width); x++ {
			col, err := ReadColor(reader, pf)
			if err != nil {
				return err
			}

			targetImage.(draw.Image).Set(int(rect.X)+x, int(rect.Y)+y, col)
		}
	}

	return nil
}
