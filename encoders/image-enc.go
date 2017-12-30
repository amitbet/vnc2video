package encoders

import (
	"fmt"
	"image"
	"image/color"
	"io"
)

func encodePPM(w io.Writer, img image.Image) error {
	maxvalue := 255
	size := img.Bounds()
	// write ppm header
	_, err := fmt.Fprintf(w, "P6\n%d %d\n%d\n", size.Dx(), size.Dy(), maxvalue)
	if err != nil {
		return err
	}

	// write the bitmap
	colModel := color.RGBAModel
	row := make([]uint8, size.Dx()*3)
	for y := size.Min.Y; y < size.Max.Y; y++ {
		i := 0
		for x := size.Min.X; x < size.Max.X; x++ {
			color := colModel.Convert(img.At(x, y)).(color.RGBA)
			row[i] = color.R
			row[i+1] = color.G
			row[i+2] = color.B
			i += 3
		}
		if _, err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

type ImageEncoder interface {
	Init(string)
	Run()
	Encode(image.Image)
	Close()
}
