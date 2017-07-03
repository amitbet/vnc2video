package vnc

import (
	"encoding/binary"
	"fmt"
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

func (srvInit ServerInit) String() string {
	return fmt.Sprintf("Width: %d, Height: %d, PixelFormat: %s, NameLength: %d, MameText: %s", srvInit.FBWidth, srvInit.FBHeight, srvInit.PixelFormat, srvInit.NameLength, srvInit.NameText)
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

func (msg *FramebufferUpdate) String() string {
	return fmt.Sprintf("rects %d rectangle[]: { %v }", msg.NumRect, msg.Rects)
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
		rect := NewRectangle()
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

type ServerCutText struct {
	_      [1]byte
	Length uint32
	Text   []byte
}

func (msg *ServerCutText) String() string {
	return fmt.Sprintf("lenght: %d text: %s", msg.Length, msg.Text)
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
	return c.Flush()
}

type Bell struct{}

func (*Bell) String() string {
	return fmt.Sprintf("bell")
}

func (*Bell) Type() ServerMessageType {
	return BellMsgType
}

func (*Bell) Read(c Conn) (ServerMessage, error) {
	return &Bell{}, nil
}

func (msg *Bell) Write(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	return c.Flush()
}

type SetColorMapEntries struct {
	_          [1]byte
	FirstColor uint16
	ColorsNum  uint16
	Colors     []Color
}

func (msg *SetColorMapEntries) String() string {
	return fmt.Sprintf("first color: %d, numcolors: %d, colors[]: { %v }", msg.FirstColor, msg.ColorsNum, msg.Colors)
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

	return c.Flush()
}

var DefaultServerMessages = []ServerMessage{
	&FramebufferUpdate{},
	&SetColorMapEntries{},
	&Bell{},
	&ServerCutText{},
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

func (msg *SetPixelFormat) String() string {
	return fmt.Sprintf("%s", msg.PF)
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
		c.SetColorMap(ColorMap{})
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

func (msg *SetEncodings) String() string {
	return fmt.Sprintf("encnum: %d, encodings[]: { %v }", msg.EncNum, msg.Encodings)
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

func (msg *FramebufferUpdateRequest) String() string {
	return fmt.Sprintf("incremental: %d, x: %d, y: %d, width: %d, height: %d", msg.Inc, msg.X, msg.Y, msg.Width, msg.Height)
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

func (msg *KeyEvent) String() string {
	return fmt.Sprintf("down: %d, key: %v", msg.Down, msg.Key)
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

func (msg *PointerEvent) String() string {
	return fmt.Sprintf("mask %d, x: %d, y: %d", msg.Mask, msg.X, msg.Y)
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

func (msg *ClientCutText) String() string {
	return fmt.Sprintf("length: %d, text: %s", msg.Length, msg.Text)
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
