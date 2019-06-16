package vnc2video

import (
	"encoding/binary"
	"fmt"
	"github.com/amitbet/vnc2video/logger"
)

// Handler represents handler of handshake
type Handler interface {
	Handle(Conn) error
}

// ProtoVersionLength protocol version length
const ProtoVersionLength = 12

const (
	// ProtoVersionUnknown unknown version
	ProtoVersionUnknown = ""
	// ProtoVersion33 sets if proto 003.003
	ProtoVersion33 = "RFB 003.003\n"
	// ProtoVersion38 sets if proto 003.008
	ProtoVersion38 = "RFB 003.008\n"
	// ProtoVersion37 sets if proto 003.007
	ProtoVersion37 = "RFB 003.007\n"
)

// ParseProtoVersion parse protocol version
func ParseProtoVersion(pv []byte) (uint, uint, error) {
	var major, minor uint

	if len(pv) < ProtoVersionLength {
		return 0, 0, fmt.Errorf("ProtocolVersion message too short (%v < %v)", len(pv), ProtoVersionLength)
	}

	l, err := fmt.Sscanf(string(pv), "RFB %d.%d\n", &major, &minor)
	if l != 2 {
		return 0, 0, fmt.Errorf("error parsing protocol version")
	}
	if err != nil {
		return 0, 0, err
	}

	return major, minor, nil
}

// DefaultClientVersionHandler represents default handler
type DefaultClientVersionHandler struct{}

// Handle provide version handler for client side
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

// DefaultServerVersionHandler represents default server handler
type DefaultServerVersionHandler struct{}

// Handle provide server version handler
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

// DefaultClientSecurityHandler used for client security handler
type DefaultClientSecurityHandler struct{}

// Handle provide client side security handler
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
		logger.Error("Authentication error: ", err)
		return err
	}

	var authCode uint32
	if err := binary.Read(c, binary.BigEndian, &authCode); err != nil {
		return err
	}

	logger.Tracef("authenticating, secType: %d, auth code(0=success): %d", secType.Type(), authCode)
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
	c.SetSecurityHandler(secType)
	return nil
}

// DefaultServerSecurityHandler used for server security handler
type DefaultServerSecurityHandler struct{}

// Handle provide server side security handler
func (*DefaultServerSecurityHandler) Handle(c Conn) error {
	cfg := c.Config().(*ServerConfig)
	var secType SecurityType
	if c.Protocol() == ProtoVersion37 || c.Protocol() == ProtoVersion38 {
		if err := binary.Write(c, binary.BigEndian, uint8(len(cfg.SecurityHandlers))); err != nil {
			return err
		}

		for _, sectype := range cfg.SecurityHandlers {
			if err := binary.Write(c, binary.BigEndian, sectype.Type()); err != nil {
				return err
			}
		}
	} else {
		st := uint32(0)
		for _, sectype := range cfg.SecurityHandlers {
			if uint32(sectype.Type()) > st {
				st = uint32(sectype.Type())
				secType = sectype.Type()
			}
		}
		if err := binary.Write(c, binary.BigEndian, st); err != nil {
			return err
		}
	}
	if err := c.Flush(); err != nil {
		return err
	}

	if c.Protocol() == ProtoVersion38 {
		if err := binary.Read(c, binary.BigEndian, &secType); err != nil {
			return err
		}
	}
	secTypes := make(map[SecurityType]SecurityHandler)
	for _, sType := range cfg.SecurityHandlers {
		secTypes[sType.Type()] = sType
	}

	sType, ok := secTypes[secType]
	if !ok {
		return fmt.Errorf("security type %d not implemented", secType)
	}

	var authCode uint32
	authErr := sType.Auth(c)
	if authErr != nil {
		authCode = uint32(1)
	}

	if err := binary.Write(c, binary.BigEndian, authCode); err != nil {
		return err
	}

	if authErr == nil {
		if err := c.Flush(); err != nil {
			return err
		}
		c.SetSecurityHandler(sType)
		return nil
	}

	if c.Protocol() == ProtoVersion38 {
		if err := binary.Write(c, binary.BigEndian, uint32(len(authErr.Error()))); err != nil {
			return err
		}
		if err := binary.Write(c, binary.BigEndian, []byte(authErr.Error())); err != nil {
			return err
		}
		if err := c.Flush(); err != nil {
			return err
		}
	}
	return authErr
}

