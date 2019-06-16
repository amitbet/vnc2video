package vnc2video

import (
	"bytes"
	"compress/zlib"
	"errors"
	"image/color"
	"image/draw"
	"io"
	"github.com/amitbet/vnc2video/logger"
)

type ZRLEEncoding struct {
	bytes      []byte
	Image      draw.Image
	unzipper   io.Reader
	zippedBuff *bytes.Buffer
}

func (*ZRLEEncoding) Supported(Conn) bool {
	return true
}

func (enc *ZRLEEncoding) SetTargetImage(img draw.Image) {
	enc.Image = img
}

func (enc *ZRLEEncoding) Reset() error {
	enc.unzipper = nil
	return nil
}

func (*ZRLEEncoding) Type() EncodingType { return EncZRLE }

func (z *ZRLEEncoding) WriteTo(w io.Writer) (n int, err error) {
	return w.Write(z.bytes)
}

func (enc *ZRLEEncoding) Write(c Conn, rect *Rectangle) error {
	return nil
}

func IsCPixelSpecific(pf *PixelFormat) bool {
	significant := int(pf.RedMax<<pf.RedShift | pf.GreenMax<<pf.GreenShift | pf.BlueMax<<pf.BlueShift)

	if pf.Depth <= 24 && 32 == pf.BPP && ((significant&0x00ff000000) == 0 || (significant&0x000000ff) == 0) {
		return true
	}
	return false
}

func CalcBytesPerCPixel(pf *PixelFormat) int {
	if IsCPixelSpecific(pf) {
		return 3
	}
	return int(pf.BPP / 8)
}

func (enc *ZRLEEncoding) Read(r Conn, rect *Rectangle) error {
	logger.Tracef("reading ZRLE:%v\n", rect)
	len, err := ReadUint32(r)
	if err != nil {
		return err
	}

	b, err := ReadBytes(int(len), r)
	if err != nil {
		return err
	}

	bytesBuff := bytes.NewBuffer(b)

	if enc.unzipper == nil {
		enc.unzipper, err = zlib.NewReader(bytesBuff)
		enc.zippedBuff = bytesBuff
		if err != nil {
			return err
		}
	} else {
		enc.zippedBuff.Write(b)
	}
	pf := r.PixelFormat()
	enc.renderZRLE(rect, &pf)

	return nil
}

func (enc *ZRLEEncoding) readZRLERaw(reader io.Reader, pf *PixelFormat, tx, ty, tw, th int) error {
	for y := 0; y < int(th); y++ {
		for x := 0; x < int(tw); x++ {
			col, err := readCPixel(reader, pf)
			if err != nil {
				return err
			}

			enc.Image.Set(tx+x, ty+y, col)
		}
	}

	return nil
}

func (enc *ZRLEEncoding) renderZRLE(rect *Rectangle, pf *PixelFormat) error {
	logger.Trace("-----renderZRLE: rendering rect:", rect)
	for tileOffsetY := 0; tileOffsetY < int(rect.Height); tileOffsetY += 64 {

		tileHeight := Min(64, int(rect.Height)-tileOffsetY)

		for tileOffsetX := 0; tileOffsetX < int(rect.Width); tileOffsetX += 64 {

			tileWidth := Min(64, int(rect.Width)-tileOffsetX)
			// read subencoding
			subEnc, err := ReadUint8(enc.unzipper)
			logger.Tracef("-----renderZRLE: rendering got tile:(%d,%d) w:%d, h:%d subEnc:%d", tileOffsetX, tileOffsetY, tileWidth, tileHeight, subEnc)
			if err != nil {
				logger.Errorf("renderZRLE: error while reading subencoding: %v", err)
				return err
			}

			switch {

			case subEnc == 0:
				// Raw subencoding: read cpixels and paint
				err = enc.readZRLERaw(enc.unzipper, pf, int(rect.X)+tileOffsetX, int(rect.Y)+tileOffsetY, tileWidth, tileHeight)
				if err != nil {
					logger.Errorf("renderZRLE: error while reading Raw tile: %v", err)
					return err
				}
			case subEnc == 1:
				// background color tile - just fill
				color, err := readCPixel(enc.unzipper, pf)
				if err != nil {
					logger.Errorf("renderZRLE: error while reading CPixel for bgColor tile: %v", err)
					return err
				}
				myRect := MakeRect(int(rect.X)+tileOffsetX, int(rect.Y)+tileOffsetY, tileWidth, tileHeight)
				FillRect(enc.Image, &myRect, color)
			case subEnc >= 2 && subEnc <= 16:
				err = enc.handlePaletteTile(tileOffsetX, tileOffsetY, tileWidth, tileHeight, subEnc, pf, rect)
				if err != nil {
					return err
				}
			case subEnc == 128:
				err = enc.handlePlainRLETile(tileOffsetX, tileOffsetY, tileWidth, tileHeight, pf, rect)
				if err != nil {
					return err
				}
			case subEnc >= 130 && subEnc <= 255:
				err = enc.handlePaletteRLETile(tileOffsetX, tileOffsetY, tileWidth, tileHeight, subEnc, pf, rect)
				if err != nil {
					return err
				}
			default:
				logger.Errorf("Unknown ZRLE subencoding: %v", subEnc)
			}
		}
	}
	return nil
}

