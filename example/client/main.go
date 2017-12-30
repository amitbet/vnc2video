package main

import (
	"context"
	"image"
	"net"
	"os"
	"time"
	vnc "vnc2webm"
	"vnc2webm/encoders"
	"vnc2webm/logger"
)

func main() {

	// Establish TCP connection to VNC server.
	nc, err := net.DialTimeout("tcp", os.Args[1], 5*time.Second)
	if err != nil {
		logger.Fatalf("Error connecting to VNC host. %v", err)
	}

	logger.Debugf("starting up the client, connecting to: %s", os.Args[1])
	// Negotiate connection with the server.
	cchServer := make(chan vnc.ServerMessage)
	cchClient := make(chan vnc.ClientMessage)
	errorCh := make(chan error)

	ccfg := &vnc.ClientConfig{
		SecurityHandlers: []vnc.SecurityHandler{
			//&vnc.ClientAuthATEN{Username: []byte(os.Args[2]), Password: []byte(os.Args[3])}
			&vnc.ClientAuthVNC{Password: []byte("12345")},
			&vnc.ClientAuthNone{},
		},
		PixelFormat:     vnc.PixelFormat32bit,
		ClientMessageCh: cchClient,
		ServerMessageCh: cchServer,
		Messages:        vnc.DefaultServerMessages,
		Encodings:       []vnc.Encoding{&vnc.RawEncoding{}, &vnc.TightEncoding{}},
		ErrorCh:         errorCh,
	}

	cc, err := vnc.Connect(context.Background(), nc, ccfg)
	if err != nil {
		logger.Fatalf("Error negotiating connection to VNC host. %v", err)
	}
	// out, err := os.Create("./output" + strconv.Itoa(counter) + ".jpg")
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }
	//vcodec := &encoders.MJPegImageEncoder{Quality: 60, Framerate: 6}
	vcodec := &encoders.X264ImageEncoder{}
	//vcodec := &encoders.DV8ImageEncoder{}
	//vcodec := &encoders.DV9ImageEncoder{}

	//counter := 0
	//vcodec.Init("./output" + strconv.Itoa(counter))
	go vcodec.Run("./ffmpeg", "./output.mp4")

	screenImage := image.NewRGBA(image.Rect(0, 0, int(cc.Width()), int(cc.Height())))
	for _, enc := range ccfg.Encodings {
		myRenderer, ok := enc.(vnc.Renderer)

		if ok {
			myRenderer.SetTargetImage(screenImage)
		}
	}
	// var out *os.File

	logger.Debugf("connected to: %s", os.Args[1])
	defer cc.Close()

	cc.SetEncodings([]vnc.EncodingType{vnc.EncTight})
	//rect := image.Rect(0, 0, int(cc.Width()), int(cc.Height()))
	//screenImage := image.NewRGBA64(rect)
	// Process messages coming in on the ServerMessage channel.
	for {
		select {
		case err := <-errorCh:
			panic(err)
		case msg := <-cchClient:
			logger.Debugf("Received client message type:%v msg:%v\n", msg.Type(), msg)
		case msg := <-cchServer:
			logger.Debugf("Received server message type:%v msg:%v\n", msg.Type(), msg)

			// out, err := os.Create("./output" + strconv.Itoa(counter) + ".jpg")
			// if err != nil {
			// 	fmt.Println(err)
			// 	os.Exit(1)
			// }
			if msg.Type() == vnc.FramebufferUpdateMsgType {
				//counter++
				//jpeg.Encode(out, screenImage, nil)
				vcodec.Encode(screenImage)
				reqMsg := vnc.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: cc.Width(), Height: cc.Height()}
				reqMsg.Write(cc)
			}
		}
	}
	cc.Wait()
}
