package main

import (
	"context"
	"log"
	"net"

	vnc "github.com/vtolstov/go-vnc"
)

func main() {
	ln, err := net.Listen("tcp", ":5900")
	if err != nil {
		log.Fatalf("Error listen. %v", err)
	}

	chServer := make(chan vnc.ClientMessage)
	chClient := make(chan vnc.ServerMessage)

	cfg := &vnc.ServerConfig{
		VersionHandler:    vnc.ServerVersionHandler,
		SecurityHandler:   vnc.ServerSecurityHandler,
		SecurityHandlers:  []vnc.ServerHandler{vnc.ServerSecurityNoneHandler},
		ClientInitHandler: vnc.ServerClientInitHandler,
		ServerInitHandler: vnc.ServerServerInitHandler,
		Encodings:         []vnc.Encoding{&vnc.RawEncoding{}},
		PixelFormat:       vnc.PixelFormat32bit,
		ClientMessageCh:   chServer,
		ServerMessageCh:   chClient,
		ClientMessages:    vnc.DefaultClientMessages,
	}
	go vnc.Serve(context.Background(), ln, cfg)

	// Process messages coming in on the ClientMessage channel.
	for {
		msg := <-chClient
		switch msg.Type() {
		default:
			log.Printf("Received message type:%v msg:%v\n", msg.Type(), msg)
		}
	}
}
