package main

import (
	"image"
	"os"
	vnc "vnc2webm"
	"vnc2webm/encoders"
	"vnc2webm/logger"
	"path/filepath"
)

func main() {

	if len(os.Args) <= 1 {
		logger.Errorf("please provide a fbs file name")
		return
	}
	if _, err := os.Stat(os.Args[1]); os.IsNotExist(err) {
		logger.Errorf("File doesn't exist", err)
		return
	}
	encs := []vnc.Encoding{
		&vnc.RawEncoding{},
		&vnc.TightEncoding{},
	}

	fbs, err := vnc.NewFbsConn(
		os.Args[1],
		encs,
	)
	if err != nil {
		logger.Error("failed to open fbs reader:", err)
		//return nil, err
	}

	//launch video encoding process:
	vcodec := &encoders.X264ImageEncoder{}
	//vcodec := &encoders.DV8ImageEncoder{}
	//vcodec := &encoders.DV9ImageEncoder{}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	logger.Debugf("current dir: %s", dir)
	go vcodec.Run("./ffmpeg", "./output.mp4")

	screenImage := image.NewRGBA(image.Rect(0, 0, int(fbs.Width()), int(fbs.Height())))
	for _, enc := range encs {
		myRenderer, ok := enc.(vnc.Renderer)

		if ok {
			myRenderer.SetTargetImage(screenImage)
		}
	}


	msgReader := vnc.NewFBSPlayHelper(fbs)

	//loop over all messages, feed images to video codec:
	for {
		msgReader.ReadFbsMessage()
		vcodec.Encode(screenImage)
	}
}
