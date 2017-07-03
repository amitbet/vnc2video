package vnc

import "encoding/binary"

type CursorPseudoEncoding struct {
	Colors  []Color
	BitMask []byte
}

func (*CursorPseudoEncoding) Type() EncodingType { return EncCursorPseudo }

func (enc *CursorPseudoEncoding) Read(c Conn, rect *Rectangle) error {
	rgba := make([]byte, int(rect.Height)*int(rect.Width)*int(c.PixelFormat().BPP/8))

	if err := binary.Read(c, binary.BigEndian, &rgba); err != nil {
		return err
	}

	bitmask := make([]byte, int((rect.Width+7)/8*rect.Height))
	if err := binary.Read(c, binary.BigEndian, &bitmask); err != nil {
		return err
	}
	/*
		rectStride := 4 * rect.Width
		for i := uint16(0); i < rect.Height; i++ {
			for j := uint16(0); j < rect.Width; j += 8 {
				for idx, k := j/8, 7; k >= 0; k-- {
					if (bitmask[idx] & (1 << uint(k))) == 0 {
						pIdx := j*4 + i*rectStride
						rgbaBuffer[pIdx] = 0
						rgbaBuffer[pIdx+1] = 0
						rgbaBuffer[pIdx+2] = 0
						rgbaBuffer[pIdx+3] = 0
					}
				}
			}
		}
	*/
	return nil
}

func (enc *CursorPseudoEncoding) Write(c Conn, rect *Rectangle) error {

	return nil
}