// DefaultClientServerInitHandler default client server init handler
type DefaultClientServerInitHandler struct{}

// Handle provide default server init handler
func (*DefaultClientServerInitHandler) Handle(c Conn) error {
	logger.Trace("starting DefaultClientServerInitHandler")
	var err error
	srvInit := ServerInit{}

	if err = binary.Read(c, binary.BigEndian, &srvInit.FBWidth); err != nil {
		return err
	}
	if err = binary.Read(c, binary.BigEndian, &srvInit.FBHeight); err != nil {
		return err
	}
	if err = binary.Read(c, binary.BigEndian, &srvInit.PixelFormat); err != nil {
		return err
	}
	if err = binary.Read(c, binary.BigEndian, &srvInit.NameLength); err != nil {
		return err
	}

	srvInit.NameText = make([]byte, srvInit.NameLength)
	if err = binary.Read(c, binary.BigEndian, &srvInit.NameText); err != nil {
		return err
	}
	logger.Tracef("DefaultClientServerInitHandler got serverInit: %v", srvInit)
	c.SetDesktopName(srvInit.NameText)
	if c.Protocol() == "aten1" {
		c.SetWidth(800)
		c.SetHeight(600)
		c.SetPixelFormat(NewPixelFormatAten())
	} else {
		c.SetWidth(srvInit.FBWidth)
		c.SetHeight(srvInit.FBHeight)

		//telling the server to use 32bit pixels (with 24 dept, tight standard format)
		pixelMsg := SetPixelFormat{PF: PixelFormat32bit}
		pixelMsg.Write(c)
		c.SetPixelFormat(PixelFormat32bit)
		//c.SetPixelFormat(srvInit.PixelFormat)
	}
	if c.Protocol() == "aten1" {
		ikvm := struct {
			_               [8]byte
			IKVMVideoEnable uint8
			IKVMKMEnable    uint8
			IKVMKickEnable  uint8
			VUSBEnable      uint8
		}{}
		if err = binary.Read(c, binary.BigEndian, &ikvm); err != nil {
			return err
		}
	}
	/*
		caps := struct {
			ServerMessagesNum uint16
			ClientMessagesNum uint16
			EncodingsNum      uint16
			_                 [2]byte
		}{}
		if err := binary.Read(c, binary.BigEndian, &caps); err != nil {
			return err
		}

		caps.ServerMessagesNum = uint16(1)
		var item [16]byte
		for i := uint16(0); i < caps.ServerMessagesNum; i++ {
			if err := binary.Read(c, binary.BigEndian, &item); err != nil {
				return err
			}
			fmt.Printf("server message cap %s\n", item)
		}

			for i := uint16(0); i < caps.ClientMessagesNum; i++ {
				if err := binary.Read(c, binary.BigEndian, &item); err != nil {
					return err
				}
				fmt.Printf("client message cap %s\n", item)
			}
			for i := uint16(0); i < caps.EncodingsNum; i++ {
				if err := binary.Read(c, binary.BigEndian, &item); err != nil {
					return err
				}
				fmt.Printf("encoding cap %s\n", item)
			}
		//	var pad [1]byte
		//	if err := binary.Read(c, binary.BigEndian, &pad); err != nil {
		//		return err
		//	}
	}*/
	return nil
}

// DefaultServerServerInitHandler default server server init handler
type DefaultServerServerInitHandler struct{}

// Handle provide default server server init handler
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

// DefaultClientClientInitHandler default client client init handler
type DefaultClientClientInitHandler struct{}

// Handle provide default client client init handler
func (*DefaultClientClientInitHandler) Handle(c Conn) error {
	logger.Trace("starting DefaultClientClientInitHandler")
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
	logger.Tracef("DefaultClientClientInitHandler sending: shared=%d", shared)
	return c.Flush()
}

// DefaultServerClientInitHandler default server client init handler
type DefaultServerClientInitHandler struct{}

// Handle provide default server client init handler
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
