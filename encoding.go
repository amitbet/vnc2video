package vnc

import (
	"bytes"
	"sync"
)

// EncodingType represents a known VNC encoding type.
type EncodingType int32

//go:generate stringer -type=EncodingType

const (
	EncRaw                           EncodingType = 0
	EncCopyRect                      EncodingType = 1
	EncRRE                           EncodingType = 2
	EncCoRRE                         EncodingType = 4
	EncHextile                       EncodingType = 5
	EncZlib                          EncodingType = 6
	EncTight                         EncodingType = 7
	EncZlibHex                       EncodingType = 8
	EncUltra1                        EncodingType = 9
	EncUltra2                        EncodingType = 10
	EncJPEG                          EncodingType = 21
	EncJRLE                          EncodingType = 22
	EncTRLE                          EncodingType = 15
	EncZRLE                          EncodingType = 16
	EncJPEGQualityLevelPseudo10      EncodingType = -23
	EncJPEGQualityLevelPseudo9       EncodingType = -24
	EncJPEGQualityLevelPseudo8       EncodingType = -25
	EncJPEGQualityLevelPseudo7       EncodingType = -26
	EncJPEGQualityLevelPseudo6       EncodingType = -27
	EncJPEGQualityLevelPseudo5       EncodingType = -28
	EncJPEGQualityLevelPseudo4       EncodingType = -29
	EncJPEGQualityLevelPseudo3       EncodingType = -30
	EncJPEGQualityLevelPseudo2       EncodingType = -31
	EncJPEGQualityLevelPseudo1       EncodingType = -32
	EncColorPseudo                   EncodingType = -239
	EncXCursorPseudo                 EncodingType = -240
	EncDesktopSizePseudo             EncodingType = -223
	EncLastRectPseudo                EncodingType = -224
	EncCompressionLevel10            EncodingType = -247
	EncCompressionLevel9             EncodingType = -248
	EncCompressionLevel8             EncodingType = -249
	EncCompressionLevel7             EncodingType = -250
	EncCompressionLevel6             EncodingType = -251
	EncCompressionLevel5             EncodingType = -252
	EncCompressionLevel4             EncodingType = -253
	EncCompressionLevel3             EncodingType = -254
	EncCompressionLevel2             EncodingType = -255
	EncCompressionLevel1             EncodingType = -256
	EncQEMUPointerMotionChangePseudo EncodingType = -257
	EncQEMUExtendedKeyEventPseudo    EncodingType = -258
	EncTightPng                      EncodingType = -260
	EncDesktopNamePseudo             EncodingType = -307
	EncExtendedDesktopSizePseudo     EncodingType = -308
	EncXvpPseudo                     EncodingType = -309
	EncFencePseudo                   EncodingType = -312
	EncContinuousUpdatesPseudo       EncodingType = -313
	EncClientRedirect                EncodingType = -311
)

var bPool = sync.Pool{
	New: func() interface{} {
		// The Pool's New function should generally only return pointer
		// types, since a pointer can be put into the return interface
		// value without an allocation:
		return new(bytes.Buffer)
	},
}

type Encoding interface {
	Type() EncodingType
	Read(Conn, *Rectangle) error
	Write(Conn, *Rectangle) error
}

func setBit(n uint8, pos uint8) uint8 {
	n |= (1 << pos)
	return n
}

func clrBit(n uint8, pos uint8) uint8 {
	n = n &^ (1 << pos)
	return n
}

func hasBit(n uint8, pos uint8) bool {
	v := n & (1 << pos)
	return (v > 0)
}

func getBit(n uint8, pos uint8) uint8 {
	n = n & (1 << pos)
	return n
}
