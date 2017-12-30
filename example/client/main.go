package main

import (
	"context"
	"image"
	"log"
	"net"
	"os"
	"time"
	vnc "vnc2webm"
	"vnc2webm/logger"
)

func main() {

	// Establish TCP connection to VNC server.
	nc, err := net.DialTimeout("tcp", os.Args[1], 5*time.Second)
	if err != nil {
		log.Fatalf("Error connecting to VNC host. %v", err)
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
		log.Fatalf("Error negotiating connection to VNC host. %v", err)
	}
	logger.Debugf("connected to: %s", os.Args[1])
	defer cc.Close()

	cc.SetEncodings([]vnc.EncodingType{vnc.EncTight})
	rect := image.Rect(0, 0, int(cc.Width()), int(cc.Height()))
	screenImage := image.NewRGBA64(rect)
	// Process messages coming in on the ServerMessage channel.
	for {
		select {
		case err := <-errorCh:
			panic(err)
		case msg := <-cchClient:
			log.Printf("Received client message type:%v msg:%v\n", msg.Type(), msg)
		case msg := <-cchServer:
			log.Printf("Received server message type:%v msg:%v\n", msg.Type(), msg)
			myRenderer, ok := msg.(vnc.Renderer)

			if ok {
				err = myRenderer.Render(screenImage)
				if err != nil {
					log.Printf("Received server message type:%v msg:%v\n", msg.Type(), msg)
				}
			}
			reqMsg := vnc.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: cc.Width(), Height: cc.Height()}
			reqMsg.Write(cc)
		}
	}
	cc.Wait()
}
