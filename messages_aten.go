package vnc2video

import (
	"encoding/binary"
	"fmt"
)

// Aten IKVM server message types
const (
	AteniKVMFrontGroundEventMsgType ServerMessageType = 4
	AteniKVMKeepAliveEventMsgType   ServerMessageType = 22
	AteniKVMVideoGetInfoMsgType     ServerMessageType = 51
	AteniKVMMouseGetInfoMsgType     ServerMessageType = 55
	AteniKVMSessionMessageMsgType   ServerMessageType = 57
	AteniKVMGetViewerLangMsgType    ServerMessageType = 60
)

// Aten IKVM client message types
const (
	AteniKVMKeyEventMsgType     ClientMessageType = 4
	AteniKVMPointerEventMsgType ClientMessageType = 5
)

// AteniKVMKeyEvent holds the wire format message
type AteniKVMKeyEvent struct {
	_    [1]byte // padding
	Down uint8   // down-flag
	_    [2]byte // padding
	Key  Key     // key
	_    [9]byte // padding
}

// AteniKVMPointerEvent holds the wire format message
type AteniKVMPointerEvent struct {
	_    [1]byte  // padding
	Mask uint8    // mask
	X    uint16   // x
	Y    uint16   // y
	_    [11]byte // padding
}

func (msg *AteniKVMPointerEvent) Supported(c Conn) bool {
	return false
}

func (msg *AteniKVMPointerEvent) String() string {
	return fmt.Sprintf("mask: %d, x:%d, y:%d", msg.Mask, msg.X, msg.Y)
}

func (msg *AteniKVMPointerEvent) Type() ClientMessageType {
	return AteniKVMPointerEventMsgType
}

func (*AteniKVMPointerEvent) Read(c Conn) (ClientMessage, error) {
	msg := AteniKVMPointerEvent{}
	if err := binary.Read(c, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *AteniKVMPointerEvent) Write(c Conn) error {
	if !msg.Supported(c) {
		return nil
	}
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, msg); err != nil {
		return err
	}
	return c.Flush()
}

func (msg *AteniKVMKeyEvent) Supported(c Conn) bool {
	return false
}

func (msg *AteniKVMKeyEvent) String() string {
	return fmt.Sprintf("down:%d, key:%s", msg.Down, msg.Key)
}

func (msg *AteniKVMKeyEvent) Type() ClientMessageType {
	return AteniKVMKeyEventMsgType
}

func (*AteniKVMKeyEvent) Read(c Conn) (ClientMessage, error) {
	msg := AteniKVMKeyEvent{}
	if err := binary.Read(c, binary.BigEndian, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

func (msg *AteniKVMKeyEvent) Write(c Conn) error {
	if !msg.Supported(c) {
		return nil
	}
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, msg); err != nil {
		return err
	}
	return c.Flush()
}

// AteniKVMFrontGroundEvent unknown aten ikvm message
type AteniKVMFrontGroundEvent struct {
	_ [20]byte
}

func (msg *AteniKVMFrontGroundEvent) Supported(c Conn) bool {
	return false
}

// String return string representation
func (msg *AteniKVMFrontGroundEvent) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

// Type return ServerMessageType
func (*AteniKVMFrontGroundEvent) Type() ServerMessageType {
	return AteniKVMFrontGroundEventMsgType
}

// Read unmarshal message from conn
func (*AteniKVMFrontGroundEvent) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMFrontGroundEvent{}
	var pad [20]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write marshal message to conn
func (msg *AteniKVMFrontGroundEvent) Write(c Conn) error {
	if !msg.Supported(c) {
		return nil
	}
	var pad [20]byte
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return c.Flush()
}

// AteniKVMKeepAliveEvent unknown aten ikvm message
type AteniKVMKeepAliveEvent struct {
	_ [1]byte
}

func (msg *AteniKVMKeepAliveEvent) Supported(c Conn) bool {
	return false
}

// String return string representation
func (msg *AteniKVMKeepAliveEvent) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

// Type return ServerMessageType
func (*AteniKVMKeepAliveEvent) Type() ServerMessageType {
	return AteniKVMKeepAliveEventMsgType
}

// Read unmarshal message from conn
func (*AteniKVMKeepAliveEvent) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMKeepAliveEvent{}
	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write marshal message to conn
func (msg *AteniKVMKeepAliveEvent) Write(c Conn) error {
	if !msg.Supported(c) {
		return nil
	}
	var pad [1]byte
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return c.Flush()
}

// AteniKVMVideoGetInfo unknown aten ikvm message
type AteniKVMVideoGetInfo struct {
	_ [20]byte
}

func (msg *AteniKVMVideoGetInfo) Supported(c Conn) bool {
	return false
}

// String return string representation
func (msg *AteniKVMVideoGetInfo) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

// Type return ServerMessageType
func (*AteniKVMVideoGetInfo) Type() ServerMessageType {
	return AteniKVMVideoGetInfoMsgType
}

// Read unmarshal message from conn
func (*AteniKVMVideoGetInfo) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMVideoGetInfo{}
	var pad [40]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write marshal message to conn
func (msg *AteniKVMVideoGetInfo) Write(c Conn) error {
	if !msg.Supported(c) {
		return nil
	}
	var pad [4]byte
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return c.Flush()
}

// AteniKVMMouseGetInfo unknown aten ikvm message
type AteniKVMMouseGetInfo struct {
	_ [2]byte
}

func (msg *AteniKVMMouseGetInfo) Supported(c Conn) bool {
	return false
}

// String return string representation
func (msg *AteniKVMMouseGetInfo) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

// Type return ServerMessageType
func (*AteniKVMMouseGetInfo) Type() ServerMessageType {
	return AteniKVMMouseGetInfoMsgType
}

// Read unmarshal message from conn
func (*AteniKVMMouseGetInfo) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMFrontGroundEvent{}
	var pad [2]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write marshal message to conn
func (msg *AteniKVMMouseGetInfo) Write(c Conn) error {
	if !msg.Supported(c) {
		return nil
	}
	var pad [2]byte
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return c.Flush()
}

// AteniKVMSessionMessage unknown aten ikvm message
type AteniKVMSessionMessage struct {
	_ [264]byte
}

func (msg *AteniKVMSessionMessage) Supported(c Conn) bool {
	return false
}

// String return string representation
func (msg *AteniKVMSessionMessage) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

// Type return ServerMessageType
func (*AteniKVMSessionMessage) Type() ServerMessageType {
	return AteniKVMSessionMessageMsgType
}

// Read unmarshal message from conn
func (*AteniKVMSessionMessage) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMSessionMessage{}
	var pad [264]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write marshal message to conn
func (msg *AteniKVMSessionMessage) Write(c Conn) error {
	if !msg.Supported(c) {
		return nil
	}
	var pad [264]byte
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

// AteniKVMGetViewerLang unknown aten ikvm message
type AteniKVMGetViewerLang struct {
	_ [8]byte
}

func (msg *AteniKVMGetViewerLang) Supported(c Conn) bool {
	return false
}

// String return string representation
func (msg *AteniKVMGetViewerLang) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

// Type return ServerMessageType
func (*AteniKVMGetViewerLang) Type() ServerMessageType {
	return AteniKVMGetViewerLangMsgType
}

// Read unmarshal message from conn
func (*AteniKVMGetViewerLang) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMGetViewerLang{}
	var pad [8]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

// Write marshal message to conn
func (msg *AteniKVMGetViewerLang) Write(c Conn) error {
	if !msg.Supported(c) {
		return nil
	}
	var pad [8]byte
	if err := binary.Write(c, binary.BigEndian, msg.Type()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return c.Flush()
}
