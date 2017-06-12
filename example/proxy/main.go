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

	schServer := make(chan vnc.ClientMessage)
	schClient := make(chan vnc.ServerMessage)

	scfg := &vnc.ServerConfig{
		Width:             800,
		Height:            600,
		VersionHandler:    vnc.ServerVersionHandler,
		SecurityHandler:   vnc.ServerSecurityHandler,
		SecurityHandlers:  []vnc.SecurityHandler{&vnc.ClientAuthNone{}},
		ClientInitHandler: vnc.ServerClientInitHandler,
		ServerInitHandler: vnc.ServerServerInitHandler,
		Encodings:         []vnc.Encoding{&vnc.RawEncoding{}},
		PixelFormat:       vnc.PixelFormat24bit,
		ClientMessageCh:   schServer,
		ServerMessageCh:   schClient,
		ClientMessages:    vnc.DefaultClientMessages,
		DesktopName:       []byte("vnc proxy"),
	}
	go vnc.Serve(context.Background(), ln, scfg)

	c, err := net.Dial("tcp", "127.0.0.1:5944")
	if err != nil {
		log.Fatalf("Error dial. %v", err)
	}
	cchServer := make(chan vnc.ServerMessage)
	cchClient := make(chan vnc.ClientMessage)

	ccfg := &vnc.ClientConfig{
		VersionHandler:    vnc.ClientVersionHandler,
		SecurityHandler:   vnc.ClientSecurityHandler,
		SecurityHandlers:  []vnc.SecurityHandler{&vnc.ClientAuthNone{}},
		ClientInitHandler: vnc.ClientClientInitHandler,
		ServerInitHandler: vnc.ClientServerInitHandler,
		PixelFormat:       vnc.PixelFormat24bit,
		ClientMessageCh:   cchClient,
		ServerMessageCh:   cchServer,
		ServerMessages:    vnc.DefaultServerMessages,
		Encodings:         []vnc.Encoding{&vnc.RawEncoding{}},
	}

	cc, err := vnc.Connect(context.Background(), c, ccfg)
	if err != nil {
		log.Fatalf("Error dial. %v", err)
	}
	defer cc.Close()
	go cc.Handle()

	for {
		select {
		case msg := <-cchClient:
			switch msg.Type() {
			default:
				log.Printf("00 Received message type:%v msg:%v\n", msg.Type(), msg)
			}
		case msg := <-cchServer:
			switch msg.Type() {
			default:
				log.Printf("01 Received message type:%v msg:%v\n", msg.Type(), msg)
			}
		case msg := <-schClient:
			switch msg.Type() {
			default:
				log.Printf("10 Received message type:%v msg:%v\n", msg.Type(), msg)
			}
		case msg := <-schServer:
			log.Printf("11 Received message type:%v msg:%v\n", msg.Type(), msg)
			switch msg.Type() {
			case vnc.SetEncodingsMsgType:
				encRaw := &vnc.RawEncoding{}
				msg1 := &vnc.SetEncodings{
					MsgType:   vnc.SetEncodingsMsgType,
					EncNum:    1,
					Encodings: []vnc.EncodingType{encRaw.Type()},
				}
				if err := msg1.Write(cc); err != nil {
					log.Fatalf("err %v\n", err)
				}
				msg2 := &vnc.FramebufferUpdateRequest{
					MsgType: vnc.FramebufferUpdateRequestMsgType,
					Inc:     0,
					X:       0,
					Y:       0,
					Width:   cc.Width(),
					Height:  cc.Height(),
				}
				if err := msg2.Write(cc); err != nil {
					log.Fatalf("err %v\n", err)
				}
			default:
				if err := msg.Write(cc); err != nil {
					log.Fatalf("err %v\n", err)
				}
			}
		}
	}
}
