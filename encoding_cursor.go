package vnc2webm

import (
	"encoding/binary"
	"image"
	"image/draw"
)

type CursorPseudoEncoding struct {
	Colors  []Color
	BitMask []byte
	Image   image.Image
}

func (*CursorPseudoEncoding) Supported(Conn) bool {
	return true
}

func (enc *CursorPseudoEncoding) SetTargetImage(img draw.Image) {
	enc.Image = img
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
	/*
			int bytesPerPixel = renderer.getBytesPerPixel();
				int length = rect.width * rect.height * bytesPerPixel;
				if (0 == length)
					return;
				byte[] buffer = ByteBuffer.getInstance().getBuffer(length);
				transport.readBytes(buffer, 0, length);

				StringBuilder sb = new StringBuilder(" ");
				for (int i=0; i<length; ++i) {
					sb.append(Integer.toHexString(buffer[i]&0xff)).append(" ");
				}
				int scanLine = (rect.width + 7) / 8;
				byte[] bitmask = new byte[scanLine * rect.height];
				transport.readBytes(bitmask, 0, bitmask.length);

				sb = new StringBuilder(" ");
		        for (byte aBitmask : bitmask) {
		            sb.append(Integer.toHexString(aBitmask & 0xff)).append(" ");
		        }
				int[] cursorPixels = new int[rect.width * rect.height];
				for (int y = 0; y < rect.height; ++y) {
					for (int x = 0; x < rect.width; ++x) {
						int offset = y * rect.width + x;
						cursorPixels[offset] = isBitSet(bitmask[y * scanLine + x / 8], x % 8) ?
							0xFF000000 | renderer.getPixelColor(buffer, offset * bytesPerPixel) :
							0; // transparent
					}
				}
				renderer.createCursor(cursorPixels, rect);
	*/
	return nil
}

func (enc *CursorPseudoEncoding) Write(c Conn, rect *Rectangle) error {

	return nil
}
