package vnc

import (
	"encoding/binary"
	"fmt"
)

// Aten IKVM message types.
const (
	AteniKVMFrontGroundEventMsgType ServerMessageType = 4
	AteniKVMKeepAliveEventMsgType   ServerMessageType = 22
	AteniKVMVideoGetInfoMsgType     ServerMessageType = 51
	AteniKVMMouseGetInfoMsgType     ServerMessageType = 55
	AteniKVMSessionMessageMsgType   ServerMessageType = 57
	AteniKVMGetViewerLangMsgType    ServerMessageType = 60
)

// AteniKVMFrontGroundEvent unknown aten ikvm message
type AteniKVMFrontGroundEvent struct {
	_ [20]byte
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
func (*AteniKVMFrontGroundEvent) Write(c Conn) error {
	var pad [20]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

// AteniKVMKeepAliveEvent unknown aten ikvm message
type AteniKVMKeepAliveEvent struct {
	_ [1]byte
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
func (*AteniKVMKeepAliveEvent) Write(c Conn) error {
	var pad [1]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

// AteniKVMVideoGetInfo unknown aten ikvm message
type AteniKVMVideoGetInfo struct {
	_ [20]byte
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
func (*AteniKVMVideoGetInfo) Write(c Conn) error {
	var pad [4]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

// AteniKVMMouseGetInfo unknown aten ikvm message
type AteniKVMMouseGetInfo struct {
	_ [2]byte
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
func (*AteniKVMMouseGetInfo) Write(c Conn) error {
	var pad [2]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

// AteniKVMSessionMessage unknown aten ikvm message
type AteniKVMSessionMessage struct {
	_ [264]byte
}

// String return string representation
func (msg *AteniKVMSessionMessage) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

// Type return ServerMessageType
func (*AteniKVMSessionMessage) Type() ServerMessageType {
	return AteniKVMFrontGroundEventMsgType
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
func (*AteniKVMSessionMessage) Write(c Conn) error {
	var pad [264]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

// AteniKVMGetViewerLang unknown aten ikvm message
type AteniKVMGetViewerLang struct {
	_ [8]byte
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
func (*AteniKVMGetViewerLang) Write(c Conn) error {
	var pad [8]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}
