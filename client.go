package vnc

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
)

var DefaultServerMessages = []ServerMessage{
	&FramebufferUpdate{},
	&SetColorMapEntries{},
	&Bell{},
	&ServerCutText{},
}

func Connect(ctx context.Context, c net.Conn, cfg *ClientConfig) (*ClientConn, error) {
	conn, err := NewClientConn(c, cfg)
	if err != nil {
		conn.Close()
		return nil, err
	}
	if err := cfg.VersionHandler(cfg, conn); err != nil {
		conn.Close()
		return nil, err
	}
	if err := cfg.SecurityHandler(cfg, conn); err != nil {
		conn.Close()
		return nil, err
	}
	if err := cfg.ClientInitHandler(cfg, conn); err != nil {
		conn.Close()
		return nil, err
	}
	if err := cfg.ServerInitHandler(cfg, conn); err != nil {
		conn.Close()
		return nil, err
	}
	/*
	   // Send client-to-server messages.
	   encs := conn.encodings
	   if err := conn.SetEncodings(encs); err != nil {
	       conn.Close()
	       return nil, Errorf("failure calling SetEncodings; %s", err)
	   }
	   pf := conn.pixelFormat
	   if err := conn.SetPixelFormat(pf); err != nil {
	       conn.Close()
	       return nil, Errorf("failure calling SetPixelFormat; %s", err)
	   }
	*/
	return conn, nil
}

var _ Conn = (*ClientConn)(nil)

func (c *ClientConn) Flush() error {
	return c.bw.Flush()
}

func (c *ClientConn) Close() error {
	return c.c.Close()
}

func (c *ClientConn) Read(buf []byte) (int, error) {
	return c.br.Read(buf)
}

func (c *ClientConn) Write(buf []byte) (int, error) {
	return c.bw.Write(buf)
}

func (c *ClientConn) ColorMap() *ColorMap {
	return c.colorMap
}

func (c *ClientConn) SetColorMap(cm *ColorMap) {
	c.colorMap = cm
}
func (c *ClientConn) DesktopName() string {
	return c.desktopName
}
func (c *ClientConn) PixelFormat() *PixelFormat {
	return c.pixelFormat
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

	// Height of the frame buffer in pixels, sent from the server.
	fbHeight uint16

	// Width of the frame buffer in pixels, sent from the server.
	fbWidth uint16

	// The pixel format associated with the connection. This shouldn't
	// be modified. If you wish to set a new pixel format, use the
	// SetPixelFormat method.
	pixelFormat *PixelFormat

	quit chan struct{}
}

func NewClientConn(c net.Conn, cfg *ClientConfig) (*ClientConn, error) {
	if cfg.ServerMessages == nil {
		return nil, fmt.Errorf("ServerMessages cannel is nil")
	}
	if len(cfg.Encodings) == 0 {
		return nil, fmt.Errorf("client can't handle encodings")
	}
	return &ClientConn{
		c:           c,
		cfg:         cfg,
		br:          bufio.NewReader(c),
		bw:          bufio.NewWriter(c),
		encodings:   cfg.Encodings,
		pixelFormat: cfg.PixelFormat,
	}, nil
}

// ClientMessage represents a Client-to-Server RFB message type.
type ClientMessageType uint8

//go:generate stringer -type=ClientMessageType

// Client-to-Server message types.
const (
	SetPixelFormatMsgType ClientMessageType = iota
	_
	SetEncodingsMsgType
	FramebufferUpdateRequestMsgType
	KeyEventMsgType
	PointerEventMsgType
	ClientCutTextMsgType
)

// SetPixelFormat holds the wire format message.
type SetPixelFormat struct {
	MsgType ClientMessageType
	_       [3]byte     // padding
	PF      PixelFormat // pixel-format
}

func (msg *SetPixelFormat) Type() ClientMessageType {
	return msg.MsgType
}

func (msg *SetPixelFormat) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, msg); err != nil {
		return err
	}

	pf := c.PixelFormat()
	// Invalidate the color map.
	if pf.TrueColor != 1 {
		c.SetColorMap(&ColorMap{})
	}

	return nil
}

func (msg *SetPixelFormat) Read(c Conn) error {
	return binary.Read(c, binary.BigEndian, msg)
}

// SetEncodings holds the wire format message, sans encoding-type field.
type SetEncodings struct {
	MsgType   ClientMessageType
	_         [1]byte // padding
	EncNum    uint16  // number-of-encodings
	Encodings []Encoding
}

func (msg *SetEncodings) Type() ClientMessageType {
	return msg.MsgType
}

func (msg *SetEncodings) Read(c Conn) error {
	if err := binary.Read(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}

	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return err
	}

	if err := binary.Read(c, binary.BigEndian, msg.EncNum); err != nil {
		return err
	}

	var enc Encoding
	for i := uint16(0); i < msg.EncNum; i++ {
		if err := binary.Read(c, binary.BigEndian, &enc); err != nil {
			return err
		}
		msg.Encodings = append(msg.Encodings, enc)
	}
	return nil
}

