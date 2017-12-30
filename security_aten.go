package vnc2webm

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ClientAuthATEN struct {
	Username []byte
	Password []byte
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

	if len(auth.Username) > definedAuthLen || len(auth.Password) > definedAuthLen {
		return fmt.Errorf("username/password is too long, allowed 0-23")
	}

	nt, err := readTightTunnels(c)
	if err != nil {
		return err
	}
	/*
		fmt.Printf("tunnels %d\n", nt)
		for i := uint32(0); i < nt; i++ {
			code, vendor, signature, err := readTightCaps(c)
			if err != nil {
				return err
			}
			fmt.Printf("code %d vendor %s signature %s\n", code, vendor, signature)
		}
	*/
	if ((nt&0xffff0ff0)>>0 == 0xaff90fb0) || (nt <= 0 || nt > 0x1000000) {
		c.SetProtoVersion("aten1")
		var skip [20]byte
		binary.Read(c, binary.BigEndian, &skip)
		//fmt.Printf("skip %v\n", skip)
	}

	username := make([]byte, definedAuthLen)
	password := make([]byte, definedAuthLen)
	copy(username, auth.Username)
	copy(password, auth.Password)
	challenge := bytes.Join([][]byte{username, password}, []byte(""))
	if err := binary.Write(c, binary.BigEndian, challenge); err != nil {
		return err
	}

	if err := c.Flush(); err != nil {
		return err
	}
	/*

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
	*/
	//var pp [10]byte
	//binary.Read(c, binary.BigEndian, &pp)
	//fmt.Printf("ddd %v\n", pp)
	return nil
}
