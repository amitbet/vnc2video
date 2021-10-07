package vnc2video

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"sync"
	"time"

	"github.com/matts1/vnc2video/logger"
)

//go:generate stringer -type=TightCompression

type TightCompression uint8

const (
	TightCompressionBasic = 0
	TightCompressionFill  = 8
	TightCompressionJPEG  = 9
	TightCompressionPNG   = 10
)

//go:generate stringer -type=TightFilter

type TightFilter uint8

const (
	TightFilterCopy     = 0
	TightFilterPalette  = 1
	TightFilterGradient = 2
)

type drawOp struct {
	op func(draw.Image) error
	ts time.Time
}

type TightEncoding struct {
	decoders     []io.Reader
	decoderBuffs []*bytes.Buffer
	ops          []drawOp
	closed       bool
	mutex        *sync.Mutex
}

func NewTightEncoder() TightEncoding {
	return TightEncoding{mutex: &sync.Mutex{}}
}

var instance *TightEncoding
var TightMinToCompress int = 12

func (enc *TightEncoding) AddOp(op func(draw.Image) error) {
	enc.mutex.Lock()
	if !enc.closed {
		enc.ops = append(enc.ops, drawOp{op: op, ts: time.Now()})
	}
	enc.mutex.Unlock()
}

func (enc *TightEncoding) DrawUntilTime(img draw.Image, ts time.Time) error {
	// Ensure the other thread isn't modifying ops.
	enc.mutex.Lock()
	enc.closed = true
	enc.mutex.Unlock()
	// The slice doesn't modify the underlying array, so it's fast.
	for ; len(enc.ops) > 0 && enc.ops[0].ts.Before(ts); enc.ops = enc.ops[1:] {
		if err := enc.ops[0].op(img); err != nil {
			return err
		}
	}
	return nil
}

func (*TightEncoding) Supported(Conn) bool {
	return true
}

func (*TightEncoding) Type() EncodingType { return EncTight }

func (*TightEncoding) GetInstance() *TightEncoding {
	if instance == nil {
		instance = &TightEncoding{}
	}
	return instance
}

func (enc *TightEncoding) Write(c Conn, rect *Rectangle) error {
	return nil
}

