package vnc2video

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"io"
	"vnc2video/logger"
)

const (
	HextileRaw                 = 1
	HextileBackgroundSpecified = 2
	HextileForegroundSpecified = 4
	HextileAnySubrects         = 8
	HextileSubrectsColoured    = 16
)

type HextileEncoding struct {
	//Colors []Color
	bytes []byte
	Image draw.Image
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

func (z *HextileEncoding) Type() int32 {
	return 5
}
func (z *HextileEncoding) WriteTo(w io.Writer) (n int, err error) {
	return w.Write(z.bytes)
}
func (z *HextileEncoding) Read(r Conn, rect *Rectangle) error {
	//func (z *HextileEncoding) Read(pixelFmt *PixelFormat, rect *Rectangle, r io.Reader) (Encoding, error) {
	//bytesPerPixel := int(r.PixelFormat().BPP) / 8
	pf := r.PixelFormat()
	var bgCol *color.RGBA
	var fgCol *color.RGBA
	var err error
	var dimensions byte
	var subencoding byte

	//r.StartByteCollection()
	// defer func() {
	// 	z.bytes = r.EndByteCollection()
	// }()

	for ty := rect.Y; ty < rect.Y+rect.Height; ty += 16 {
		th := 16
		if rect.Y+rect.Height-ty < 16 {
			th = int(rect.Y) + int(rect.Height) - int(ty)
		}

		for tx := rect.X; tx < rect.X+rect.Width; tx += 16 {
			tw := 16
			if rect.X+rect.Width-tx < 16 {
				tw = int(rect.X) + int(rect.Width) - int(tx)
			}

			//handle Hextile Subrect(tx, ty, tw, th):
			subencoding, err = ReadUint8(r)

			if err != nil {
				logger.Errorf("HextileEncoding.Read: error in hextile reader: %v", err)
				return err
			}

			if (subencoding & HextileRaw) != 0 {
				rawEnc := r.GetEncInstance(EncRaw)
				rawEnc.Read(r, &Rectangle{0, 0, uint16(tw), uint16(th), EncRaw, rawEnc})
				//ReadBytes(tw*th*bytesPerPixel, r)
				continue
			}
			if (subencoding & HextileBackgroundSpecified) != 0 {
				//ReadBytes(int(bytesPerPixel), r)

				bgCol, err = ReadColor(r, &pf)
				rBounds := image.Rectangle{Min: image.Point{int(tx), int(ty)}, Max: image.Point{int(tw), int(th)}}
				FillRect(z.Image, &rBounds, bgCol)
			}
			if (subencoding & HextileForegroundSpecified) != 0 {
				fgCol, err = ReadColor(r, &pf)
			}
			if (subencoding & HextileAnySubrects) == 0 {
				//logger.Debug("hextile reader: no Subrects")
				continue
			}

			nSubrects, err := ReadUint8(r)
			if err != nil {
				return err
			}
			//bufsize := int(nSubrects) * 2
			colorSpecified := ((subencoding & HextileSubrectsColoured) != 0)
			for i := 0; i < int(nSubrects); i++ {
				var color *color.RGBA
				if colorSpecified {
					color, err = ReadColor(r, &pf)
					if err != nil {
						logger.Error("Hextile decoder: problem reading color from connection: ", err)
						return err
					}
				} else {
					color = fgCol
				}
				//int color = colorSpecified ? renderer.readPixelColor(transport) : colors[FG_COLOR_INDEX];
				fgCol = color
				dimensions, err = ReadUint8(r) // bits 7-4 for x, bits 3-0 for y
				if err != nil {
					logger.Error("Hextile decoder: problem reading dimensions from connection: ", err)
					return err
				}
				subtileX := dimensions >> 4 & 0x0f
				subtileY := dimensions & 0x0f
				dimensions, err = ReadUint8(r) // bits 7-4 for w, bits 3-0 for h
				if err != nil {
					logger.Error("Hextile decoder: problem reading 2nd dimensions from connection: ", err)
					return err
				}
				subtileWidth := 1 + (dimensions >> 4 & 0x0f)
				subtileHeight := 1 + (dimensions & 0x0f)
				subrectBounds := image.Rectangle{Min: image.Point{int(tx) + int(subtileX), int(ty) + int(subtileY)}, Max: image.Point{int(subtileWidth), int(subtileHeight)}}
				FillRect(z.Image, &subrectBounds, color)
			}
		}
	}

	return nil
}
