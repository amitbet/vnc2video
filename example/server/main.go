package main

import (
	"context"
	"fmt"
	"image"
	"math"
	"net"
	"time"
	vnc "vnc2video"
	log "github.com/sirupsen/logrus"
)

func main() {
	ln, err := net.Listen("tcp", ":5900")
	if err != nil {
		log.Fatalf("Error listen. %v", err)
	}

	chServer := make(chan vnc.ClientMessage)
	chClient := make(chan vnc.ServerMessage)

	im := image.NewRGBA(image.Rect(0, 0, width, height))
	tick := time.NewTicker(time.Second / 2)
	defer tick.Stop()

	cfg := &vnc.ServerConfig{
		Width:  800,
		Height: 600,
		//VersionHandler:    vnc.ServerVersionHandler,
		//SecurityHandler:   vnc.ServerSecurityHandler,
		SecurityHandlers: []vnc.SecurityHandler{&vnc.ClientAuthNone{}},
		//ClientInitHandler: vnc.ServerClientInitHandler,
		//ServerInitHandler: vnc.ServerServerInitHandler,
		Encodings:       []vnc.Encoding{&vnc.RawEncoding{}},
		PixelFormat:     vnc.PixelFormat32bit,
		ClientMessageCh: chServer,
		ServerMessageCh: chClient,
		Messages:        vnc.DefaultClientMessages,
	}
	cfg.Handlers = vnc.DefaultServerHandlers
	go vnc.Serve(context.Background(), ln, cfg)

	// Process messages coming in on the ClientMessage channel.
	for {
		select {
		case <-tick.C:
			drawImage(im, 0)
			fmt.Printf("tick\n")
		case msg := <-chClient:
			switch msg.Type() {
			default:
				log.Debugf("11 Received message type:%v msg:%v\n", msg.Type(), msg)
			}
		case msg := <-chServer:
			switch msg.Type() {
			default:
				log.Debugf("22 Received message type:%v msg:%v\n", msg.Type(), msg)
			}
		}
	}
}

const (
	width  = 800
	height = 600
)

func drawImage(im *image.RGBA, anim int) {
	pos := 0
	const border = 50
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b uint8
			switch {
			case x < border*2.5 && x < int((1.1+math.Sin(float64(y+anim*2)/40))*border):
				r = 255
			case x > width-border*2.5 && x > width-int((1.1+math.Sin(math.Pi+float64(y+anim*2)/40))*border):
				g = 255
			case y < border*2.5 && y < int((1.1+math.Sin(float64(x+anim*2)/40))*border):
				r, g = 255, 255
			case y > height-border*2.5 && y > height-int((1.1+math.Sin(math.Pi+float64(x+anim*2)/40))*border):
				b = 255
			default:
				r, g, b = uint8(x+anim), uint8(y+anim), uint8(x+y+anim*3)
			}
			im.Pix[pos] = r
			im.Pix[pos+1] = g
			im.Pix[pos+2] = b
			pos += 4 // skipping alpha
		}
	}
}
