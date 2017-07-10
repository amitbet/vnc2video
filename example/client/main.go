package main

import (
	"context"
	"log"
	"net"
	"os"
	"time"

	vnc "github.com/vtolstov/go-vnc"
)

func main() {
	// Establish TCP connection to VNC server.
	nc, err := net.DialTimeout("tcp", os.Args[1], 5*time.Second)
	if err != nil {
		log.Fatalf("Error connecting to VNC host. %v", err)
	}

	// Negotiate connection with the server.
	cchServer := make(chan vnc.ServerMessage)
	cchClient := make(chan vnc.ClientMessage)
	errorCh := make(chan error)

	ccfg := &vnc.ClientConfig{
		SecurityHandlers: []vnc.SecurityHandler{&vnc.ClientAuthATEN{Username: []byte(os.Args[2]), Password: []byte(os.Args[3])}},
		PixelFormat:      vnc.PixelFormat32bit,
		ClientMessageCh:  cchClient,
		ServerMessageCh:  cchServer,
		ServerMessages:   vnc.DefaultServerMessages,
		Encodings:        []vnc.Encoding{&vnc.RawEncoding{}},
		ErrorCh:          errorCh,
	}

	cc, err := vnc.Connect(context.Background(), nc, ccfg)

	if err != nil {
		log.Fatalf("Error negotiating connection to VNC host. %v", err)
	}
	// Process messages coming in on the ServerMessage channel.
	for {
		select {
		case err := <-errorCh:
			panic(err)
		case msg := <-cchClient:
			log.Printf("Received message type:%v msg:%v\n", msg.Type(), msg)
		case msg := <-cchServer:
			log.Printf("Received message type:%v msg:%v\n", msg.Type(), msg)
		}
	}
	cc.Wait()
}
