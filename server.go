package vnc

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
)

var DefaultClientMessages = []ClientMessage{
	&SetPixelFormat{},
	&SetEncodings{},
	&FramebufferUpdateRequest{},
	&KeyEvent{},
	&PointerEvent{},
	&ClientCutText{},
}

var _ Conn = (*ServerConn)(nil)

func (c *ServerConn) Flush() error {
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
	MsgType ServerMessageType
	NumRect uint16      // number-of-rectangles
	_       [1]byte     // pad
	Rects   []Rectangle // rectangles
}

func (msg *FramebufferUpdate) Type() ServerMessageType {
	return msg.MsgType
}

func (msg *FramebufferUpdate) Read(c Conn) error {
	if err := binary.Read(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}

	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return err
	}

	if err := binary.Read(c, binary.BigEndian, msg.NumRect); err != nil {
		return err
	}
	/*
		// Extract rectangles.
		rects := make([]Rectangle, msg.NumRect)
		for i := uint16(0); i < msg.NumRect; i++ {
			rect := NewRectangle(c)
			if err := rect.Read(c); err != nil {
				return err
			}
			msg.Rects = append(msg.Rects, *rect)
		}
	*/
	return nil
}

func (msg *FramebufferUpdate) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}
	var pad [1]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, msg.NumRect); err != nil {
		return err
	}
	/*
		for _, rect := range msg.Rects {
			if err := rect.Write(c); err != nil {
				return err
			}
		}
	*/
	return c.Flush()
}

type ServerConn struct {
	c        net.Conn
	cfg      *ServerConfig
	br       *bufio.Reader
	bw       *bufio.Writer
	protocol string

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
	SecurityHandlers  []ServerHandler
	ClientInitHandler ServerHandler
	ServerInitHandler ServerHandler
	Encodings         []Encoding
	PixelFormat       *PixelFormat
	ColorMap          *ColorMap
	ClientMessageCh   chan ClientMessage
	ServerMessageCh   chan ServerMessage
	ClientMessages    []ClientMessage
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
		encodings:   cfg.Encodings,
		pixelFormat: cfg.PixelFormat,
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
	defer c.Close()
	clientMessages := make(map[ClientMessageType]ClientMessage)
	for _, m := range c.cfg.ClientMessages {
		clientMessages[m.Type()] = m
	}

serverLoop:
	for {
		select {
		case msg := <-c.cfg.ServerMessageCh:
			if err = msg.Write(c); err != nil {
				return err
			}
			c.Flush()
		case <-c.quit:
			break serverLoop
		}
	}

clientLoop:
	for {
		select {
		case <-c.quit:
			break clientLoop
		default:
			var messageType ClientMessageType
			if err := binary.Read(c, binary.BigEndian, &messageType); err != nil {
				break clientLoop
			}

			msg, ok := clientMessages[messageType]
			if !ok {
				err = fmt.Errorf("unsupported message-type: %v", messageType)
				break clientLoop
			}
			if err := msg.Read(c); err != nil {
				break clientLoop
			}

			c.cfg.ClientMessageCh <- msg
		}
	}
	return nil
}

type ServerCutText struct {
	MsgType ServerMessageType
	_       [1]byte
	Length  uint32
	Text    []byte
}

func (msg *ServerCutText) Type() ServerMessageType {
	return msg.MsgType
}

func (msg *ServerCutText) Read(c Conn) error {
	if err := binary.Read(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}
	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return err
	}

	var length uint32
	if err := binary.Read(c, binary.BigEndian, &length); err != nil {
		return err
	}

	text := make([]byte, length)
	if err := binary.Read(c, binary.BigEndian, &text); err != nil {
		return err
	}
	msg.Text = text
	return nil
}

func (msg *ServerCutText) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}
	var pad [1]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, msg.Length); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, msg.Text); err != nil {
		return err
	}
	return nil
}

type Bell struct {
	MsgType ServerMessageType
}

func (msg *Bell) Type() ServerMessageType {
	return msg.MsgType
}

func (msg *Bell) Read(c Conn) error {
	return binary.Read(c, binary.BigEndian, msg.MsgType)
}

func (msg *Bell) Write(c Conn) error {
	return binary.Write(c, binary.BigEndian, msg.MsgType)
}

type SetColorMapEntries struct {
	MsgType    ServerMessageType
	_          [1]byte
	FirstColor uint16
	ColorsNum  uint16
	Colors     []Color
}

func (msg *SetColorMapEntries) Type() ServerMessageType {
	return msg.MsgType
}

func (msg *SetColorMapEntries) Read(c Conn) error {
	if err := binary.Read(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}
	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return err
	}

	if err := binary.Read(c, binary.BigEndian, msg.FirstColor); err != nil {
		return err
	}

	if err := binary.Read(c, binary.BigEndian, msg.ColorsNum); err != nil {
		return err
	}

	msg.Colors = make([]Color, msg.ColorsNum)
	colorMap := c.ColorMap()

	for i := uint16(0); i < msg.ColorsNum; i++ {
		color := &msg.Colors[i]
		if err := binary.Read(c, binary.BigEndian, &color); err != nil {
			return err
		}
		colorMap[msg.FirstColor+i] = *color
	}
	c.SetColorMap(colorMap)
	return nil
}

func (msg *SetColorMapEntries) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.MsgType); err != nil {
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
