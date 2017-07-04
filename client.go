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
	// DefaultClientHandlers represents default client handlers
	DefaultClientHandlers = []Handler{
		&DefaultClientVersionHandler{},
		&DefaultClientSecurityHandler{},
		&DefaultClientClientInitHandler{},
		&DefaultClientServerInitHandler{},
		&DefaultClientMessageHandler{},
	}
)

// Connect handshake with remote server using underlining net.Conn
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

// Config returns connection config
func (c *ClientConn) Config() interface{} {
	return c.cfg
}

// Wait waiting for connection close
func (c *ClientConn) Wait() {
	<-c.quit
}

// Conn return underlining net.Conn
func (c *ClientConn) Conn() net.Conn {
	return c.c
}

// SetProtoVersion sets proto version
func (c *ClientConn) SetProtoVersion(pv string) {
	c.protocol = pv
}

// SetEncodings write SetEncodings message
func (c *ClientConn) SetEncodings(encs []EncodingType) error {

	msg := &SetEncodings{
		EncNum:    uint16(len(encs)),
		Encodings: encs,
	}

	return msg.Write(c)
}

// Flush flushes data to conn
func (c *ClientConn) Flush() error {
	return c.bw.Flush()
}

// Close closing conn
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

// Read reads data from conn
func (c *ClientConn) Read(buf []byte) (int, error) {
	return c.br.Read(buf)
}

// Write data to conn must be Flushed
func (c *ClientConn) Write(buf []byte) (int, error) {
	return c.bw.Write(buf)
}

// ColorMap returns color map
func (c *ClientConn) ColorMap() ColorMap {
	return c.colorMap
}

// SetColorMap sets color map
func (c *ClientConn) SetColorMap(cm ColorMap) {
	c.colorMap = cm
}

// DesktopName returns connection desktop name
func (c *ClientConn) DesktopName() []byte {
	return c.desktopName
}

// PixelFormat returns connection pixel format
func (c *ClientConn) PixelFormat() PixelFormat {
	return c.pixelFormat
}

// SetDesktopName sets desktop name
func (c *ClientConn) SetDesktopName(name []byte) {
	c.desktopName = name
}

// SetPixelFormat sets pixel format
func (c *ClientConn) SetPixelFormat(pf PixelFormat) error {
	c.pixelFormat = pf
	return nil
}

// Encodings returns client encodings
func (c *ClientConn) Encodings() []Encoding {
	return c.encodings
}

// Width returns width
func (c *ClientConn) Width() uint16 {
	return c.fbWidth
}

// Height returns height
func (c *ClientConn) Height() uint16 {
	return c.fbHeight
}

// Protocol returns protocol
func (c *ClientConn) Protocol() string {
	return c.protocol
}

// SetWidth sets width of client conn
func (c *ClientConn) SetWidth(width uint16) {
	c.fbWidth = width
}

// SetHeight sets height of client conn
func (c *ClientConn) SetHeight(height uint16) {
	c.fbHeight = height
}

// The ClientConn type holds client connection information
type ClientConn struct {
	c        net.Conn
	br       *bufio.Reader
	bw       *bufio.Writer
	cfg      *ClientConfig
	protocol string

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

// NewClientConn creates new client conn using config
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

// DefaultClientMessageHandler represents default client message handler
type DefaultClientMessageHandler struct{}

// Handle handles server messages.
func (*DefaultClientMessageHandler) Handle(c Conn) error {
	cfg := c.Config().(*ClientConfig)
	var err error
	var wg sync.WaitGroup
	wg.Add(2)
	defer c.Close()

	serverMessages := make(map[ServerMessageType]ServerMessage)
	for _, m := range cfg.Messages {
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
	Handlers         []Handler
	SecurityHandlers []SecurityHandler
	Encodings        []Encoding
	PixelFormat      PixelFormat
	ColorMap         ColorMap
	ClientMessageCh  chan ClientMessage
	ServerMessageCh  chan ServerMessage
	Exclusive        bool
	Messages         []ServerMessage
	QuitCh           chan struct{}
	ErrorCh          chan error
	quit             chan struct{}
}
