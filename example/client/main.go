package main

import (
	"context"
	"log"
	"net"
	"time"

	vnc "github.com/kward/go-vnc"
	"github.com/kward/go-vnc/logging"
	"github.com/kward/go-vnc/messages"
	"github.com/kward/go-vnc/rfbflags"
)

func main() {
	logging.V(logging.FnDeclLevel)
	// Establish TCP connection to VNC server.
	nc, err := net.Dial("tcp", "127.0.0.1:5923")
	if err != nil {
		log.Fatalf("Error connecting to VNC host. %v", err)
	}

	// Negotiate connection with the server.
	ch := make(chan vnc.ServerMessage)
	vc, err := vnc.Connect(context.Background(), nc,
		&vnc.ClientConfig{
			Auth:            []vnc.ClientAuth{&vnc.ClientAuthNone{}},
			ServerMessageCh: ch,
		})

	if err != nil {
		log.Fatalf("Error negotiating connection to VNC host. %v", err)
	}

	// Periodically request framebuffer updates.
	go func() {
		w, h := vc.FramebufferWidth(), vc.FramebufferHeight()
		for {
			if err := vc.FramebufferUpdateRequest(rfbflags.RFBTrue, 0, 0, w, h); err != nil {
				log.Printf("error requesting framebuffer update: %v", err)
				return
			}
			time.Sleep(1 * time.Second)
		}
	}()

	// Listen and handle server messages.
	go vc.ListenAndHandle()

	// Process messages coming in on the ServerMessage channel.
	for {
		msg := <-ch
		switch msg.Type() {
		case messages.FramebufferUpdate:
			log.Println("Received FramebufferUpdate message.")
		default:
			log.Printf("Received message type:%v msg:%v\n", msg.Type(), msg)
		}
	}
}
