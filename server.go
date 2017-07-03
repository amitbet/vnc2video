package vnc

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

var _ Conn = (*ServerConn)(nil)

func (c *ServerConn) Config() interface{} {
	return c.cfg
}

func (c *ServerConn) Conn() net.Conn {
	return c.c
}

func (c *ServerConn) Wait() {
	<-c.quit
}

func (c *ServerConn) SetEncodings(encs []EncodingType) error {
	encodings := make(map[EncodingType]Encoding)
	for _, enc := range c.cfg.Encodings {
		encodings[enc.Type()] = enc
	}
	for _, encType := range encs {
		if enc, ok := encodings[encType]; ok {
			c.encodings = append(c.encodings, enc)
		}
	}
	return nil
}

func (c *ServerConn) SetProtoVersion(pv string) {
	c.protocol = pv
}

func (c *ServerConn) Flush() error {
	return c.bw.Flush()
}

func (c *ServerConn) Close() error {
	if c.quit != nil {
		close(c.quit)
		c.quit = nil
	}
	return c.c.Close()
}

func (c *ServerConn) Read(buf []byte) (int, error) {
	return c.br.Read(buf)
}

func (c *ServerConn) Write(buf []byte) (int, error) {
	return c.bw.Write(buf)
}

func (c *ServerConn) ColorMap() ColorMap {
	return c.colorMap
}

func (c *ServerConn) SetColorMap(cm ColorMap) {
	c.colorMap = cm
}
func (c *ServerConn) DesktopName() []byte {
	return c.desktopName
}
func (c *ServerConn) PixelFormat() PixelFormat {
	return c.pixelFormat
}
func (c *ServerConn) SetDesktopName(name []byte) {
	copy(c.desktopName, name)
}
func (c *ServerConn) SetPixelFormat(pf PixelFormat) error {
	c.pixelFormat = pf
	return nil
}
func (c *ServerConn) Encodings() []Encoding {
	return c.encodings
}
func (c *ServerConn) Width() uint16 {
	return c.fbWidth
}
func (c *ServerConn) Height() uint16 {
	return c.fbHeight
}
func (c *ServerConn) Protocol() string {
	return c.protocol
}

// TODO send desktopsize pseudo encoding
func (c *ServerConn) SetWidth(w uint16) {
	c.fbWidth = w
}
func (c *ServerConn) SetHeight(h uint16) {
	c.fbHeight = h
}

type ServerConn struct {
	c        net.Conn
	cfg      *ServerConfig
	br       *bufio.Reader
	bw       *bufio.Writer
	protocol string
	m        sync.Mutex
	// If the pixel format uses a color map, then this is the color
	// map that is used. This should not be modified directly, since
	// the data comes from the server.
	// Definition in §5 - Representation of Pixel Data.
	colorMap ColorMap

	// Name associated with the desktop, sent from the server.
	desktopName []byte

	// Encodings supported by the client. This should not be modified
	// directly. Instead, SetEncodings() should be used.
	encodings []Encoding

	// Height of the frame buffer in pixels, sent to the client.
	fbHeight uint16

	// Width of the frame buffer in pixels, sent to the client.
	fbWidth uint16

	// The pixel format associated with the connection. This shouldn't
	// be modified. If you wish to set a new pixel format, use the
	// SetPixelFormat method.
	pixelFormat PixelFormat

	quit chan struct{}
}

type ServerHandler interface {
	Handle(Conn) error
}

var (
	DefaultServerHandlers []ServerHandler = []ServerHandler{
		&DefaultServerVersionHandler{},
		&DefaultServerSecurityHandler{},
		&DefaultServerClientInitHandler{},
		&DefaultServerServerInitHandler{},
		&DefaultServerMessageHandler{},
	}
)

type ServerConfig struct {
	Handlers         []ServerHandler
	SecurityHandlers []SecurityHandler
	Encodings        []Encoding
	PixelFormat      PixelFormat
	ColorMap         ColorMap
	ClientMessageCh  chan ClientMessage
	ServerMessageCh  chan ServerMessage
	ClientMessages   []ClientMessage
	DesktopName      []byte
	Height           uint16
	Width            uint16
	ErrorCh          chan error
}

func NewServerConn(c net.Conn, cfg *ServerConfig) (*ServerConn, error) {
	return &ServerConn{
		c:           c,
		br:          bufio.NewReader(c),
		bw:          bufio.NewWriter(c),
		cfg:         cfg,
		desktopName: cfg.DesktopName,
		encodings:   cfg.Encodings,
		pixelFormat: cfg.PixelFormat,
		fbWidth:     cfg.Width,
		fbHeight:    cfg.Height,
		quit:        make(chan struct{}),
	}, nil
}

func Serve(ctx context.Context, ln net.Listener, cfg *ServerConfig) error {
	for {

		c, err := ln.Accept()
		if err != nil {
			continue
		}

		conn, err := NewServerConn(c, cfg)
		if err != nil {
			cfg.ErrorCh <- err
			continue
		}

		if len(cfg.Handlers) == 0 {
			cfg.Handlers = DefaultServerHandlers
		}

	handlerLoop:
		for _, h := range cfg.Handlers {
			if err := h.Handle(conn); err != nil {
				cfg.ErrorCh <- err
				conn.Close()
				break handlerLoop
			}
		}
	}
}

type DefaultServerMessageHandler struct{}

func (*DefaultServerMessageHandler) Handle(c Conn) error {
	cfg := c.Config().(*ServerConfig)
	var err error
	var wg sync.WaitGroup

	defer c.Close()
	clientMessages := make(map[ClientMessageType]ClientMessage)
	for _, m := range cfg.ClientMessages {
		clientMessages[m.Type()] = m
	}
	wg.Add(2)

	quit := make(chan struct{})

	// server
	go func() {
		defer wg.Done()
		for {
			select {
			case <-quit:
				return
			case msg := <-cfg.ServerMessageCh:
				if err = msg.Write(c); err != nil {
					cfg.ErrorCh <- err
					if quit != nil {
						close(quit)
						quit = nil
					}
					return
				}
			}
		}
	}()

	// client
	go func() {
		defer wg.Done()
		for {
			select {
			case <-quit:
				return
			default:
				var messageType ClientMessageType
				if err := binary.Read(c, binary.BigEndian, &messageType); err != nil {
					cfg.ErrorCh <- err
					if quit != nil {
						close(quit)
						quit = nil
					}
					return
				}
				msg, ok := clientMessages[messageType]
				if !ok {
					cfg.ErrorCh <- fmt.Errorf("unsupported message-type: %v", messageType)
					close(quit)
					return
				}
				parsedMsg, err := msg.Read(c)
				if err != nil {
					cfg.ErrorCh <- err
					if quit != nil {
						close(quit)
						quit = nil
					}
					return
				}
				cfg.ClientMessageCh <- parsedMsg
			}
		}
	}()

	wg.Wait()
	return nil
}
