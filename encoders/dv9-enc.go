package encoders

import (
	"image"
	"io"
	"os"
	"os/exec"
	"strings"
	"github.com/amitbet/vnc2video/logger"
)

type DV9ImageEncoder struct {
	cmd           *exec.Cmd
	FFMpegBinPath string
	input         io.WriteCloser
	Framerate     int
}

func (enc *DV9ImageEncoder) Init(videoFileName string) {
	fileExt := ".mp4"
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
		"-r", "5",
		//"-i", "pipe:0",
		"-i", "-",
		"-vcodec", "libvpx-vp9", //"libvpx",//"libvpx-vp9"//"libx264"
		"-b:v", "1M",
		"-threads", "8",
		//"-speed", "0",
		//"-lossless", "1", //for vpx
		// "-tile-columns", "6",
		//"-frame-parallel", "1",
		// "-an", "-f", "webm",
		"-cpu-used", "-8",

		//"-preset", "ultrafast",
		"-deadline", "realtime",
		//"-cpu-used", "-5",
		"-maxrate", "2.5M",
		"-bufsize", "10M",
		"-g", "120",

		//"-rc_lookahead", "16",
		//"-profile", "0",
		"-qmax", "51",
		"-qmin", "11",
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
func (enc *DV9ImageEncoder) Run(videoFileName string) {
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
func (enc *DV9ImageEncoder) Encode(img image.Image) {
	err := encodePPM(enc.input, img)
	if err != nil {
		logger.Error("error while encoding image:", err)
	}
}
func (enc *DV9ImageEncoder) Close() {

}