func (msg *SetEncodings) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}

	var pad [1]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}

	if uint16(len(msg.Encodings)) > msg.EncNum {
		msg.EncNum = uint16(len(msg.Encodings))
	}

	if err := binary.Write(c, binary.BigEndian, msg.EncNum); err != nil {
		return err
	}

	for _, enc := range msg.Encodings {
		if err := binary.Write(c, binary.BigEndian, enc); err != nil {
			return err
		}
	}
	return c.Flush()
}

// FramebufferUpdateRequest holds the wire format message.
type FramebufferUpdateRequest struct {
	MsgType       ClientMessageType
	Inc           uint8  // incremental
	X, Y          uint16 // x-, y-position
	Width, Height uint16 // width, height
}

func (msg *FramebufferUpdateRequest) Type() ClientMessageType {
	return msg.MsgType
}

func (msg *FramebufferUpdateRequest) Read(c Conn) error {
	return binary.Read(c, binary.BigEndian, msg)
}

func (msg *FramebufferUpdateRequest) Write(c Conn) error {
	return binary.Write(c, binary.BigEndian, msg)
}

// KeyEvent holds the wire format message.
type KeyEvent struct {
	MsgType ClientMessageType // message-type
	Down    uint8             // down-flag
	_       [2]byte           // padding
	Key     Key               // key
}

func (msg *KeyEvent) Type() ClientMessageType {
	return msg.MsgType
}

func (msg *KeyEvent) Read(c Conn) error {
	return binary.Read(c, binary.BigEndian, msg)
}

func (msg *KeyEvent) Write(c Conn) error {
	return binary.Write(c, binary.BigEndian, msg)
}

// PointerEventMessage holds the wire format message.
type PointerEvent struct {
	MsgType ClientMessageType // message-type
	Mask    uint8             // button-mask
	X, Y    uint16            // x-, y-position
}

func (msg *PointerEvent) Type() ClientMessageType {
	return msg.MsgType
}

func (msg *PointerEvent) Read(c Conn) error {
	return binary.Read(c, binary.BigEndian, msg)
}

func (msg *PointerEvent) Write(c Conn) error {
	return binary.Write(c, binary.BigEndian, msg)
}

// ClientCutText holds the wire format message, sans the text field.
type ClientCutText struct {
	MsgType ClientMessageType // message-type
	_       [3]byte           // padding
	Length  uint32            // length
	Text    []byte
}

func (msg *ClientCutText) Type() ClientMessageType {
	return msg.MsgType
}

func (msg *ClientCutText) Read(c Conn) error {
	if err := binary.Read(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}

	var pad [3]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return err
	}

	if err := binary.Read(c, binary.BigEndian, msg.Length); err != nil {
		return err
	}

	text := make([]uint8, msg.Length)
	if err := binary.Read(c, binary.BigEndian, &text); err != nil {
		return err
	}
	msg.Text = text
	return nil
}

func (msg *ClientCutText) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.MsgType); err != nil {
		return err
	}

	var pad [3]byte
	if err := binary.Write(c, binary.BigEndian, &pad); err != nil {
		return err
	}

	if uint32(len(msg.Text)) > msg.Length {
		msg.Length = uint32(len(msg.Text))
	}

	if err := binary.Write(c, binary.BigEndian, msg.Length); err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, msg.Text); err != nil {
		return err
	}

	return c.Flush()
}

// ListenAndHandle listens to a VNC server and handles server messages.
func (c *ClientConn) Handle() error {
	var err error

	defer c.Close()

	serverMessages := make(map[ServerMessageType]ServerMessage)
	for _, m := range c.cfg.ServerMessages {
		serverMessages[m.Type()] = m
	}

clientLoop:
	for {
		select {
		case msg := <-c.cfg.ServerMessageCh:
			if err = msg.Write(c); err != nil {
				return err
			}
		case <-c.quit:
			break clientLoop
		}
	}

serverLoop:
	for {
		select {
		case <-c.quit:
			break serverLoop
		default:
			var messageType ServerMessageType
			if err = binary.Read(c, binary.BigEndian, &messageType); err != nil {
				break serverLoop
			}
			msg, ok := serverMessages[messageType]
			if !ok {
				break serverLoop
			}
			if err = msg.Read(c); err != nil {
				break serverLoop
			}
			if c.cfg.ServerMessageCh == nil {
				continue
			}
			c.cfg.ServerMessageCh <- msg
		}
	}
	return err
}

type ClientHandler func(*ClientConfig, Conn) error

// A ClientConfig structure is used to configure a ClientConn. After
// one has been passed to initialize a connection, it must not be modified.
type ClientConfig struct {
	VersionHandler    ClientHandler
	SecurityHandler   ClientHandler
	SecurityHandlers  []ClientHandler
	ClientInitHandler ClientHandler
	ServerInitHandler ClientHandler
	Encodings         []Encoding
	PixelFormat       *PixelFormat
	ColorMap          *ColorMap
	ClientMessageCh   chan ClientMessage
	ServerMessageCh   chan ServerMessage
	Exclusive         bool
	ServerMessages    []ServerMessage
}
