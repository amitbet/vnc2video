package vnc2video

import (
	"bytes"
	"crypto/des"
	"encoding/binary"
	"fmt"
)

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

	encrypted, err := AuthVNCEncode(auth.Password, challenge[:])
	if err != nil {
		return err
	}
	// Send the encrypted challenge back to server
	if err := binary.Write(c, binary.BigEndian, encrypted); err != nil {
		return err
	}

	return c.Flush()
}

func AuthVNCEncode(password []byte, challenge []byte) ([]byte, error) {
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
