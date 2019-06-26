package vnc2video

type SecurityType uint8

//go:generate stringer -type=SecurityType

const (
	SecTypeUnknown           SecurityType = SecurityType(0)
	SecTypeNone              SecurityType = SecurityType(1)
	SecTypeVNC               SecurityType = SecurityType(2)
	SecTypeTight             SecurityType = SecurityType(16)
	SecTypeATEN              SecurityType = SecurityType(16)
	SecTypeVeNCrypt          SecurityType = SecurityType(19)
	SecTypeUltraMsAutoLogon2 SecurityType = SecurityType(113)
)

type SecuritySubType uint32

//go:generate stringer -type=SecuritySubType

const (
	SecSubTypeUnknown SecuritySubType = SecuritySubType(0)
)

const (
	SecSubTypeVeNCrypt01Unknown   SecuritySubType = SecuritySubType(0)
	SecSubTypeVeNCrypt01Plain     SecuritySubType = SecuritySubType(19)
	SecSubTypeVeNCrypt01TLSNone   SecuritySubType = SecuritySubType(20)
	SecSubTypeVeNCrypt01TLSVNC    SecuritySubType = SecuritySubType(21)
	SecSubTypeVeNCrypt01TLSPlain  SecuritySubType = SecuritySubType(22)
	SecSubTypeVeNCrypt01X509None  SecuritySubType = SecuritySubType(23)
	SecSubTypeVeNCrypt01X509VNC   SecuritySubType = SecuritySubType(24)
	SecSubTypeVeNCrypt01X509Plain SecuritySubType = SecuritySubType(25)
)

const (
	SecSubTypeVeNCrypt02Unknown   SecuritySubType = SecuritySubType(0)
	SecSubTypeVeNCrypt02Plain     SecuritySubType = SecuritySubType(256)
	SecSubTypeVeNCrypt02TLSNone   SecuritySubType = SecuritySubType(257)
	SecSubTypeVeNCrypt02TLSVNC    SecuritySubType = SecuritySubType(258)
	SecSubTypeVeNCrypt02TLSPlain  SecuritySubType = SecuritySubType(259)
	SecSubTypeVeNCrypt02X509None  SecuritySubType = SecuritySubType(260)
	SecSubTypeVeNCrypt02X509VNC   SecuritySubType = SecuritySubType(261)
	SecSubTypeVeNCrypt02X509Plain SecuritySubType = SecuritySubType(262)
)

type SecurityHandler interface {
	Type() SecurityType
	SubType() SecuritySubType
	Auth(Conn) error
}
