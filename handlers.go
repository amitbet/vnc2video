package vnc

import (
	"encoding/binary"
	"fmt"
)

// ClientMessage is the interface
type ClientMessage interface {
	Type() ClientMessageType
	Read(Conn) (ClientMessage, error)
	Write(Conn) error
}

// ServerMessage is the interface
type ServerMessage interface {
	Type() ServerMessageType
	Read(Conn) (ServerMessage, error)
	Write(Conn) error
}

const ProtoVersionLength = 12

const (
	ProtoVersionUnknown = ""
	ProtoVersion33      = "RFB 003.003\n"
	ProtoVersion38      = "RFB 003.008\n"
)

func ParseProtoVersion(pv []byte) (uint, uint, error) {
	var major, minor uint

	if len(pv) < ProtoVersionLength {
		return 0, 0, fmt.Errorf("ProtocolVersion message too short (%v < %v)", len(pv), ProtoVersionLength)
	}

	l, err := fmt.Sscanf(string(pv), "RFB %d.%d\n", &major, &minor)
	if l != 2 {
		return 0, 0, fmt.Errorf("error parsing ProtocolVersion.")
	}
	if err != nil {
		return 0, 0, err
	}

	return major, minor, nil
}

type DefaultClientVersionHandler struct{}

func (*DefaultClientVersionHandler) Handle(c Conn) error {
	var version [ProtoVersionLength]byte

	if err := binary.Read(c, binary.BigEndian, &version); err != nil {
		return err
	}

	major, minor, err := ParseProtoVersion(version[:])
	if err != nil {
		return err
	}

	pv := ProtoVersionUnknown
	if major == 3 {
		if minor >= 8 {
			pv = ProtoVersion38
		} else if minor >= 3 {
			pv = ProtoVersion38
		}
	}
	if pv == ProtoVersionUnknown {
		return fmt.Errorf("ProtocolVersion handshake failed; unsupported version '%v'", string(version[:]))
	}
	c.SetProtoVersion(string(version[:]))

	if err := binary.Write(c, binary.BigEndian, []byte(pv)); err != nil {
		return err
	}
	return c.Flush()
}

type DefaultServerVersionHandler struct{}

func (*DefaultServerVersionHandler) Handle(c Conn) error {
	var version [ProtoVersionLength]byte
	if err := binary.Write(c, binary.BigEndian, []byte(ProtoVersion38)); err != nil {
		return err
	}
	if err := c.Flush(); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &version); err != nil {
		return err
	}

	major, minor, err := ParseProtoVersion(version[:])
	if err != nil {
		return err
	}

	pv := ProtoVersionUnknown
	if major == 3 {
		if minor >= 8 {
			pv = ProtoVersion38
		} else if minor >= 3 {
			pv = ProtoVersion33
		}
	}
	if pv == ProtoVersionUnknown {
		return fmt.Errorf("ProtocolVersion handshake failed; unsupported version '%v'", string(version[:]))
	}

	c.SetProtoVersion(pv)
	return nil
}

type DefaultClientSecurityHandler struct{}

func (*DefaultClientSecurityHandler) Handle(c Conn) error {
	cfg := c.Config().(*ClientConfig)
	var numSecurityTypes uint8
	if err := binary.Read(c, binary.BigEndian, &numSecurityTypes); err != nil {
		return err
	}
	secTypes := make([]SecurityType, numSecurityTypes)
	if err := binary.Read(c, binary.BigEndian, &secTypes); err != nil {
		return err
	}

	var secType SecurityHandler
	for _, st := range cfg.SecurityHandlers {
		for _, sc := range secTypes {
			if st.Type() == sc {
				secType = st
			}
		}
	}

	if err := binary.Write(c, binary.BigEndian, cfg.SecurityHandlers[0].Type()); err != nil {
		return err
	}

	if err := c.Flush(); err != nil {
		return err
	}

	err := secType.Auth(c)
	if err != nil {
		return err
	}

	var authCode uint32
	if err := binary.Read(c, binary.BigEndian, &authCode); err != nil {
		return err
	}

	if authCode == 1 {
		var reasonLength uint32
		if err := binary.Read(c, binary.BigEndian, &reasonLength); err != nil {
			return err
		}
		reasonText := make([]byte, reasonLength)
		if err := binary.Read(c, binary.BigEndian, &reasonText); err != nil {
			return err
		}
		return fmt.Errorf("%s", reasonText)
	}

	return nil
}

