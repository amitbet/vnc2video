package vnc

import (
	"encoding/binary"
	"fmt"
)

const (
	AteniKVMFrontGroundEventMsgType ServerMessageType = 4
	AteniKVMKeepAliveEventMsgType   ServerMessageType = 22
	AteniKVMVideoGetInfoMsgType     ServerMessageType = 51
	AteniKVMMouseGetInfoMsgType     ServerMessageType = 55
	AteniKVMSessionMessageMsgType   ServerMessageType = 57
	AteniKVMGetViewerLangMsgType    ServerMessageType = 60
)

type AteniKVMFrontGroundEvent struct {
	_ [20]byte
}

func (msg *AteniKVMFrontGroundEvent) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

func (*AteniKVMFrontGroundEvent) Type() ServerMessageType {
	return AteniKVMFrontGroundEventMsgType
}

func (*AteniKVMFrontGroundEvent) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMFrontGroundEvent{}
	var pad [20]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

func (*AteniKVMFrontGroundEvent) Write(c Conn) error {
	var pad [20]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

type AteniKVMKeepAliveEvent struct {
	_ [1]byte
}

func (msg *AteniKVMKeepAliveEvent) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

func (*AteniKVMKeepAliveEvent) Type() ServerMessageType {
	return AteniKVMKeepAliveEventMsgType
}

func (*AteniKVMKeepAliveEvent) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMKeepAliveEvent{}
	var pad [1]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

func (*AteniKVMKeepAliveEvent) Write(c Conn) error {
	var pad [1]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

type AteniKVMVideoGetInfo struct {
	_ [20]byte
}

func (msg *AteniKVMVideoGetInfo) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

func (*AteniKVMVideoGetInfo) Type() ServerMessageType {
	return AteniKVMVideoGetInfoMsgType
}

func (*AteniKVMVideoGetInfo) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMVideoGetInfo{}
	var pad [40]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

func (*AteniKVMVideoGetInfo) Write(c Conn) error {
	var pad [4]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

type AteniKVMMouseGetInfo struct {
	_ [2]byte
}

func (msg *AteniKVMMouseGetInfo) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

func (*AteniKVMMouseGetInfo) Type() ServerMessageType {
	return AteniKVMMouseGetInfoMsgType
}

func (*AteniKVMMouseGetInfo) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMFrontGroundEvent{}
	var pad [2]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

func (*AteniKVMMouseGetInfo) Write(c Conn) error {
	var pad [2]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

type AteniKVMSessionMessage struct {
	_ [264]byte
}

func (msg *AteniKVMSessionMessage) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

func (*AteniKVMSessionMessage) Type() ServerMessageType {
	return AteniKVMFrontGroundEventMsgType
}

func (*AteniKVMSessionMessage) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMSessionMessage{}
	var pad [264]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

func (*AteniKVMSessionMessage) Write(c Conn) error {
	var pad [264]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}

type AteniKVMGetViewerLang struct {
	_ [8]byte
}

func (msg *AteniKVMGetViewerLang) String() string {
	return fmt.Sprintf("%s", msg.Type())
}

func (*AteniKVMGetViewerLang) Type() ServerMessageType {
	return AteniKVMGetViewerLangMsgType
}

func (*AteniKVMGetViewerLang) Read(c Conn) (ServerMessage, error) {
	msg := &AteniKVMGetViewerLang{}
	var pad [8]byte
	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		return nil, err
	}
	return msg, nil
}

func (*AteniKVMGetViewerLang) Write(c Conn) error {
	var pad [8]byte
	if err := binary.Write(c, binary.BigEndian, pad); err != nil {
		return err
	}
	return nil
}
