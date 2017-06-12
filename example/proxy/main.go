package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"time"

	vnc "github.com/kward/go-vnc"
	"github.com/kward/go-vnc/logging"
	"github.com/kward/go-vnc/messages"
	"github.com/kward/go-vnc/rfbflags"
)

func main() {
	flag.Parse()
	logging.V(logging.FnDeclLevel)
	ln, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		log.Fatalf("Error listen. %v", err)
	}

	// Negotiate connection with the server.
	sch := make(chan vnc.ClientMessage)

	// handle client messages.
	vcc := vnc.NewServerConfig()
	vcc.Auth = []vnc.ServerAuth{&vnc.ServerAuthNone{}}
	vcc.ClientMessageCh = sch
	go vnc.Serve(context.Background(), ln, vcc)

	nc, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		log.Fatalf("Error connecting to VNC host. %v", err)
	}

	// Negotiate connection with the server.
	cch := make(chan vnc.ServerMessage)
	vc, err := vnc.Connect(context.Background(), nc,
		&vnc.ClientConfig{
			Auth:            []vnc.ClientAuth{&vnc.ClientAuthNone{}},
			ServerMessageCh: cch,
		})

	if err != nil {
		log.Fatalf("Error negotiating connection to VNC host. %v", err)
	}

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

	// Process messages coming in on the ClientMessage channel.
	for {
		msg := <-ch
		msg.Write(
	}
}
