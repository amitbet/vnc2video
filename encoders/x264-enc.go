package encoders

import (
	"errors"
	"image"
	"io"
	"os"
	"os/exec"
	"strings"
	"vnc2video/logger"
)

type X264ImageEncoder struct {
	cmd        *exec.Cmd
	binaryPath string
	input      io.WriteCloser
}

func (enc *X264ImageEncoder) Init(videoFileName string) {
	fileExt := ".mp4"
	if !strings.HasSuffix(videoFileName, fileExt) {
		videoFileName = videoFileName + fileExt
	}
	//binary := "./ffmpeg"
	cmd := exec.Command(enc.binaryPath,
		"-f", "image2pipe",
		"-vcodec", "ppm",
		//"-r", strconv.Itoa(framerate),
		"-r", "5",
		//"-re",
		//"-i", "pipe:0",

		"-vsync", "2",
		///"-probesize", "10000000",
		"-y",
		"-i", "-",
		"-vcodec", "libx264", //"libvpx",//"libvpx-vp9"//"libx264"
		"-b:v", "0.5M",
		"-threads", "8",
		//"-speed", "0",
		//"-lossless", "1", //for vpx
		// "-an", "-f", "webm",
		"-preset", "veryfast",
		"-tune", "animation",
		"-maxrate", "0.6M",
		"-bufsize", "50M",
		"-g", "120",

		//"-crf", "0",  //for lossless encoding!!!!

		//"-rc_lookahead", "16",
		//"-profile", "0",
		//"-crf", "18",
		"-qmax", "51",
		"-qmin", "7",
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
func (enc *X264ImageEncoder) Run(encoderFilePath string, videoFileName string) error {
	if _, err := os.Stat(encoderFilePath); os.IsNotExist(err) {
		logger.Error("encoder file doesn't exist in path:", encoderFilePath)
		return errors.New("encoder file doesn't exist in path" + videoFileName)
	}

	enc.binaryPath = encoderFilePath
	enc.Init(videoFileName)
	logger.Debugf("launching binary: %v", enc.cmd)
	err := enc.cmd.Run()
	if err != nil {
		logger.Errorf("error while launching ffmpeg: %v\n err: %v", enc.cmd.Args, err)
		return err
	}
	return nil
}
func (enc *X264ImageEncoder) Encode(img image.Image) {
	err := encodePPM(enc.input, img)
	if err != nil {
		logger.Error("error while encoding image:", err)
	}
}
func (enc *X264ImageEncoder) Close() {

}