// Read unmarshal color from conn
func getTightColor(c io.Reader, pf *PixelFormat) (*color.RGBA, error) {

	if pf.TrueColor == 0 {
		return nil, errors.New("support for non true color formats was not implemented")
	}
	order := pf.order()
	var pixel uint32
	isTightFormat := pf.TrueColor != 0 && pf.Depth == 24 && pf.BPP == 32 && pf.BlueMax <= 255 && pf.RedMax <= 255 && pf.GreenMax <= 255
	if isTightFormat {
		//tbytes := make([]byte, 3)
		tbytes, err := ReadBytes(3, c)
		if err != nil {
			return nil, err
		}
		rgb := color.RGBA{
			R: uint8(tbytes[0]),
			G: uint8(tbytes[1]),
			B: uint8(tbytes[2]),
			A: uint8(1),
		}
		return &rgb, nil
	}

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

func calcTightBytePerPixel(pf *PixelFormat) int {
	bytesPerPixel := int(pf.BPP / 8)

	var bytesPerPixelTight int
	if 24 == pf.Depth && 32 == pf.BPP {
		bytesPerPixelTight = 3
	} else {
		bytesPerPixelTight = bytesPerPixel
	}
	return bytesPerPixelTight
}

func (enc *TightEncoding) Reset() error {
	//enc.decoders = make([]io.Reader, 4)
	//enc.decoderBuffs = make([]*bytes.Buffer, 4)
	return nil
}

func (enc *TightEncoding) resetDecoders(compControl uint8) {
	logger.Tracef("###resetDecoders compctl :%d", 0x0F&compControl)
	for i := 0; i < 4; i++ {
		if (compControl&1) != 0 && enc.decoders[i] != nil {
			logger.Tracef("###resetDecoders - resetting decoder #%d", i)
			enc.decoders[i] = nil //.(zlib.Resetter).Reset(nil,nil);
		}
		compControl >>= 1
	}
}

var counter int = 0
var disablePalette bool = false
var disableGradient bool = false
var disableCopy bool = false
var disableJpeg bool = false
var disableFill bool = false

func (enc *TightEncoding) Read(c Conn, rect *Rectangle) error {

	var err error
	////////////
	// if counter > 40 {
	// 	os.Exit(1)
	// }
	////////////
	pixelFmt := c.PixelFormat()
	bytesPixel := calcTightBytePerPixel(&pixelFmt)

	compctl, err := ReadUint8(c)

	/////////////////
	// var out *os.File
	// if out == nil {
	// 	out, err = os.Create("./output" + strconv.Itoa(counter) + "-" + strconv.Itoa(int(compctl)) + ".jpg")
	// 	if err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(1)
	// 	}
	// }
	// defer func() { counter++ }()
	// defer jpeg.Encode(out, enc.Image, nil)
	//////////////
	logger.Tracef("-----------READ-Tight-encoding compctl=%d -------------", compctl)

	if err != nil {
		return fmt.Errorf("error in handling tight encoding: %w", err)
	}
	//logger.Tracef("bytesPixel= %d, subencoding= %d", bytesPixel, compctl)
	enc.resetDecoders(compctl)

	//move it to position (remove zlib flush commands)
	compType := compctl >> 4 & 0x0F

	switch compType {
	case TightCompressionFill:
		logger.Tracef("--TIGHT_FILL: reading fill size=%d,counter=%d", bytesPixel, counter)
		//read color

		rectColor, err := getTightColor(c, &pixelFmt)
		if err != nil {
			return fmt.Errorf("error in reading tight encoding: %w", err)
		}

		myRect := MakeRectFromVncRect(rect)
		logger.Tracef("--TIGHT_FILL: fill rect=%v,color=%v", myRect, rectColor)
		if !disableFill {
			enc.AddOp(func(img draw.Image) error {
				FillRect(img, &myRect, rectColor)
				return nil
			})
		}

		if bytesPixel != 3 {
			return fmt.Errorf("non tight bytesPerPixel format, should be 3 bytes")
		}
		return nil
	case TightCompressionJPEG:
		logger.Tracef("--TIGHT_JPEG,counter=%d", counter)
		if pixelFmt.BPP == 8 {
			return errors.New("Tight encoding: JPEG is not supported in 8 bpp mode")
		}

		len, err := readTightLength(c)

		if err != nil {
			return err
		}
		jpegBytes, err := ReadBytes(len, c)
		if err != nil {
			return err
		}

		enc.AddOp(func(img draw.Image) error {
			buff := bytes.NewBuffer(jpegBytes)
			decoded, err := jpeg.Decode(buff)
			if err != nil {
				return err
			}
			if !disableJpeg {
				pos := image.Point{int(rect.X), int(rect.Y)}
				DrawImage(img, decoded, pos)
			}
			return nil
		})

		return nil
	default:

		if compType > TightCompressionJPEG {
			return fmt.Errorf("Compression control byte is incorrect!")
		}

		return enc.handleTightFilters(compctl, &pixelFmt, rect, c)
	}
}

func (enc *TightEncoding) handleTightFilters(compCtl uint8, pixelFmt *PixelFormat, rect *Rectangle, r Conn) error {

	var STREAM_ID_MASK uint8 = 0x30
	var FILTER_ID_MASK uint8 = 0x40

	var filterid uint8
	var err error

	decoderId := (compCtl & STREAM_ID_MASK) >> 4

	for len(enc.decoders) < 4 {
		enc.decoders = append(enc.decoders, nil)
		enc.decoderBuffs = append(enc.decoderBuffs, nil)
	}

	if (compCtl & FILTER_ID_MASK) > 0 {
		filterid, err = ReadUint8(r)

		if err != nil {
			return fmt.Errorf("error in handling tight encoding, reading filterid: %w", err)
		}
	}

	bytesPixel := calcTightBytePerPixel(pixelFmt)

	lengthCurrentbpp := int(bytesPixel) * int(rect.Width) * int(rect.Height)

	switch filterid {
	case TightFilterPalette: //PALETTE_FILTER

		palette, err := enc.readTightPalette(r, bytesPixel)
		if err != nil {
			return fmt.Errorf("handleTightFilters: error in Reading Palette: %w", err)
		}
		logger.Debugf("----PALETTE_FILTER,palette len=%d counter=%d, rect= %v", len(palette), counter, rect)

		var dataLength int
		if len(palette) == 2 {
			dataLength = int(rect.Height) * ((int(rect.Width) + 7) / 8)
		} else {
			dataLength = int(rect.Width) * int(rect.Height)
		}
		tightBytes, err := enc.ReadTightData(dataLength, r, int(decoderId))
		if err != nil {
			return fmt.Errorf("handleTightFilters: error in handling tight encoding, reading palette filter data: %w", err)
		}
		if !disablePalette {
			enc.drawTightPalette(rect, palette, tightBytes)
		}
	case TightFilterGradient: //GRADIENT_FILTER
		logger.Debugf("----GRADIENT_FILTER: bytesPixel=%d, counter=%d", bytesPixel, counter)
		data, err := enc.ReadTightData(lengthCurrentbpp, r, int(decoderId))
		if err != nil {
			return fmt.Errorf("handleTightFilters: error in handling tight encoding, Reading GRADIENT_FILTER: %w", err)
		}

		enc.decodeGradData(rect, data)

	case TightFilterCopy: //BASIC_FILTER
		logger.Debugf("----BASIC_FILTER: bytesPixel=%d, counter=%d", bytesPixel, counter)

		tightBytes, err := enc.ReadTightData(lengthCurrentbpp, r, int(decoderId))
		if err != nil {
			return fmt.Errorf("handleTightFilters: error in handling tight encoding, Reading BASIC_FILTER: %w", err)
		}
		logger.Tracef("tightBytes len= %d", len(tightBytes))
		if !disableCopy {
			enc.drawTightBytes(tightBytes, rect)
		}
	default:
		return fmt.Errorf("handleTightFilters: Bad tight filter id: %d", filterid)
	}

	return nil
}

func (enc *TightEncoding) drawTightPalette(rect *Rectangle, palette color.Palette, tightBytes []byte) {
	bytePos := 0
	bitPos := uint8(7)
	var palettePos int
	logger.Tracef("drawTightPalette numbytes=%d", len(tightBytes))

	enc.AddOp(func(img draw.Image) error {
		for y := 0; y < int(rect.Height); y++ {
			for x := 0; x < int(rect.Width); x++ {
				if len(palette) == 2 {
					currByte := tightBytes[bytePos]
					mask := byte(1) << bitPos

					palettePos = 0
					if currByte&mask > 0 {
						palettePos = 1
					}

					if bitPos == 0 {
						bytePos++
					}
					bitPos = ((bitPos - 1) + 8) % 8
				} else {
					palettePos = int(tightBytes[bytePos])
					bytePos++
				}
				img.Set(int(rect.X)+x, int(rect.Y)+y, palette[palettePos])
			}

			// reset bit alignment to first bit in byte (msb)
			bitPos = 7
		}
		return nil
	})
}
func (enc *TightEncoding) decodeGradData(rect *Rectangle, buffer []byte) {
	enc.AddOp(func(img draw.Image) error {

		prevRow := make([]byte, rect.Width*3+3) //new byte[w * 3];
		thisRow := make([]byte, rect.Width*3+3) //new byte[w * 3];

		bIdx := 0

		for i := 0; i < int(rect.Height); i++ {
			for j := 3; j < int(rect.Width*3+3); j += 3 {
				d := int(0xff&prevRow[j]) + // "upper" pixel (from prev row)
					int(0xff&thisRow[j-3]) - // prev pixel
					int(0xff&prevRow[j-3]) // "diagonal" prev pixel
				if d < 0 {
					d = 0
				}
				if d > 255 {
					d = 255
				}
				red := int(buffer[bIdx]) + d
				thisRow[j] = byte(red & 255)

				d = int(0xff&prevRow[j+1]) +
					int(0xff&thisRow[j+1-3]) -
					int(0xff&prevRow[j+1-3])
				if d < 0 {
					d = 0
				}
				if d > 255 {
					d = 255
				}
				green := int(buffer[bIdx+1]) + d
				thisRow[j+1] = byte(green & 255)

				d = int(0xff&prevRow[j+2]) +
					int(0xff&thisRow[j+2-3]) -
					int(0xff&prevRow[j+2-3])
				if d < 0 {
					d = 0
				}
				if d > 255 {
					d = 255
				}
				blue := int(buffer[bIdx+2]) + d
				thisRow[j+2] = byte(blue & 255)

				bIdx += 3
			}

			for idx := 3; idx < (len(thisRow) - 3); idx += 3 {
				myColor := color.RGBA{R: (thisRow[idx]), G: (thisRow[idx+1]), B: (thisRow[idx+2]), A: 1}
				if !disableGradient {
					img.Set(idx/3+int(rect.X)-1, int(rect.Y)+i, myColor)
				}
			}

			// exchange thisRow and prevRow:
			tempRow := thisRow
			thisRow = prevRow
			prevRow = tempRow
		}
		return nil
	})
}

func ReadBytes(count int, r io.Reader) ([]byte, error) {
	buff := make([]byte, count)

	lengthRead, err := io.ReadFull(r, buff)

	if lengthRead != count {
		return nil, fmt.Errorf("RfbReadHelper.ReadBytes unable to read bytes: lengthRead=%d, countExpected=%d", lengthRead, count)
	}

	if err != nil {
		return nil, fmt.Errorf("RfbReadHelper.ReadBytes error while reading bytes: %w", err)
	}

	return buff, nil
}

func (enc *TightEncoding) readTightPalette(connReader Conn, bytesPixel int) (color.Palette, error) {

	colorCount, err := ReadUint8(connReader)
	if err != nil {
		return nil, fmt.Errorf("handleTightFilters: error in handling tight encoding, reading TightFilterPalette: %w", err)
	}

	paletteSize := colorCount + 1 // add one more
	//complete palette
	paletteColorBytes, err := ReadBytes(int(paletteSize)*bytesPixel, connReader)
	if err != nil {
		return nil, fmt.Errorf("handleTightFilters: error in handling tight encoding, reading TightFilterPalette.paletteSize: %w", err)
	}
	var paletteColors color.Palette = make([]color.Color, 0)
	for i := 0; i < int(paletteSize)*bytesPixel; i += 3 {
		col := color.RGBA{R: paletteColorBytes[i], G: paletteColorBytes[i+1], B: paletteColorBytes[i+2], A: 1}
		paletteColors = append(paletteColors, col)
	}
	return paletteColors, nil
}

func (enc *TightEncoding) ReadTightData(dataSize int, c Conn, decoderId int) ([]byte, error) {

	logger.Tracef(">>> Reading zipped tight data from decoder Id: %d, openSize: %d", decoderId, dataSize)
	if int(dataSize) < TightMinToCompress {
		return ReadBytes(int(dataSize), c)
	}
	zlibDataLen, err := readTightLength(c)
	if err != nil {
		return nil, err
	}
	zippedBytes, err := ReadBytes(zlibDataLen, c)
	if err != nil {
		return nil, err
	}
	var r io.Reader
	if enc.decoders[decoderId] == nil {
		b := bytes.NewBuffer(zippedBytes)
		r, err = zlib.NewReader(b)
		if err != nil {
			return nil, err
		}
		enc.decoders[decoderId] = r
		enc.decoderBuffs[decoderId] = b
	} else {
		b := enc.decoderBuffs[decoderId]

		b.Write(zippedBytes) //set the underlaying buffer to new content (not resetting the decoder zlib stream)
		r = enc.decoders[decoderId]
	}

	retBytes := make([]byte, dataSize)
	count, err := io.ReadFull(r, retBytes)
	if err != nil {
		return nil, err
	}
	if count != dataSize {
		return nil, errors.New("ReadTightData: reading inflating zip didn't produce expected number of bytes")
	}
	return retBytes, nil
}

type TightCC struct {
	Compression TightCompression
	Filter      TightFilter
}

type TightPixel struct {
	R uint8
	G uint8
	B uint8
}

func readTightLength(c Conn) (int, error) {
	var length int
	var err error
	var b uint8

	if err = binary.Read(c, binary.BigEndian, &b); err != nil {
		return 0, err
	}

	length = int(b) & 0x7F
	if (b & 0x80) == 0 {
		return length, nil
	}

	if err = binary.Read(c, binary.BigEndian, &b); err != nil {
		return 0, err
	}
	length |= (int(b) & 0x7F) << 7
	if (b & 0x80) == 0 {
		return length, nil
	}

	if err = binary.Read(c, binary.BigEndian, &b); err != nil {
		return 0, err
	}
	length |= (int(b) & 0xFF) << 14

	return length, nil
}

/**
 * Draw byte array bitmap data (for Tight)
 */
func (enc *TightEncoding) drawTightBytes(bytes []byte, rect *Rectangle) {
	logger.Tracef("drawTightBytes: len(bytes)= %d, %v", len(bytes), rect)

	enc.AddOp(func(img draw.Image) error {
		bytesPos := 0
		for ly := rect.Y; ly < rect.Y+rect.Height; ly++ {
			for lx := rect.X; lx < rect.X+rect.Width; lx++ {
				color := color.RGBA{R: bytes[bytesPos], G: bytes[bytesPos+1], B: bytes[bytesPos+2], A: 1}
				img.Set(int(lx), int(ly), color)

				bytesPos += 3
			}
		}
		return nil
	})
}
