package vnc2video

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"os/exec"
	"strconv"
	"strings"

	"github.com/monnand/dhkx"
)

// ClientAuthUltraMsAutoLogon2 is implemented by Ultra VNC when using the MS-Logon II authentication.
// see http://www.uvnc.com/features/authentication.html
type ClientAuthUltraMsAutoLogon2 struct {
	Username []byte
	Password []byte
}

func (*ClientAuthUltraMsAutoLogon2) Type() SecurityType {
	return SecTypeUltraMsAutoLogon2
}

func (*ClientAuthUltraMsAutoLogon2) SubType() SecuritySubType {
	return SecSubTypeUnknown
}

func (auth *ClientAuthUltraMsAutoLogon2) Auth(c Conn) error {
	if len(auth.Username) == 0 {
		return fmt.Errorf("security handshake failed; no username provided for UltraMsAutoLogon2Auth")
	}
	if len(auth.Password) == 0 {
		return fmt.Errorf("security handshake failed; no password provided for UltraMsAutoLogon2Auth")
	}

	fmt.Printf("calling binary.Read generator...\n")

	var generator int64
	if err := binary.Read(c, binary.BigEndian, &generator); err != nil {
		return err
	}

	fmt.Printf("calling binary.Read prime...\n")

	var prime int64
	if err := binary.Read(c, binary.BigEndian, &prime); err != nil {
		return err
	}

	fmt.Printf("calling binary.Read public server key...\n")

	var a [8]byte
	if err := binary.Read(c, binary.BigEndian, &a); err != nil {
		return err
	}

	fmt.Printf("calling authUltraMsAutoLogon2Encode...\n")

	b, encryptedUsername, encryptedPassword, err := authUltraMsAutoLogon2Encode(auth.Username, auth.Password, generator, prime, a[:])
	if err != nil {
		return err
	}

	if err := binary.Write(c, binary.BigEndian, b); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, encryptedUsername); err != nil {
		return err
	}
	if err := binary.Write(c, binary.BigEndian, encryptedPassword); err != nil {
		return err
	}

	fmt.Printf("flushing...\n")

	return c.Flush()
}

func authUltraMsAutoLogon2Encode(username []byte, password []byte, generator int64, prime int64, a []byte) (b []byte, encryptedUsername []byte, encryptedPassword []byte, err error) {
	if len(a) != 8 {
		return nil, nil, nil, fmt.Errorf("server public key size not 8 byte long")
	}

	group := dhkx.CreateGroup(big.NewInt(prime), big.NewInt(generator))

	clientPrivateKey, err := group.GeneratePrivateKey(nil)
	if err != nil {
		return nil, nil, nil, err
	}

	b = make([]byte, 8)
	copyWithLeftPad(b, clientPrivateKey.Bytes()) // b is the client public key.

	serverPublicKey := dhkx.NewPublicKey(a)

	key, err := group.ComputeKey(serverPublicKey, clientPrivateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	s := make([]byte, 8)
	copyWithLeftPad(s, key.Bytes())

	// dump a log line exactly as the ultravnc server shows in its winvnc.log file.
	i := int64(binary.BigEndian.Uint64(a))
	k := int64(binary.BigEndian.Uint64(s))
	fmt.Printf("After DH: g=%d m=%d i=%d key=%d\n", generator, prime, i, k)

	// // Each byte of the password needs to be reversed. This is a
	// // non RFC-documented behaviour of VNC clients and servers
	// for i := range s {
	// 	s[i] = (s[i]&0x55)<<1 | (s[i]&0xAA)>>1 // Swap adjacent bits
	// 	s[i] = (s[i]&0x33)<<2 | (s[i]&0xCC)>>2 // Swap adjacent pairs
	// 	s[i] = (s[i]&0x0F)<<4 | (s[i]&0xF0)>>4 // Swap the 2 halves
	// }

	fmt.Printf("username=%s\n", username)
	fmt.Printf("password=%s\n", password)

	encryptedUsername, err = encrypt(256, username, s)
	if err != nil {
		return nil, nil, nil, err
	}

	encryptedPassword, err = encrypt(64, password, s)
	if err != nil {
		return nil, nil, nil, err
	}

	fmt.Printf("encryptedUsername=%s\n", hex.EncodeToString(encryptedUsername))
	fmt.Printf("encryptedPassword=%s\n", hex.EncodeToString(encryptedPassword))
	fmt.Printf("key=%s\n", hex.EncodeToString(s))

	return
}

func encrypt(cipherTextLength int, plainText []byte, key []byte) ([]byte, error) {
	out, err := exec.Command(
		"./ultra-ms-logon-2-encrypt",
		hex.EncodeToString(key),
		strconv.Itoa(cipherTextLength),
		string(plainText)).Output()
	if err != nil {
		return nil, err
	}
	return hex.DecodeString(strings.TrimSpace(string(out)))

	// XXX so the following code should have worked... but the vnc des
	//     implementation does not seem to be standard... so I had to
	//	   create an external application that uses the same C code as
	//     TightVNC/UltraVNC and that works... any idea why?
	// // create zero-padded slice.
	// cipherText := make([]byte, cipherTextLength)
	// copy(cipherText, plainText)

	// block, err := des.NewCipher(key)
	// if err != nil {
	// 	return nil, err
	// }

	// mode := cipher.NewCBCEncrypter(block, key)
	// mode.CryptBlocks(cipherText, cipherText)

	// return cipherText, nil
}

func copyWithLeftPad(dest, src []byte) {
	numPaddingBytes := len(dest) - len(src)
	for i := 0; i < numPaddingBytes; i++ {
		dest[i] = 0
	}
	copy(dest[numPaddingBytes:], src)
}
