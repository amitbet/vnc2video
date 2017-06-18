package vnc

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

var DefaultClientMessages = []ClientMessage{
	&SetPixelFormat{},
	&SetEncodings{},
	&FramebufferUpdateRequest{},
	&KeyEvent{},
	&PointerEvent{},
	&ClientCutText{},
}

type ServerInit struct {
	FBWidth, FBHeight uint16
	PixelFormat       PixelFormat
	NameLength        uint32
	NameText          []byte
}

var _ Conn = (*ServerConn)(nil)

func (c *ServerConn) UnreadByte() error {
	return c.br.UnreadByte()
}

func (c *ServerConn) Conn() net.Conn {
	return c.c
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
	c.m.Lock()
	defer c.m.Unlock()
	return c.bw.Flush()
}

func (c *ServerConn) Close() error {
	return c.c.Close()
}

/*
func (c *ServerConn) Input() chan *ServerMessage {
	return c.cfg.ServerMessageCh
}

func (c *ServerConn) Output() chan *ClientMessage {
	return c.cfg.ClientMessageCh
}
*/
func (c *ServerConn) Read(buf []byte) (int, error) {
	return c.br.Read(buf)
}

func (c *ServerConn) Write(buf []byte) (int, error) {
	c.m.Lock()
	defer c.m.Unlock()
	return c.bw.Write(buf)
}

func (c *ServerConn) ColorMap() *ColorMap {
	return c.colorMap
}

func (c *ServerConn) SetColorMap(cm *ColorMap) {
	c.colorMap = cm
}
func (c *ServerConn) DesktopName() string {
	return c.desktopName
}
func (c *ServerConn) PixelFormat() *PixelFormat {
	return c.pixelFormat
}
func (c *ServerConn) SetDesktopName(name string) {
	c.desktopName = name
}
func (c *ServerConn) SetPixelFormat(pf *PixelFormat) error {
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

// ServerMessage represents a Client-to-Server RFB message type.
type ServerMessageType uint8

//go:generate stringer -type=ServerMessageType

// Client-to-Server message types.
const (
	FramebufferUpdateMsgType ServerMessageType = iota
	SetColorMapEntriesMsgType
	BellMsgType
	ServerCutTextMsgType
)

// FramebufferUpdate holds a FramebufferUpdate wire format message.
type FramebufferUpdate struct {
	_       [1]byte      // pad
	NumRect uint16       // number-of-rectangles
	Rects   []*Rectangle // rectangles
}

func (*FramebufferUpdate) Type() ServerMessageType {
	return FramebufferUpdateMsgType
}

func (*FramebufferUpdate) Read(c Conn) (ServerMessage, error) {
	msg := FramebufferUpdate{}
	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}

	if err := binary.Read(c, binary.BigEndian, &msg.NumRect); err != nil {
		return nil, err
	}
	for i := uint16(0); i < msg.NumRect; i++ {
		rect := &Rectangle{}
		if err := rect.Read(c); err != nil {
			return nil, err
		}
		msg.Rects = append(msg.Rects, rect)
	}
	return &msg, nil
}

func (msg *FramebufferUpdate) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	var pad [1]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, msg.NumRect); err != nil {
		return err
	}
	for _, rect := range msg.Rects {
		if err := rect.Write(c); err != nil {
			return err
		}
	}
	return c.Flush()
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
	colorMap *ColorMap

	// Name associated with the desktop, sent from the server.
	desktopName string

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
	pixelFormat *PixelFormat

	quit chan struct{}
}

type ServerHandler func(*ServerConfig, Conn) error

type ServerConfig struct {
	VersionHandler    ServerHandler
	SecurityHandler   ServerHandler
	SecurityHandlers  []SecurityHandler
	ClientInitHandler ServerHandler
	ServerInitHandler ServerHandler
	Encodings         []Encoding
	PixelFormat       *PixelFormat
	ColorMap          *ColorMap
	ClientMessageCh   chan ClientMessage
	ServerMessageCh   chan ServerMessage
	ClientMessages    []ClientMessage
	DesktopName       []byte
	Height            uint16
	Width             uint16
}

