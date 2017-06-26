package vnc

import "encoding/binary"

func readTightTunnels(c Conn) (uint32, error) {
	var n uint32
	if err := binary.Read(c, binary.BigEndian, &n); err != nil {
		return 0, err
	}
	return n, nil
}
