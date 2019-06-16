package vnc2video

import (
	"image"
	"image/color"
	"image/draw"
	"io"
	"github.com/amitbet/vnc2video/logger"
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

func (enc *HextileEncoding) SetTargetImage(img draw.Image) {
	enc.Image = img
}

func (*HextileEncoding) Supported(Conn) bool {
	return true
}

func (enc *HextileEncoding) Reset() error {
	//enc.decoders = make([]io.Reader, 4)
	//enc.decoderBuffs = make([]*bytes.Buffer, 4)
	return nil
}

func (z *HextileEncoding) Type() EncodingType {
	return EncHextile
}

func (z *HextileEncoding) WriteTo(w io.Writer) (n int, err error) {
	return w.Write(z.bytes)
}

func (enc *HextileEncoding) Write(c Conn, rect *Rectangle) error {
	return nil
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
	logger.Tracef("HextileEncoding.Read: got hextile rect: %v", rect)
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
				rawEnc.Read(r, &Rectangle{X: uint16(tx), Y: uint16(ty), Width: uint16(tw), Height: uint16(th), EncType: EncRaw, Enc: rawEnc})
				//ReadBytes(tw*th*int(pf.BPP)/8, r)
				continue
			}
			if (subencoding & HextileBackgroundSpecified) != 0 {
				//ReadBytes(int(bytesPerPixel), r)

				bgCol, err = ReadColor(r, &pf)
				if err != nil {
					logger.Errorf("HextileEncoding.Read: error in hextile bg color reader: %v", err)
					return err
				}

				//logger.Tracef("%v %v", rBounds, bgCol)
			}
			rBounds := image.Rectangle{Min: image.Point{int(tx), int(ty)}, Max: image.Point{int(tx) + int(tw), int(ty) + int(th)}}
			//logger.Tracef("filling background rect: %v, col: %v", rBounds, bgCol)
			FillRect(z.Image, &rBounds, bgCol)

			if (subencoding & HextileForegroundSpecified) != 0 {
				fgCol, err = ReadColor(r, &pf)
				if err != nil {
					logger.Errorf("HextileEncoding.Read: error in hextile fg color reader: %v", err)
					return err
				}
			}
			if (subencoding & HextileAnySubrects) == 0 {
				//logger.Trace("hextile reader: no Subrects")
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
						logger.Error("HextileEncoding.Read: problem reading color from connection: ", err)
						return err
					}
				} else {
					color = fgCol
				}
				//int color = colorSpecified ? renderer.readPixelColor(transport) : colors[FG_COLOR_INDEX];
				fgCol = color
				dimensions, err = ReadUint8(r) // bits 7-4 for x, bits 3-0 for y
				if err != nil {
					logger.Error("HextileEncoding.Read: problem reading dimensions from connection: ", err)
					return err
				}
				subtileX := dimensions >> 4 & 0x0f
				subtileY := dimensions & 0x0f
				dimensions, err = ReadUint8(r) // bits 7-4 for w, bits 3-0 for h
				if err != nil {
					logger.Error("HextileEncoding.Read: problem reading 2nd dimensions from connection: ", err)
					return err
				}
				subtileWidth := 1 + (dimensions >> 4 & 0x0f)
				subtileHeight := 1 + (dimensions & 0x0f)
				subrectBounds := image.Rectangle{Min: image.Point{int(tx) + int(subtileX), int(ty) + int(subtileY)}, Max: image.Point{int(tx) + int(subtileX) + int(subtileWidth), int(ty) + int(subtileY) + int(subtileHeight)}}
				FillRect(z.Image, &subrectBounds, color)
				//logger.Tracef("%v", subrectBounds)
			}
		}
	}

	return nil
}
