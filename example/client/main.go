package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"
	vnc "github.com/amitbet/vnc2video"
	"github.com/amitbet/vnc2video/encoders"
	"github.com/amitbet/vnc2video/logger"
)

func main() {
	runtime.GOMAXPROCS(4)
	framerate := 12
	runWithProfiler := false

	// Establish TCP connection to VNC server.
	nc, err := net.DialTimeout("tcp", os.Args[1], 5*time.Second)
	if err != nil {
		logger.Fatalf("Error connecting to VNC host. %v", err)
	}

	logger.Tracef("starting up the client, connecting to: %s", os.Args[1])
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
		DrawCursor:      true,
		PixelFormat:     vnc.PixelFormat32bit,
		ClientMessageCh: cchClient,
		ServerMessageCh: cchServer,
		Messages:        vnc.DefaultServerMessages,
		Encodings: []vnc.Encoding{
			&vnc.RawEncoding{},
			&vnc.TightEncoding{},
			&vnc.HextileEncoding{},
			&vnc.ZRLEEncoding{},
			&vnc.CopyRectEncoding{},
			&vnc.CursorPseudoEncoding{},
			&vnc.CursorPosPseudoEncoding{},
			&vnc.ZLibEncoding{},
			&vnc.RREEncoding{},
		},
		ErrorCh: errorCh,
	}

	cc, err := vnc.Connect(context.Background(), nc, ccfg)
	screenImage := cc.Canvas
	if err != nil {
		logger.Fatalf("Error negotiating connection to VNC host. %v", err)
	}
	// out, err := os.Create("./output" + strconv.Itoa(counter) + ".jpg")
	// if err != nil {
	// 	fmt.Println(err)p
	// 	os.Exit(1)
	// }
	//vcodec := &encoders.MJPegImageEncoder{Quality: 60 , Framerate: framerate}
	//vcodec := &encoders.X264ImageEncoder{FFMpegBinPath: "./ffmpeg", Framerate: framerate}
	//vcodec := &encoders.HuffYuvImageEncoder{FFMpegBinPath: "./ffmpeg", Framerate: framerate}
	vcodec := &encoders.QTRLEImageEncoder{FFMpegBinPath: "./ffmpeg", Framerate: framerate}
	//vcodec := &encoders.VP8ImageEncoder{FFMpegBinPath:"./ffmpeg", Framerate: framerate}
	//vcodec := &encoders.DV9ImageEncoder{FFMpegBinPath:"./ffmpeg", Framerate: framerate}

	//counter := 0
	//vcodec.Init("./output" + strconv.Itoa(counter))

	go vcodec.Run("./output.mp4")
	//windows
	///go vcodec.Run("/Users/amitbet/Dropbox/go/src/vnc2webm/example/file-reader/ffmpeg", "./output.mp4")

	//go vcodec.Run("C:\\Users\\betzalel\\Dropbox\\go\\src\\vnc2video\\example\\client\\ffmpeg.exe", "output.mp4")
	//vcodec.Run("./output")

	// var out *os.File

	logger.Tracef("connected to: %s", os.Args[1])
	defer cc.Close()

	cc.SetEncodings([]vnc.EncodingType{
		vnc.EncCursorPseudo,
		vnc.EncPointerPosPseudo,
		vnc.EncCopyRect,
		vnc.EncTight,
		vnc.EncZRLE,
		//vnc.EncHextile,
		//vnc.EncZlib,
		//vnc.EncRRE,
	})
	//rect := image.Rect(0, 0, int(cc.Width()), int(cc.Height()))
	//screenImage := image.NewRGBA64(rect)
	// Process messages coming in on the ServerMessage channel.

	go func() {
		for {
			timeStart := time.Now()

			vcodec.Encode(screenImage.Image)

			timeTarget := timeStart.Add((1000 / time.Duration(framerate)) * time.Millisecond)
			timeLeft := timeTarget.Sub(time.Now())
			if timeLeft > 0 {
				time.Sleep(timeLeft)
			}
		}
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	frameBufferReq := 0
	timeStart := time.Now()

	if runWithProfiler {
		profFile := "prof.file"
		f, err := os.Create(profFile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	for {
		select {
		case err := <-errorCh:
			panic(err)
		case msg := <-cchClient:
			logger.Tracef("Received client message type:%v msg:%v\n", msg.Type(), msg)
		case msg := <-cchServer:
			//logger.Tracef("Received server message type:%v msg:%v\n", msg.Type(), msg)

			// out, err := os.Create("./output" + strconv.Itoa(counter) + ".jpg")
			// if err != nil {
			// 	fmt.Println(err)
			// 	os.Exit(1)
			// }

			if msg.Type() == vnc.FramebufferUpdateMsgType {
				secsPassed := time.Now().Sub(timeStart).Seconds()
				frameBufferReq++
				reqPerSec := float64(frameBufferReq) / secsPassed
				//counter++
				//jpeg.Encode(out, screenImage, nil)
				///vcodec.Encode(screenImage)
				logger.Infof("reqs=%d, seconds=%f, Req Per second= %f", frameBufferReq, secsPassed, reqPerSec)

				reqMsg := vnc.FramebufferUpdateRequest{Inc: 1, X: 0, Y: 0, Width: cc.Width(), Height: cc.Height()}
				//cc.ResetAllEncodings()
				reqMsg.Write(cc)
			}
		case signal := <-sigc:
			if signal != nil {
				vcodec.Close()
				pprof.StopCPUProfile()
				time.Sleep(2 * time.Second)
				os.Exit(1)
			}
		}
	}
	//cc.Wait()
}
