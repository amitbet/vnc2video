package vnc

type SecurityType uint8

//go:generate stringer -type=SecurityType

const (
	SecTypeUnknown  = SecurityType(0)
	SecTypeNone     = SecurityType(1)
	SecTypeVNC      = SecurityType(2)
	SecTypeTight    = SecurityType(16)
	SecTypeATEN     = SecurityType(16)
	SecTypeVeNCrypt = SecurityType(19)
)

type SecuritySubType uint32

//go:generate stringer -type=SecuritySubType

const (
	SecSubTypeUnknown = SecuritySubType(0)
)

const (
	SecSubTypeVeNCrypt01Unknown   = SecuritySubType(0)
	SecSubTypeVeNCrypt01Plain     = SecuritySubType(19)
	SecSubTypeVeNCrypt01TLSNone   = SecuritySubType(20)
	SecSubTypeVeNCrypt01TLSVNC    = SecuritySubType(21)
	SecSubTypeVeNCrypt01TLSPlain  = SecuritySubType(22)
	SecSubTypeVeNCrypt01X509None  = SecuritySubType(23)
	SecSubTypeVeNCrypt01X509VNC   = SecuritySubType(24)
	SecSubTypeVeNCrypt01X509Plain = SecuritySubType(25)
)

const (
	SecSubTypeVeNCrypt02Unknown   = SecuritySubType(0)
	SecSubTypeVeNCrypt02Plain     = SecuritySubType(256)
	SecSubTypeVeNCrypt02TLSNone   = SecuritySubType(257)
	SecSubTypeVeNCrypt02TLSVNC    = SecuritySubType(258)
	SecSubTypeVeNCrypt02TLSPlain  = SecuritySubType(259)
	SecSubTypeVeNCrypt02X509None  = SecuritySubType(260)
	SecSubTypeVeNCrypt02X509VNC   = SecuritySubType(261)
	SecSubTypeVeNCrypt02X509Plain = SecuritySubType(262)
)

type SecurityHandler interface {
	Type() SecurityType
	SubType() SecuritySubType
	Auth(Conn) error
}
