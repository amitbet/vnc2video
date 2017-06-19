package main

import (
	"context"
	"log"
	"net"

	vnc "github.com/vtolstov/go-vnc"
)

func main() {
	// Establish TCP connection to VNC server.
	nc, err := net.Dial("tcp", "192.168.100.41:5900")
	if err != nil {
		log.Fatalf("Error connecting to VNC host. %v", err)
	}

	// Negotiate connection with the server.
	cchServer := make(chan vnc.ServerMessage)
	cchClient := make(chan vnc.ClientMessage)

	ccfg := &vnc.ClientConfig{
		VersionHandler:    vnc.ClientVersionHandler,
		SecurityHandler:   vnc.ClientSecurityHandler,
		SecurityHandlers:  []vnc.SecurityHandler{&vnc.ClientAuthATEN{Username: []byte("ADMIN"), Password: []byte("ADMIN")}},
		ClientInitHandler: vnc.ClientClientInitHandler,
		ServerInitHandler: vnc.ClientServerInitHandler,
		PixelFormat:       vnc.PixelFormat32bit,
		ClientMessageCh:   cchClient,
		ServerMessageCh:   cchServer,
		ServerMessages:    vnc.DefaultServerMessages,
		Encodings:         []vnc.Encoding{&vnc.RawEncoding{}},
	}

	cc, err := vnc.Connect(context.Background(), nc, ccfg)

	if err != nil {
		log.Fatalf("Error negotiating connection to VNC host. %v", err)
	}

	// Listen and handle server messages.
	go cc.Handle()

	// Process messages coming in on the ServerMessage channel.
	for {
		select {
		case msg := <-cchClient:
			log.Printf("Received message type:%v msg:%v\n", msg.Type(), msg)
		case msg := <-cchServer:
			log.Printf("Received message type:%v msg:%v\n", msg.Type(), msg)
		}
	}
}
