package vnc

import (
	"bytes"
	"crypto/des"
	"encoding/binary"
	"fmt"
)

type SecurityType uint8

const (
	SecTypeUnknown  = SecurityType(0)
	SecTypeNone     = SecurityType(1)
	SecTypeVNC      = SecurityType(2)
	SecTypeTight    = SecurityType(16)
	SecTypeATEN     = SecurityType(16)
	SecTypeVeNCrypt = SecurityType(19)
)

type SecuritySubType uint32

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

type ClientAuthATEN struct {
	Username []byte
	Password []byte
}

func readTightTunnels(c Conn) (uint32, error) {
	var n uint32
	if err := binary.Read(c, binary.BigEndian, &n); err != nil {
		return 0, err
	}
	return n, nil
}

func (*ClientAuthATEN) Type() SecurityType {
	return SecTypeATEN
}

func (*ClientAuthATEN) SubType() SecuritySubType {
	return SecSubTypeUnknown
}

func charCodeAt(s string, n int) rune {
	for i, r := range s {
		if i == n {
			return r
		}
	}
	return 0
}

func (auth *ClientAuthATEN) Auth(c Conn) error {
	var definedAuthLen = 24

	nt, err := readTightTunnels(c)
	if err != nil {
		return err
	}
	if (nt&0xffff0ff0)>>0 == 0xaff90fb0 {
		c.SetProtoVersion("aten")
		var skip [20]byte
		binary.Read(c, binary.BigEndian, &skip)
		fmt.Printf("skip %s\n", skip)
	}
	/*
		if len(auth.Username) > 24 || len(auth.Password) > 24 {
			return fmt.Errorf("username/password > 24")
		}

		username := make([]byte, 24-len(auth.Username)+1)
		password := make([]byte, 24-len(auth.Password)+1)
		copy(username, auth.Username)
		copy(password, auth.Password)
		username = append(username, []byte("\x00")...)
		password = append(password, []byte("\x00")...)
		challenge := bytes.Join([][]byte{username, password}, []byte(":"))
		if err := binary.Write(c, binary.BigEndian, challenge); err != nil {
			return err
		}

		if err := c.Flush(); err != nil {
			return err
		}
	*/
	sendUsername := make([]byte, definedAuthLen)
	for i := 0; i < definedAuthLen; i++ {
		if i < len(auth.Username) {
			sendUsername[i] = byte(charCodeAt(string(auth.Username), i))
		} else {
			sendUsername[i] = 0
		}
	}

	sendPassword := make([]byte, definedAuthLen)

	for i := 0; i < definedAuthLen; i++ {
		if i < len(auth.Password) {
			sendPassword[i] = byte(charCodeAt(string(auth.Password), i))
		} else {
			sendPassword[i] = 0
		}
	}

	if err := binary.Write(c, binary.BigEndian, sendUsername); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, sendPassword); err != nil {
		return err
	}

	if err := c.Flush(); err != nil {
		return err
	}

	//var pp [10]byte
	//binary.Read(c, binary.BigEndian, &pp)
	//fmt.Printf("ddd %v\n", pp)
	return nil
}

func (*ClientAuthVeNCrypt02Plain) Type() SecurityType {
	return SecTypeVeNCrypt
}

func (*ClientAuthVeNCrypt02Plain) SubType() SecuritySubType {
	return SecSubTypeVeNCrypt02Plain
}

// ClientAuthVeNCryptPlain see https://www.berrange.com/~dan/vencrypt.txt
type ClientAuthVeNCrypt02Plain struct {
	Username []byte
	Password []byte
}

