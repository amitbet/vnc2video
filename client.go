package vnc

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

var (
	DefaultClientHandlers []ClientHandler = []ClientHandler{
		&DefaultClientVersionHandler{},
		&DefaultClientSecurityHandler{},
		&DefaultClientClientInitHandler{},
		&DefaultClientServerInitHandler{},
		&DefaultClientMessageHandler{},
	}
)

func Connect(ctx context.Context, c net.Conn, cfg *ClientConfig) (*ClientConn, error) {
	conn, err := NewClientConn(c, cfg)
	if err != nil {
		conn.Close()
		cfg.ErrorCh <- err
		return nil, err
	}

	if len(cfg.Handlers) == 0 {
		cfg.Handlers = DefaultClientHandlers
	}

	for _, h := range cfg.Handlers {
		if err := h.Handle(conn); err != nil {
			conn.Close()
			cfg.ErrorCh <- err
			return nil, err
		}
	}

	return conn, nil
}

var _ Conn = (*ClientConn)(nil)

func (c *ClientConn) Config() interface{} {
	return c.cfg
}

func (c *ClientConn) Wait() {
	<-c.quit
}

func (c *ClientConn) Conn() net.Conn {
	return c.c
}

func (c *ClientConn) SetProtoVersion(pv string) {
	c.protocol = pv
}

func (c *ClientConn) SetEncodings(encs []EncodingType) error {

	msg := &SetEncodings{
		EncNum:    uint16(len(encs)),
		Encodings: encs,
	}

	return msg.Write(c)
}

func (c *ClientConn) Flush() error {
	return c.bw.Flush()
}

func (c *ClientConn) Close() error {
	if c.quit != nil {
		close(c.quit)
		c.quit = nil
	}
	if c.quitCh != nil {
		close(c.quitCh)
	}
	return c.c.Close()
}

func (c *ClientConn) Read(buf []byte) (int, error) {
	return c.br.Read(buf)
}

func (c *ClientConn) Write(buf []byte) (int, error) {
	return c.bw.Write(buf)
}

func (c *ClientConn) ColorMap() ColorMap {
	return c.colorMap
}

func (c *ClientConn) SetColorMap(cm ColorMap) {
	c.colorMap = cm
}
func (c *ClientConn) DesktopName() []byte {
	return c.desktopName
}
func (c *ClientConn) PixelFormat() PixelFormat {
	return c.pixelFormat
}
func (c *ClientConn) SetDesktopName(name []byte) {
	copy(c.desktopName, name)
}
func (c *ClientConn) SetPixelFormat(pf PixelFormat) error {
	c.pixelFormat = pf
	return nil
}
func (c *ClientConn) Encodings() []Encoding {
	return c.encodings
}
func (c *ClientConn) Width() uint16 {
	return c.fbWidth
}
func (c *ClientConn) Height() uint16 {
	return c.fbHeight
}
func (c *ClientConn) Protocol() string {
	return c.protocol
}
func (c *ClientConn) SetWidth(w uint16) {
	c.fbWidth = w
}
func (c *ClientConn) SetHeight(h uint16) {
	c.fbHeight = h
}

// The ClientConn type holds client connection information.
type ClientConn struct {
	c        net.Conn
	br       *bufio.Reader
	bw       *bufio.Writer
	cfg      *ClientConfig
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

	// Height of the frame buffer in pixels, sent from the server.
	fbHeight uint16

	// Width of the frame buffer in pixels, sent from the server.
	fbWidth uint16

	// The pixel format associated with the connection. This shouldn't
	// be modified. If you wish to set a new pixel format, use the
	// SetPixelFormat method.
	pixelFormat PixelFormat

	quitCh  chan struct{}
	quit    chan struct{}
	errorCh chan error
}

func NewClientConn(c net.Conn, cfg *ClientConfig) (*ClientConn, error) {
	if len(cfg.Encodings) == 0 {
		return nil, fmt.Errorf("client can't handle encodings")
	}
	return &ClientConn{
		c:           c,
		cfg:         cfg,
		br:          bufio.NewReader(c),
		bw:          bufio.NewWriter(c),
		encodings:   cfg.Encodings,
		quitCh:      cfg.QuitCh,
		errorCh:     cfg.ErrorCh,
		pixelFormat: cfg.PixelFormat,
		quit:        make(chan struct{}),
	}, nil
}

type DefaultClientMessageHandler struct{}

//  listens to a VNC server and handles server messages.
func (*DefaultClientMessageHandler) Handle(c Conn) error {
	cfg := c.Config().(*ClientConfig)
	var err error
	var wg sync.WaitGroup
	wg.Add(2)
	defer c.Close()

	serverMessages := make(map[ServerMessageType]ServerMessage)
	for _, m := range cfg.ServerMessages {
		serverMessages[m.Type()] = m
	}

	go func() {
		defer wg.Done()
		for {
			select {
			case msg := <-cfg.ClientMessageCh:
				if err = msg.Write(c); err != nil {
					cfg.ErrorCh <- err
					return
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		for {
			select {
			default:
				var messageType ServerMessageType
				if err = binary.Read(c, binary.BigEndian, &messageType); err != nil {
					cfg.ErrorCh <- err
					return
				}
				msg, ok := serverMessages[messageType]
				if !ok {
					err = fmt.Errorf("unknown message-type: %v", messageType)
					cfg.ErrorCh <- err
					return
				}
				parsedMsg, err := msg.Read(c)
				if err != nil {
					cfg.ErrorCh <- err
					return
				}
				cfg.ServerMessageCh <- parsedMsg
			}
		}
	}()

	wg.Wait()
	return nil
}

// A ClientConfig structure is used to configure a ClientConn. After
// one has been passed to initialize a connection, it must not be modified.
type ClientConfig struct {
	Handlers         []ClientHandler
	SecurityHandlers []SecurityHandler
	Encodings        []Encoding
	PixelFormat      PixelFormat
	ColorMap         ColorMap
	ClientMessageCh  chan ClientMessage
	ServerMessageCh  chan ServerMessage
	Exclusive        bool
	ServerMessages   []ServerMessage
	QuitCh           chan struct{}
	ErrorCh          chan error
	quit             chan struct{}
}