func NewServerConn(c net.Conn, cfg *ServerConfig) (*ServerConn, error) {
	if cfg.ClientMessageCh == nil {
		return nil, fmt.Errorf("ClientMessageCh nil")
	}

	if len(cfg.ClientMessages) == 0 {
		return nil, fmt.Errorf("ClientMessage 0")
	}

	return &ServerConn{
		c:           c,
		br:          bufio.NewReader(c),
		bw:          bufio.NewWriter(c),
		cfg:         cfg,
		quit:        make(chan struct{}),
		encodings:   cfg.Encodings,
		pixelFormat: cfg.PixelFormat,
		fbWidth:     cfg.Width,
		fbHeight:    cfg.Height,
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
			continue
		}

		if err := cfg.VersionHandler(cfg, conn); err != nil {
			conn.Close()
			continue
		}

		if err := cfg.SecurityHandler(cfg, conn); err != nil {
			conn.Close()
			continue
		}

		if err := cfg.ClientInitHandler(cfg, conn); err != nil {
			conn.Close()
			continue
		}

		if err := cfg.ServerInitHandler(cfg, conn); err != nil {
			conn.Close()
			continue
		}

		go conn.Handle()
	}
}

func (c *ServerConn) Handle() error {
	var err error
	var wg sync.WaitGroup

	defer c.Close()
	clientMessages := make(map[ClientMessageType]ClientMessage)
	for _, m := range c.cfg.ClientMessages {
		clientMessages[m.Type()] = m
	}
	wg.Add(2)

	// server
	go func() error {
		defer wg.Done()
		for {
			select {
			case msg := <-c.cfg.ServerMessageCh:
				if err = msg.Write(c); err != nil {
					return err
				}
			case <-c.quit:
				return nil
			}
		}
	}()

	// client
	go func() error {
		defer wg.Done()
		for {
			select {
			case <-c.quit:
				return nil
			default:
				var messageType ClientMessageType
				if err := binary.Read(c, binary.BigEndian, &messageType); err != nil {
					return err
				}
				msg, ok := clientMessages[messageType]
				if !ok {
					return fmt.Errorf("unsupported message-type: %v", messageType)

				}
				parsedMsg, err := msg.Read(c)
				if err != nil {
					fmt.Printf("srv err %s\n", err.Error())
					return err
				}
				c.cfg.ClientMessageCh <- parsedMsg
			}
		}
	}()

	wg.Wait()
	return nil
}

type ServerCutText struct {
	_      [1]byte
	Length uint32
	Text   []byte
}

func (*ServerCutText) Type() ServerMessageType {
	return ServerCutTextMsgType
}

func (*ServerCutText) Read(c Conn) (ServerMessage, error) {
	msg := ServerCutText{}

	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}

	if err := binary.Read(c, binary.BigEndian, &msg.Length); err != nil {
		return nil, err
	}

	msg.Text = make([]byte, msg.Length)
	if err := binary.Read(c, binary.BigEndian, &msg.Text); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *ServerCutText) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	var pad [1]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}

	if msg.Length < uint32(len(msg.Text)) {
		msg.Length = uint32(len(msg.Text))
	}
	if err := binary.Write(c, binary.BigEndian, msg.Length); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, msg.Text); err != nil {
		return err
	}
	return nil
}

type Bell struct{}

func (*Bell) Type() ServerMessageType {
	return BellMsgType
}

func (*Bell) Read(c Conn) (ServerMessage, error) {
	return &Bell{}, nil
}

func (msg *Bell) Write(c Conn) error {
	return binary.Write(c, binary.BigEndian, msg.Type())
}

type SetColorMapEntries struct {
	_          [1]byte
	FirstColor uint16
	ColorsNum  uint16
	Colors     []Color
}

func (*SetColorMapEntries) Type() ServerMessageType {
	return SetColorMapEntriesMsgType
}

func (*SetColorMapEntries) Read(c Conn) (ServerMessage, error) {
	msg := SetColorMapEntries{}
	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}

	if err := binary.Read(c, binary.BigEndian, &msg.FirstColor); err != nil {
		return nil, err
	}

	if err := binary.Read(c, binary.BigEndian, &msg.ColorsNum); err != nil {
		return nil, err
	}

	msg.Colors = make([]Color, msg.ColorsNum)
	colorMap := c.ColorMap()

	for i := uint16(0); i < msg.ColorsNum; i++ {
		color := &msg.Colors[i]
		if err := binary.Read(c, binary.BigEndian, &color); err != nil {
			return nil, err
		}
		colorMap[msg.FirstColor+i] = *color
	}
	c.SetColorMap(colorMap)
	return &msg, nil
}

func (msg *SetColorMapEntries) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	var pad [1]byte
	if err := binary.Write(c, binary.BigEndian, &pad); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, msg.FirstColor); err != nil {
		return err
	}

	if msg.ColorsNum < uint16(len(msg.Colors)) {
		msg.ColorsNum = uint16(len(msg.Colors))
	}
	if err := binary.Write(c, binary.BigEndian, msg.ColorsNum); err != nil {
		return err
	}

	for i := 0; i < len(msg.Colors); i++ {
		color := msg.Colors[i]
		if err := binary.Write(c, binary.BigEndian, color); err != nil {
			return err
		}
	}

	return nil
}
