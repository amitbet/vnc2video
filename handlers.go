package vnc

// ClientMessage is the interface
type ClientMessage interface {
	Type() ClientMessageType
	Read(Conn) error
	Write(Conn) error
}

// ServerMessage is the interface
type ServerMessage interface {
	Type() ServerMessageType
	Read(Conn) error
	Write(Conn) error
}

func ServerVersionHandler(cfg *ServerConfig, c Conn) error {
	return nil
}

func ServerSecurityHandler(cfg *ServerConfig, c Conn) error {
	return nil
}

func ServerSecurityNoneHandler(cfg *ServerConfig, c Conn) error {
	return nil
}

func ServerServerInitHandler(cfg *ServerConfig, c Conn) error {
	return nil
}

func ServerClientInitHandler(cfg *ServerConfig, c Conn) error {
	return nil
}
