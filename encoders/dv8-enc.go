package encoders

import (
	"image"
	"io"
	"os"
	"os/exec"
	"strings"
	"github.com/amitbet/vnc2video/logger"
)

type VP8ImageEncoder struct {
	cmd           *exec.Cmd
	FFMpegBinPath string
	input         io.WriteCloser
	closed        bool
	Framerate     int
}

func (enc *VP8ImageEncoder) Init(videoFileName string) {
	fileExt := ".webm"
	if enc.Framerate == 0 {
		enc.Framerate = 12
	}
	if !strings.HasSuffix(videoFileName, fileExt) {
		videoFileName = videoFileName + fileExt
	}
	binary := "./ffmpeg"
	cmd := exec.Command(binary,
		"-f", "image2pipe",
		"-vcodec", "ppm",
		//"-r", strconv.Itoa(framerate),
		"-vsync", "2",
		"-r", "5",
		"-probesize", "10000000",
		"-an", //no audio
		//"-vsync", "2",
		///"-probesize", "10000000",
		"-y",
		//"-i", "pipe:0",
		"-i", "-",

		//"-crf", "4",
		"-vcodec", "libvpx", //"libvpx",//"libvpx-vp9"//"libx264"
		"-b:v", "0.5M",
		//"-maxrate", "1.5M",

		"-threads", "8",
		//"-speed", "0",
		//"-lossless", "1", //for vpx
		// "-tile-columns", "6",
		//"-frame-parallel", "1",
		// "-an", "-f", "webm",

		//"-preset", "ultrafast",
		//"-deadline", "realtime",
		"-quality", "good",
		"-cpu-used", "-16",
		"-minrate", "0.2M",
		"-maxrate", "0.7M",
		"-bufsize", "50M",
		"-g", "180",
		"-keyint_min", "180",
		"-rc_lookahead", "20",
		//"-crf", "34",
		//"-profile", "0",
		"-qmax", "51",
		"-qmin", "3",
		//"-slices", "4",
		//"-vb", "2M",

		videoFileName,
	)
	//cmd := exec.Command("/bin/echo")

	//io.Copy(cmd.Stdout, os.Stdout)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	encInput, err := cmd.StdinPipe()
	enc.input = encInput
	if err != nil {
		logger.Error("can't get ffmpeg input pipe")
	}
	enc.cmd = cmd
}
func (enc *VP8ImageEncoder) Run(videoFileName string) {
	if _, err := os.Stat(enc.FFMpegBinPath); os.IsNotExist(err) {
		logger.Error("encoder file doesn't exist in path:", enc.FFMpegBinPath)
		return
	}

	enc.Init(videoFileName)
	logger.Debugf("launching binary: %v", enc.cmd)
	err := enc.cmd.Run()
	if err != nil {
		logger.Errorf("error while launching ffmpeg: %v\n err: %v", enc.cmd.Args, err)
	}
}
func (enc *VP8ImageEncoder) Encode(img image.Image) {
	if enc.input == nil || enc.closed {
		return
	}

	err := encodePPM(enc.input, img)
	if err != nil {
		logger.Error("error while encoding image:", err)
	}
}

func (enc *VP8ImageEncoder) Close() {
	enc.closed = true
}
