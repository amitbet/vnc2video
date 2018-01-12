package vnc2video

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

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
