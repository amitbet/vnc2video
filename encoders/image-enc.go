package encoders

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"github.com/amitbet/vnc2video"
)

func encodePPMGeneric(w io.Writer, img image.Image) error {
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

var convImage []uint8

func encodePPMforRGBA(w io.Writer, img *image.RGBA) error {
	maxvalue := 255
	size := img.Bounds()
	// write ppm header
	_, err := fmt.Fprintf(w, "P6\n%d %d\n%d\n", size.Dx(), size.Dy(), maxvalue)
	if err != nil {
		return err
	}

	if convImage == nil {
		convImage = make([]uint8, size.Dy()*size.Dx()*3)
	}

	rowCount := 0
	for i := 0; i < len(img.Pix); i++ {
		if (i % 4) != 3 {
			convImage[rowCount] = img.Pix[i]
			rowCount++
		}
	}

	if _, err := w.Write(convImage); err != nil {
		return err
	}

	return nil
}

func encodePPM(w io.Writer, img image.Image) error {
	if img == nil {
		return errors.New("nil image")
	}
	img1, isRGBImage := img.(*vnc2video.RGBImage)
	img2, isRGBA := img.(*image.RGBA)
	if isRGBImage {
		return encodePPMforRGBImage(w, img1)
	} else if isRGBA {
		return encodePPMforRGBA(w, img2)
	}
	return encodePPMGeneric(w, img)
}

func encodePPMforRGBImage(w io.Writer, img *vnc2video.RGBImage) error {
	maxvalue := 255
	size := img.Bounds()
	// write ppm header
	_, err := fmt.Fprintf(w, "P6\n%d %d\n%d\n", size.Dx(), size.Dy(), maxvalue)
	if err != nil {
		return err
	}

	if _, err := w.Write(img.Pix); err != nil {
		return err
	}
	return nil
}

type ImageEncoder interface {
	Init(string)
	Run()
	Encode(image.Image)
	Close()
}
