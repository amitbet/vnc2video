package vnc2video

import (
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
)

const (
	BlockWidth  = 16
	BlockHeight = 16
)

type VncCanvas struct {
	draw.Image
	//DisplayBuff draw.Image
	//WriteBuff      draw.Image
	imageBuffs     [2]draw.Image
	Cursor         draw.Image
	CursorMask     [][]bool
	CursorBackup   draw.Image
	CursorOffset   *image.Point
	CursorLocation *image.Point
	DrawCursor     bool
	Changed        map[string]bool
}

func NewVncCanvas(width, height int) *VncCanvas {
	//dispImg := NewRGBImage(image.Rect(0, 0, width, height))
	writeImg := NewRGBImage(image.Rect(0, 0, width, height))
	canvas := VncCanvas{
		Image: writeImg,
		//DisplayBuff: dispImg,
		//WriteBuff:   writeImg,
	}
	return &canvas
}

func (c *VncCanvas) SetChanged(rect *Rectangle) {
	if c.Changed == nil {
		c.Changed = make(map[string]bool)
	}
	for x := int(rect.X) / BlockWidth; x*BlockWidth < int(rect.X+rect.Width); x++ {
		for y := int(rect.Y) / BlockHeight; y*BlockHeight < int(rect.Y+rect.Height); y++ {
			key := fmt.Sprintf("%d,%d", x, y)
			//fmt.Println("setting block: ", key)
			c.Changed[key] = true
		}
	}
}

func (c *VncCanvas) Reset(rect *Rectangle) {
	c.Changed = nil
}

func (c *VncCanvas) RemoveCursor() image.Image {
	if c.Cursor == nil || c.CursorLocation == nil {
		return c.Image
	}
	if !c.DrawCursor {
		return c.Image
	}
	rect := c.Cursor.Bounds()
	loc := c.CursorLocation
	img := c.Image
	for y := rect.Min.Y; y < int(rect.Max.Y); y++ {
		for x := rect.Min.X; x < int(rect.Max.X); x++ {
			// offset := y*int(rect.Width) + x
			// if bitmask[y*int(scanLine)+x/8]&(1<<uint(7-x%8)) > 0 {
			col := c.CursorBackup.At(x, y)
			//mask := c.CursorMask.At(x, y).(color.RGBA)
			mask := c.CursorMask[x][y]
			//log.Info("Drawing Cursor: ", x, y, col, mask)
			if mask {
				//log.Info("Drawing Cursor for real: ", x, y, col)
				img.Set(x+loc.X-c.CursorOffset.X, y+loc.Y-c.CursorOffset.Y, col)
			}
			// 	//log.Debugf("CursorPseudoEncoding.Read: setting pixel: (%d,%d) %v", x+int(rect.X), y+int(rect.Y), colors[offset])
			// }
		}
	}
	return img
}

// func (c *VncCanvas) SwapBuffers() {
// 	swapSpace := c.DisplayBuff
// 	c.DisplayBuff = c.WriteBuff
// 	c.WriteBuff = swapSpace
// 	c.Image = c.WriteBuff
// }

func (c *VncCanvas) PaintCursor() image.Image {
	if c.Cursor == nil || c.CursorLocation == nil {
		return c.Image
	}
	if !c.DrawCursor {
		return c.Image
	}
	rect := c.Cursor.Bounds()
	if c.CursorBackup == nil {
		c.CursorBackup = image.NewRGBA(c.Cursor.Bounds())
	}

	loc := c.CursorLocation
	img := c.Image
	for y := rect.Min.Y; y < int(rect.Max.Y); y++ {
		for x := rect.Min.X; x < int(rect.Max.X); x++ {
			// offset := y*int(rect.Width) + x
			// if bitmask[y*int(scanLine)+x/8]&(1<<uint(7-x%8)) > 0 {
			col := c.Cursor.At(x, y)
			//mask := c.CursorMask.At(x, y).(RGBColor)
			mask := c.CursorMask[x][y]
			backup := c.Image.At(x+loc.X-c.CursorOffset.X, y+loc.Y-c.CursorOffset.Y)
			//c.CursorBackup.Set(x, y, backup)
			//backup the previous data at this point

			//log.Info("Drawing Cursor: ", x, y, col, mask)
			if mask {

				c.CursorBackup.Set(x, y, backup)
				//log.Info("Drawing Cursor for real: ", x, y, col)
				img.Set(x+loc.X-c.CursorOffset.X, y+loc.Y-c.CursorOffset.Y, col)
			}
			// 	//log.Debugf("CursorPseudoEncoding.Read: setting pixel: (%d,%d) %v", x+int(rect.X), y+int(rect.Y), colors[offset])
			// }
		}
	}
	return img
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func DrawImage(target draw.Image, imageToApply image.Image, pos image.Point) {
	rect := imageToApply.Bounds()
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			target.Set(x+pos.X, y+pos.Y, imageToApply.At(x, y))
		}
	}
}

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

func ReadUint8(r io.Reader) (uint8, error) {
	var myUint uint8
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}
func ReadUint16(r io.Reader) (uint16, error) {
	var myUint uint16
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}

func ReadUint32(r io.Reader) (uint32, error) {
	var myUint uint32
	if err := binary.Read(r, binary.BigEndian, &myUint); err != nil {
		return 0, err
	}

	return myUint, nil
}

func MakeRect(x, y, width, height int) image.Rectangle {
	return image.Rectangle{Min: image.Point{X: x, Y: y}, Max: image.Point{X: x + width, Y: y + height}}
}
func MakeRectFromVncRect(rect *Rectangle) image.Rectangle {
	return MakeRect(int(rect.X), int(rect.Y), int(rect.Width), int(rect.Height))
}
