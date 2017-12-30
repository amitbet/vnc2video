package encoders

import (
	"bytes"
	"image"
	"image/jpeg"
	"strings"
	"vnc2webm/logger"

	"github.com/icza/mjpeg"
)

type MJPegImageEncoder struct {
	avWriter  mjpeg.AviWriter
	Quality   int
	Framerate int32
}

func (enc *MJPegImageEncoder) Init(videoFileName string) {
	fileExt := ".avi"
	if !strings.HasSuffix(videoFileName, fileExt) {
		videoFileName = videoFileName + fileExt
	}
	if enc.Framerate <= 0 {
		enc.Framerate = 5
	}
	avWriter, err := mjpeg.New(videoFileName, 1024, 768, enc.Framerate)
	if err != nil {
		logger.Error("Error during mjpeg init: ", err)
	}
	enc.avWriter = avWriter
}
func (enc *MJPegImageEncoder) Run() {
}
func (enc *MJPegImageEncoder) Encode(img image.Image) {
	buf := &bytes.Buffer{}
	jOpts := &jpeg.Options{Quality: enc.Quality}
	if enc.Quality <= 0 {
		jOpts = nil
	}
	err := jpeg.Encode(buf, img, jOpts)
	if err != nil {
		logger.Error("Error while creating jpeg: ", err)
	}
	err = enc.avWriter.AddFrame(buf.Bytes())
	if err != nil {
		logger.Error("Error while adding frame to mjpeg: ", err)
	}

}
func (enc *MJPegImageEncoder) Close() {
	err := enc.avWriter.Close()
	if err != nil {
		logger.Error("Error while closing mjpeg: ", err)
	}
}