func (enc *ZRLEEncoding) handlePaletteRLETile(tileOffsetX, tileOffsetY, tileWidth, tileHeight int, subEnc uint8, pf *PixelFormat, rect *Rectangle) error {
	// Palette RLE
	paletteSize := subEnc - 128
	palette := make([]*color.RGBA, paletteSize)
	var err error

	// Read RLE palette
	for j := 0; j < int(paletteSize); j++ {
		palette[j], err = readCPixel(enc.unzipper, pf)
		if err != nil {
			logger.Errorf("renderZRLE: error while reading color in palette RLE subencoding: %v", err)
			return err
		}
	}
	var index uint8
	runLen := 0
	for y := 0; y < tileHeight; y++ {
		for x := 0; x < tileWidth; x++ {

			if runLen == 0 {

				// Read length and index
				index, err = ReadUint8(enc.unzipper)
				if err != nil {
					logger.Errorf("renderZRLE: error while reading length and index in palette RLE subencoding: %v", err)
					//return err
				}
				runLen = 1

				// Run is represented by index | 0x80
				// Otherwise, single pixel
				if (index & 0x80) != 0 {

					index -= 128

					runLen, err = readRunLength(enc.unzipper)
					if err != nil {
						logger.Errorf("handlePlainRLETile: error while reading runlength in plain RLE subencoding: %v", err)
						return err
					}

				}
				//logger.Tracef("renderZRLE: writing pixel: col=%v times=%d", palette[index], runLen)
			}

			// Write pixel to image
			enc.Image.Set(tileOffsetX+int(rect.X)+x, tileOffsetY+int(rect.Y)+y, palette[index])
			runLen--
		}
	}
	return nil
}

func (enc *ZRLEEncoding) handlePaletteTile(tileOffsetX, tileOffsetY, tileWidth, tileHeight int, subEnc uint8, pf *PixelFormat, rect *Rectangle) error {
	//subenc here is also palette size
	paletteSize := subEnc
	palette := make([]*color.RGBA, paletteSize)
	var err error
	// Read palette
	for j := 0; j < int(paletteSize); j++ {
		palette[j], err = readCPixel(enc.unzipper, pf)
		if err != nil {
			logger.Errorf("renderZRLE: error while reading CPixel for palette tile: %v", err)
			return err
		}
	}
	// Calculate index size
	var indexBits, mask uint32
	if paletteSize == 2 {
		indexBits = 1
		mask = 0x80
	} else if paletteSize <= 4 {
		indexBits = 2
		mask = 0xC0
	} else {
		indexBits = 4
		mask = 0xF0
	}
	for y := 0; y < tileHeight; y++ {

		// Packing only occurs per-row
		bitsAvailable := uint32(0)
		buffer := uint32(0)

		for x := 0; x < tileWidth; x++ {

			// Buffer more bits if necessary
			if bitsAvailable == 0 {
				bits, err := ReadUint8(enc.unzipper)
				if err != nil {
					logger.Errorf("renderZRLE: error while reading first uint8 into buffer: %v", err)
					return err
				}
				buffer = uint32(bits)
				bitsAvailable = 8
			}

			// Read next pixel
			index := (buffer & mask) >> (8 - indexBits)
			buffer <<= indexBits
			bitsAvailable -= indexBits

			// Write pixel to image
			enc.Image.Set(tileOffsetX+int(rect.X)+x, tileOffsetY+int(rect.Y)+y, palette[index])
		}
	}
	return err
}

func (enc *ZRLEEncoding) handlePlainRLETile(tileOffsetX int, tileOffsetY int, tileWidth int, tileHeight int, pf *PixelFormat, rect *Rectangle) error {
	var col *color.RGBA
	var err error
	runLen := 0
	for y := 0; y < tileHeight; y++ {
		for x := 0; x < tileWidth; x++ {

			if runLen == 0 {

				// Read length and color
				col, err = readCPixel(enc.unzipper, pf)
				if err != nil {
					logger.Errorf("handlePlainRLETile: error while reading CPixel in plain RLE subencoding: %v", err)
					return err
				}
				runLen, err = readRunLength(enc.unzipper)
				if err != nil {
					logger.Errorf("handlePlainRLETile: error while reading runlength in plain RLE subencoding: %v", err)
					return err
				}

			}

			// Write pixel to image
			enc.Image.Set(tileOffsetX+int(rect.X)+x, tileOffsetY+int(rect.Y)+y, col)
			runLen--
		}
	}
	return err
}

func readRunLength(r io.Reader) (int, error) {
	runLen := 1

	addition, err := ReadUint8(r)
	if err != nil {
		logger.Errorf("renderZRLE: error while reading addition to runLen in plain RLE subencoding: %v", err)
		return 0, err
	}
	runLen += int(addition)

	for addition == 255 {
		addition, err = ReadUint8(r)
		if err != nil {
			logger.Errorf("renderZRLE: error while reading addition to runLen in-loop plain RLE subencoding: %v", err)
			return 0, err
		}
		runLen += int(addition)
	}
	return runLen, nil
}

// Reads cpixel color from reader
func readCPixel(c io.Reader, pf *PixelFormat) (*color.RGBA, error) {
	if pf.TrueColor == 0 {
		return nil, errors.New("support for non true color formats was not implemented")
	}

	isZRLEFormat := IsCPixelSpecific(pf)
	var col *color.RGBA
	if isZRLEFormat {
		tbytes, err := ReadBytes(3, c)
		if err != nil {
			return nil, err
		}

		if pf.BigEndian != 1 {
			col = &color.RGBA{
				B: uint8(tbytes[0]),
				G: uint8(tbytes[1]),
				R: uint8(tbytes[2]),
				A: uint8(1),
			}
		} else {
			col = &color.RGBA{
				R: uint8(tbytes[0]),
				G: uint8(tbytes[1]),
				B: uint8(tbytes[2]),
				A: uint8(1),
			}
		}
		return col, nil
	}

	col, err := ReadColor(c, pf)
	if err != nil {
		logger.Errorf("readCPixel: Error while reading zrle: %v", err)
	}

	return col, nil
}
