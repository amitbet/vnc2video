package vnc2video

import (
	"image"
	"image/color"
	"image/draw"
)

func FillRect(img draw.Image, rect *image.Rectangle, c color.Color) {
	for x := 0; x < rect.Max.X; x++ {
		for y := 0; y < rect.Max.Y; y++ {
			img.Set(x, y, c)
		}
	}
}

func DrawLine(img draw.Image, rect *image.Rectangle, c color.Color) {

}