type DefaultServerSecurityHandler struct{}

func (*DefaultServerSecurityHandler) Handle(c Conn) error {
	cfg := c.Config().(*ServerConfig)
	if err := binary.Write(c, binary.BigEndian, uint8(len(cfg.SecurityHandlers))); err != nil {
		return err
	}

	for _, sectype := range cfg.SecurityHandlers {
		if err := binary.Write(c, binary.BigEndian, sectype.Type()); err != nil {
			return err
		}
	}

	if err := c.Flush(); err != nil {
		return err
	}

	var secType SecurityType
	if err := binary.Read(c, binary.BigEndian, &secType); err != nil {
		return err
	}

	secTypes := make(map[SecurityType]SecurityHandler)
	for _, sType := range cfg.SecurityHandlers {
		secTypes[sType.Type()] = sType
	}

	sType, ok := secTypes[secType]
	if !ok {
		return fmt.Errorf("server type %d not implemented")
	}

	var authCode uint32
	authErr := sType.Auth(c)
	if authErr != nil {
		authCode = uint32(1)
	}

	if err := binary.Write(c, binary.BigEndian, authCode); err != nil {
		return err
	}
	if err := c.Flush(); err != nil {
		return err
	}

	if authErr != nil {
		if err := binary.Write(c, binary.BigEndian, len(authErr.Error())); err != nil {
			return err
		}
		if err := binary.Write(c, binary.BigEndian, []byte(authErr.Error())); err != nil {
			return err
		}
		if err := c.Flush(); err != nil {
			return err
		}
		return authErr
	}

	return nil
}

type DefaultClientServerInitHandler struct{}

func (*DefaultClientServerInitHandler) Handle(c Conn) error {
	srvInit := &ServerInit{}

	if err := binary.Read(c, binary.BigEndian, &srvInit.FBWidth); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &srvInit.FBHeight); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &srvInit.PixelFormat); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &srvInit.NameLength); err != nil {
		return err
	}

	nameText := make([]byte, srvInit.NameLength)
	if err := binary.Read(c, binary.BigEndian, nameText); err != nil {
		return err
	}

	srvInit.NameText = nameText
	c.SetDesktopName(srvInit.NameText)
	c.SetWidth(srvInit.FBWidth)
	c.SetHeight(srvInit.FBHeight)
	c.SetPixelFormat(&srvInit.PixelFormat)

	if c.Protocol() == "aten" {
		fmt.Printf("$$$$$$\n")
		var pad [28]byte
		/*  12
		8 byte unknown
		1 byte IKVMVideoEnable
		1 byte IKVMKMEnable
		1 byte IKVMKickEnable
		1 byte VUSBEnable
		*/
		if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
			return err
		}
		fmt.Printf("rrrr\n")
	}
	return nil
}

type DefaultServerServerInitHandler struct{}

func (*DefaultServerServerInitHandler) Handle(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, c.Width()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, c.Height()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, c.PixelFormat()); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, uint32(len(c.DesktopName()))); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, []byte(c.DesktopName())); err != nil {
		return err
	}

	return c.Flush()
}

type DefaultClientClientInitHandler struct{}

func (*DefaultClientClientInitHandler) Handle(c Conn) error {
	cfg := c.Config().(*ClientConfig)
	var shared uint8
	if cfg.Exclusive {
		shared = 0
	} else {
		shared = 1
	}
	if err := binary.Write(c, binary.BigEndian, shared); err != nil {
		return err
	}
	return c.Flush()
}

type DefaultServerClientInitHandler struct{}

func (*DefaultServerClientInitHandler) Handle(c Conn) error {
	var shared uint8
	if err := binary.Read(c, binary.BigEndian, &shared); err != nil {
		return err
	}
	/* TODO
	if shared != 1 {
		c.SetShared(false)
	}
	*/
	return nil
}
