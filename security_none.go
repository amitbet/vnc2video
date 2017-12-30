package vnc2webm

type ClientAuthNone struct{}

func (*ClientAuthNone) Type() SecurityType {
	return SecTypeNone
}

func (*ClientAuthNone) SubType() SecuritySubType {
	return SecSubTypeUnknown
}

func (*ClientAuthNone) Auth(conn Conn) error {
	return nil
}

// ServerAuthNone is the "none" authentication. See 7.2.1.
type ServerAuthNone struct{}

func (*ServerAuthNone) Type() SecurityType {
	return SecTypeNone
}

func (*ServerAuthNone) SubType() SecuritySubType {
	return SecSubTypeUnknown
}

func (*ServerAuthNone) Auth(c Conn) error {
	return nil
}
