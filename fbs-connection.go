package vnc2video

import (
	"encoding/binary"
	"net"
	"vnc2video/logger"

	"io"
	"time"
)

// Conn represents vnc conection
type FbsConn struct {
	FbsReader

	protocol string
	//c      net.IServerConn
	//config *ClientConfig
	colorMap ColorMap

	// Encodings supported by the client. This should not be modified
	// directly. Instead, SetEncodings should be used.
	encodings []Encoding

	// Height of the frame buffer in pixels, sent from the server.
	fbHeight uint16

	// Width of the frame buffer in pixels, sent from the server.
	fbWidth     uint16
	desktopName string
	// The pixel format associated with the connection. This shouldn't
	// be modified. If you wish to set a new pixel format, use the
	// SetPixelFormat method.
	pixelFormat PixelFormat
}

// func (c *FbsConn) Close() error {
// 	return c.fbs.Close()
// }

// // Read reads data from conn
// func (c *FbsConn) Read(buf []byte) (int, error) {
// 	return c.fbs.Read(buf)
// }

//dummy, no writing to this conn...
func (c *FbsConn) Write(buf []byte) (int, error) {
	return len(buf), nil
}

func (c *FbsConn) Conn() net.Conn {
	return nil
}

func (c *FbsConn) Config() interface{} {
	return nil
}

func (c *FbsConn) Protocol() string {
	return "RFB 003.008"
}
func (c *FbsConn) PixelFormat() PixelFormat {
	return c.pixelFormat
}

func (c *FbsConn) SetPixelFormat(pf PixelFormat) error {
	c.pixelFormat = pf
	return nil
}

func (c *FbsConn) ColorMap() ColorMap                       { return c.colorMap }
func (c *FbsConn) SetColorMap(cm ColorMap)                  { c.colorMap = cm }
func (c *FbsConn) Encodings() []Encoding                    { return c.encodings }
func (c *FbsConn) SetEncodings([]EncodingType) error        { return nil }
func (c *FbsConn) Width() uint16                            { return c.fbWidth }
func (c *FbsConn) Height() uint16                           { return c.fbHeight }
func (c *FbsConn) SetWidth(w uint16)                        { c.fbWidth = w }
func (c *FbsConn) SetHeight(h uint16)                       { c.fbHeight = h }
func (c *FbsConn) DesktopName() []byte                      { return []byte(c.desktopName) }
func (c *FbsConn) SetDesktopName(d []byte)                  { c.desktopName = string(d) }
func (c *FbsConn) Flush() error                             { return nil }
func (c *FbsConn) Wait()                                    {}
func (c *FbsConn) SetProtoVersion(string)                   {}
func (c *FbsConn) SetSecurityHandler(SecurityHandler) error { return nil }
func (c *FbsConn) SecurityHandler() SecurityHandler         { return nil }
func (c *FbsConn) GetEncInstance(typ EncodingType) Encoding     {
	for _, enc := range c.encodings {
		if enc.Type() == typ {
			return enc
		}
	}
	return nil
	}

type VncStreamFileReader interface {
	io.Reader
	CurrentTimestamp() int
	ReadStartSession() (*ServerInit, error)
	CurrentPixelFormat() *PixelFormat
	Encodings() []Encoding
}

type FBSPlayHelper struct {
	Conn *FbsConn
	//Fbs              VncStreamFileReader
	serverMessageMap map[uint8]ServerMessage
	firstSegDone     bool
	startTime        int
}

func NewFbsConn(filename string, encs []Encoding) (*FbsConn, error) {

	fbs, err := NewFbsReader(filename)
	if err != nil {
		logger.Error("failed to open fbs reader:", err)
		return nil, err
	}


	//NewFbsReader("/Users/amitbet/vncRec/recording.rbs")
	initMsg, err := fbs.ReadStartSession()
	if err != nil {
		logger.Error("failed to open read fbs start session:", err)
		return nil, err
	}
	fbsConn := &FbsConn{FbsReader: *fbs}
	fbsConn.encodings = encs
	fbsConn.SetPixelFormat(initMsg.PixelFormat)
	fbsConn.SetHeight(initMsg.FBHeight)
	fbsConn.SetWidth(initMsg.FBWidth)
	fbsConn.SetDesktopName([]byte(initMsg.NameText))

	return fbsConn, nil
}

func NewFBSPlayHelper(r *FbsConn) *FBSPlayHelper {
	h := &FBSPlayHelper{Conn: r}
	h.startTime = int(time.Now().UnixNano() / int64(time.Millisecond))

	h.serverMessageMap = make(map[uint8]ServerMessage)
	h.serverMessageMap[0] = &FramebufferUpdate{}
	h.serverMessageMap[1] = &SetColorMapEntries{}
	h.serverMessageMap[2] = &Bell{}
	h.serverMessageMap[3] = &ServerCutText{}

	return h
}

// func (handler *FBSPlayHelper) Consume(seg *RfbSegment) error {

// 	switch seg.SegmentType {
// 	case SegmentFullyParsedClientMessage:
// 		clientMsg := seg.Message.(ClientMessage)
// 		logger.Debugf("ClientUpdater.Consume:(vnc-server-bound) got ClientMessage type=%s", clientMsg.Type())
// 		switch clientMsg.Type() {

// 		case FramebufferUpdateRequestMsgType:
// 			if !handler.firstSegDone {
// 				handler.firstSegDone = true
// 				handler.startTime = int(time.Now().UnixNano() / int64(time.Millisecond))
// 			}
// 			handler.sendFbsMessage()
// 		}
// 		// server.MsgFramebufferUpdateRequest:
// 	}
// 	return nil
// }

func (h *FBSPlayHelper) ReadFbsMessage() ServerMessage {
	var messageType uint8
	//messages := make(map[uint8]ServerMessage)
	fbs := h.Conn
	//conn := h.Conn
	err := binary.Read(fbs, binary.BigEndian, &messageType)
	if err != nil {
		logger.Error("TestServer.NewConnHandler: Error in reading FBS: ", err)
		return nil
	}
	//IClientConn{}
	//binary.Write(h.Conn, binary.BigEndian, messageType)
	msg := h.serverMessageMap[messageType]
	if msg == nil {
		logger.Error("TestServer.NewConnHandler: Error unknown message type: ", messageType)
		return nil
	}
	//read the actual message data
	//err = binary.Read(fbs, binary.BigEndian, &msg)
	parsedMsg, err := msg.Read(fbs)
	if err != nil {
		logger.Error("TestServer.NewConnHandler: Error in reading FBS message: ", err)
		return nil
	}

	timeSinceStart := int(time.Now().UnixNano()/int64(time.Millisecond)) - h.startTime
	timeToSleep := fbs.CurrentTimestamp() - timeSinceStart
	if timeToSleep > 0 {
		time.Sleep(time.Duration(timeToSleep) * time.Millisecond)
	}

	return parsedMsg
}