func (auth *ClientAuthVeNCrypt02Plain) Auth(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, []uint8{0, 2}); err != nil {
		return err
	}
	if err := c.Flush(); err != nil {
		return err
	}
	var (
		major, minor uint8
	)

	if err := binary.Read(c, binary.BigEndian, &major); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &minor); err != nil {
		return err
	}
	res := uint8(1)
	if major == 0 && minor == 2 {
		res = uint8(0)
	}
	if err := binary.Write(c, binary.BigEndian, res); err != nil {
		return err
	}
	c.Flush()
	if err := binary.Write(c, binary.BigEndian, uint8(1)); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, auth.SubType()); err != nil {
		return err
	}
	if err := c.Flush(); err != nil {
		return err
	}
	var secType SecuritySubType
	if err := binary.Read(c, binary.BigEndian, &secType); err != nil {
		return err
	}
	if secType != auth.SubType() {
		binary.Write(c, binary.BigEndian, uint8(1))
		c.Flush()
		return fmt.Errorf("invalid sectype")
	}
	if len(auth.Password) == 0 || len(auth.Username) == 0 {
		return fmt.Errorf("Security Handshake failed; no username and/or password provided for VeNCryptAuth.")
	}
	/*
		if err := binary.Write(c, binary.BigEndian, uint32(len(auth.Username))); err != nil {
			return err
		}

		if err := binary.Write(c, binary.BigEndian, uint32(len(auth.Password))); err != nil {
			return err
		}

		if err := binary.Write(c, binary.BigEndian, auth.Username); err != nil {
			return err
		}

		if err := binary.Write(c, binary.BigEndian, auth.Password); err != nil {
			return err
		}
	*/
	var (
		uLength, pLength uint32
	)
	if err := binary.Read(c, binary.BigEndian, &uLength); err != nil {
		return err
	}
	if err := binary.Read(c, binary.BigEndian, &pLength); err != nil {
		return err
	}

	username := make([]byte, uLength)
	password := make([]byte, pLength)
	if err := binary.Read(c, binary.BigEndian, &username); err != nil {
		return err
	}

	if err := binary.Read(c, binary.BigEndian, &password); err != nil {
		return err
	}
	if !bytes.Equal(auth.Username, username) || !bytes.Equal(auth.Password, password) {
		return fmt.Errorf("invalid username/password")
	}
	return nil
}

// ServerAuthVNC is the standard password authentication. See 7.2.2.
type ServerAuthVNC struct {
	Challenge []byte
	Password  []byte
	Crypted   []byte
}

func (*ServerAuthVNC) Type() SecurityType {
	return SecTypeVNC
}
func (*ServerAuthVNC) SubType() SecuritySubType {
	return SecSubTypeUnknown
}

func (auth *ServerAuthVNC) WriteChallenge(c Conn) error {
	if err := binary.Write(c, binary.BigEndian, auth.Challenge); err != nil {
		return err
	}
	return c.Flush()
}

func (auth *ServerAuthVNC) ReadChallenge(c Conn) error {
	var crypted [16]byte
	if err := binary.Read(c, binary.BigEndian, &crypted); err != nil {
		return err
	}
	auth.Crypted = crypted[:]
	return nil
}

func (auth *ServerAuthVNC) Auth(c Conn) error {
	if err := auth.WriteChallenge(c); err != nil {
		return err
	}

	if err := auth.ReadChallenge(c); err != nil {
		return err
	}

	encrypted, err := AuthVNCEncode(auth.Password, auth.Challenge)
	if err != nil {
		return err
	}
	if !bytes.Equal(encrypted, auth.Crypted) {
		return fmt.Errorf("password invalid")
	}
	return nil
}

// ClientAuthVNC is the standard password authentication. See 7.2.2.
type ClientAuthVNC struct {
	Challenge []byte
	Password  []byte
}

func (*ClientAuthVNC) Type() SecurityType {
	return SecTypeVNC
}
func (*ClientAuthVNC) SubType() SecuritySubType {
	return SecSubTypeUnknown
}

func (auth *ClientAuthVNC) Auth(c Conn) error {
	if len(auth.Password) == 0 {
		return fmt.Errorf("Security Handshake failed; no password provided for VNCAuth.")
	}
	var challenge [16]byte
	if err := binary.Read(c, binary.BigEndian, &challenge); err != nil {
		return err
	}

	crypted, err := AuthVNCEncode(auth.Password, challenge[:])
	if err != nil {
		return err
	}

	// Send the encrypted challenge back to server
	if err := binary.Write(c, binary.BigEndian, crypted); err != nil {
		return err
	}

	return c.Flush()
}

func AuthVNCEncode(password []byte, challenge []byte) ([]byte, error) {
	if len(password) > 8 {
		return nil, fmt.Errorf("password too long")
	}
	if len(challenge) != 16 {
		return nil, fmt.Errorf("challenge size not 16 byte long")
	}
	// Copy password string to 8 byte 0-padded slice
	key := make([]byte, 8)
	copy(key, password)

	// Each byte of the password needs to be reversed. This is a
	// non RFC-documented behaviour of VNC clients and servers
	for i := range key {
		key[i] = (key[i]&0x55)<<1 | (key[i]&0xAA)>>1 // Swap adjacent bits
		key[i] = (key[i]&0x33)<<2 | (key[i]&0xCC)>>2 // Swap adjacent pairs
		key[i] = (key[i]&0x0F)<<4 | (key[i]&0xF0)>>4 // Swap the 2 halves
	}

	// Encrypt challenge with key.
	cipher, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(challenge); i += cipher.BlockSize() {
		cipher.Encrypt(challenge[i:i+cipher.BlockSize()], challenge[i:i+cipher.BlockSize()])
	}

	return challenge, nil
}
