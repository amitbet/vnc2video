package vnc

import (
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
