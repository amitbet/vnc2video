package vnc2video

import "encoding/binary"

func readTightTunnels(c Conn) (uint32, error) {
	var n uint32
	if err := binary.Read(c, binary.BigEndian, &n); err != nil {
		return 0, err
	}
	return n, nil
}

func readTightCaps(c Conn) (int32, []byte, []byte, error) {
	var code int32
	var vendor [4]byte
	var signature [8]byte
	if err := binary.Read(c, binary.BigEndian, &code); err != nil {
		return 0, nil, nil, err
	}
	if err := binary.Read(c, binary.BigEndian, &vendor); err != nil {
		return 0, nil, nil, err
	}
	if err := binary.Read(c, binary.BigEndian, &signature); err != nil {
		return 0, nil, nil, err
	}
	return code, vendor[:], signature[:], nil
}
