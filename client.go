package vnc

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

var DefaultServerMessages = []ServerMessage{
	&FramebufferUpdate{},
	&SetColorMapEntries{},
	&Bell{},
	&ServerCutText{},
}

var (
	DefaultClientHandlers []ClientHandler = []ClientHandler{
		&DefaultClientVersionHandler{},
		&DefaultClientSecurityHandler{},
		&DefaultClientClientInitHandler{},
		&DefaultClientServerInitHandler{},
		//		&DefaultClientMessageHandler{},
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
		fmt.Printf("%#+v\n", h)
		if err := h.Handle(conn); err != nil {
			fmt.Printf("rrr %v\n", err)
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
func (c *ClientConn) DesktopName() []byte {
	return c.desktopName
}
func (c *ClientConn) PixelFormat() *PixelFormat {
	return c.pixelFormat
}
func (c *ClientConn) SetDesktopName(name []byte) {
	c.desktopName = name
}
func (c *ClientConn) SetPixelFormat(pf *PixelFormat) error {
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
	colorMap *ColorMap

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
	pixelFormat *PixelFormat

	quitCh  chan struct{}
	errorCh chan error
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
		quitCh:      cfg.QuitCh,
		errorCh:     cfg.ErrorCh,
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
	_  [3]byte     // padding
	PF PixelFormat // pixel-format
}

func (*SetPixelFormat) Type() ClientMessageType {
	return SetPixelFormatMsgType
}

func (msg *SetPixelFormat) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
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

	return c.Flush()
}

func (*SetPixelFormat) Read(c Conn) (ClientMessage, error) {
	msg := SetPixelFormat{}
	if err := binary.Read(c, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// SetEncodings holds the wire format message, sans encoding-type field.
type SetEncodings struct {
	_         [1]byte // padding
	EncNum    uint16  // number-of-encodings
	Encodings []EncodingType
}

func (*SetEncodings) Type() ClientMessageType {
	return SetEncodingsMsgType
}

func (*SetEncodings) Read(c Conn) (ClientMessage, error) {
	msg := SetEncodings{}
	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}

	if err := binary.Read(c, binary.BigEndian, &msg.EncNum); err != nil {
		return nil, err
	}
	var enc EncodingType
	for i := uint16(0); i < msg.EncNum; i++ {
		if err := binary.Read(c, binary.BigEndian, &enc); err != nil {
			return nil, err
		}
		msg.Encodings = append(msg.Encodings, enc)
	}
	c.SetEncodings(msg.Encodings)
	return &msg, nil
}

func (msg *SetEncodings) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
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
	Inc           uint8  // incremental
	X, Y          uint16 // x-, y-position
	Width, Height uint16 // width, height
}

func (*FramebufferUpdateRequest) Type() ClientMessageType {
	return FramebufferUpdateRequestMsgType
}

func (*FramebufferUpdateRequest) Read(c Conn) (ClientMessage, error) {
	msg := FramebufferUpdateRequest{}
	if err := binary.Read(c, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *FramebufferUpdateRequest) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, msg); err != nil {
		return err
	}
	return c.Flush()
}

// KeyEvent holds the wire format message.
type KeyEvent struct {
	Down uint8   // down-flag
	_    [2]byte // padding
	Key  Key     // key
}

func (*KeyEvent) Type() ClientMessageType {
	return KeyEventMsgType
}

func (*KeyEvent) Read(c Conn) (ClientMessage, error) {
	msg := KeyEvent{}
	if err := binary.Read(c, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *KeyEvent) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, msg); err != nil {
		return err
	}
	return c.Flush()
}

// PointerEventMessage holds the wire format message.
type PointerEvent struct {
	Mask uint8  // button-mask
	X, Y uint16 // x-, y-position
}

func (*PointerEvent) Type() ClientMessageType {
	return PointerEventMsgType
}

func (*PointerEvent) Read(c Conn) (ClientMessage, error) {
	msg := PointerEvent{}
	if err := binary.Read(c, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *PointerEvent) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, msg); err != nil {
		return err
	}
	return c.Flush()
}

// ClientCutText holds the wire format message, sans the text field.
type ClientCutText struct {
	_      [3]byte // padding
	Length uint32  // length
	Text   []byte
}

func (*ClientCutText) Type() ClientMessageType {
	return ClientCutTextMsgType
}

func (*ClientCutText) Read(c Conn) (ClientMessage, error) {
	msg := ClientCutText{}
	var pad [3]byte
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

func (msg *ClientCutText) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
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

type ClientHandler interface {
	Handle(Conn) error
}

// A ClientConfig structure is used to configure a ClientConn. After
// one has been passed to initialize a connection, it must not be modified.
type ClientConfig struct {
	Handlers         []ClientHandler
	SecurityHandlers []SecurityHandler
	Encodings        []Encoding
	PixelFormat      *PixelFormat
	ColorMap         *ColorMap
	ClientMessageCh  chan ClientMessage
	ServerMessageCh  chan ServerMessage
	Exclusive        bool
	ServerMessages   []ServerMessage
	QuitCh           chan struct{}
	ErrorCh          chan error
}
