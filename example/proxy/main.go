package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"strings"
	"sync"
	"time"
	vnc "vnc2video"
	log "github.com/sirupsen/logrus"
)

type Auth struct {
	Username []byte
	Password []byte
}

type Proxy struct {
	cc    vnc.Conn
	conns chan vnc.Conn
	inp   chan vnc.ClientMessage
	out   chan vnc.ServerMessage
}

var (
	cliconns = make(map[string]*Proxy)
	srvconns = make(map[vnc.Conn]string)
	m        sync.Mutex
)

func newConn(hostport string, password []byte) (vnc.Conn, chan vnc.ClientMessage, chan vnc.ServerMessage, chan vnc.Conn, error) {
	fmt.Printf("new conn to %s with %s\n", hostport, password)
	if cc, ok := cliconns[hostport]; ok {
		return cc.cc, cc.inp, cc.out, cc.conns, nil
	}
	c, err := net.DialTimeout("tcp", hostport, 10*time.Second)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	cchServer := make(chan vnc.ServerMessage)
	cchClient := make(chan vnc.ClientMessage)
	errorCh := make(chan error)
	ccfg := &vnc.ClientConfig{
		SecurityHandlers: []vnc.SecurityHandler{&vnc.ClientAuthVNC{Password: password}},
		PixelFormat:      vnc.PixelFormat32bit,
		ClientMessageCh:  cchClient,
		ServerMessageCh:  cchServer,
		//ServerMessages:   vnc.DefaultServerMessages,
		Encodings: []vnc.Encoding{&vnc.RawEncoding{}},
		ErrorCh:   errorCh,
	}
	csrv := make(chan vnc.Conn)
	inp := make(chan vnc.ClientMessage)
	out := make(chan vnc.ServerMessage)
	fmt.Printf("connect to vnc\n")
	cc, err := vnc.Connect(context.Background(), c, ccfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	fmt.Printf("connected to vnc %#+v\n", cc)
	ds := &vnc.DefaultClientMessageHandler{}
	go ds.Handle(cc)
	go handleIO(cc, inp, out, csrv)

	return cc, inp, out, csrv, nil
}

func handleIO(cli vnc.Conn, inp chan vnc.ClientMessage, out chan vnc.ServerMessage, csrv chan vnc.Conn) {
	fmt.Printf("handle io\n")
	ccfg := cli.Config().(*vnc.ClientConfig)
	defer cli.Close()
	var conns []vnc.Conn
	//var prepared bool

	for {
		select {
		case err := <-ccfg.ErrorCh:
			for _, srv := range conns {
				srv.Close()
			}
			fmt.Printf("err %v\n", err)
			return
		case msg := <-ccfg.ServerMessageCh:
			for _, srv := range conns {
				scfg := srv.Config().(*vnc.ServerConfig)
				scfg.ServerMessageCh <- msg
			}
		case msg := <-inp:
			// messages from real clients
			fmt.Printf("3 %#+v\n", msg)
			switch msg.Type() {
			case vnc.SetPixelFormatMsgType:

			case vnc.SetEncodingsMsgType:
				var encTypes []vnc.EncodingType
				encs := []vnc.Encoding{
					//		&vnc.TightPngEncoding{},
					&vnc.CopyRectEncoding{},
					&vnc.RawEncoding{},
				}
				for _, senc := range encs {
					for _, cenc := range msg.(*vnc.SetEncodings).Encodings {
						if cenc == senc.Type() {
							encTypes = append(encTypes, senc.Type())
						}
					}
				}
				ccfg.ClientMessageCh <- &vnc.SetEncodings{Encodings: encTypes}
			default:
				ccfg.ClientMessageCh <- msg
			}
		case msg := <-out:
			fmt.Printf("4 %#+v\n", msg)
		case srv := <-csrv:
			conns = append(conns, srv)
		}

	}

}

type HijackHandler struct{}

func (*HijackHandler) Handle(c vnc.Conn) error {
	m.Lock()
	defer m.Unlock()
	hostport, ok := srvconns[c]
	if !ok {
		return fmt.Errorf("client connect in server pool not found")
	}
	proxy, ok := cliconns[hostport]
	if !ok {
		return fmt.Errorf("client connect to qemu not found")
	}
	cfg := c.Config().(*vnc.ServerConfig)
	cfg.ClientMessageCh = proxy.inp
	cfg.ServerMessageCh = proxy.out

	proxy.conns <- c
	ds := &vnc.DefaultServerMessageHandler{}
	go ds.Handle(c)
	return nil
}

type AuthVNCHTTP struct {
	c *http.Client
	vnc.ServerAuthVNC
}

func (auth *AuthVNCHTTP) Auth(c vnc.Conn) error {
	auth.ServerAuthVNC.Challenge = []byte("clodo.ruclodo.ru")
	if err := auth.ServerAuthVNC.WriteChallenge(c); err != nil {
		return err
	}
	if err := auth.ServerAuthVNC.ReadChallenge(c); err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	enc := base64.NewEncoder(base64.StdEncoding, buf)
	enc.Write(auth.ServerAuthVNC.Crypted)
	enc.Close()

	v := url.Values{}
	v.Set("hash", buf.String())
	buf.Reset()
	src, _, _ := net.SplitHostPort(c.Conn().RemoteAddr().String())
	v.Set("ip", src)
	res, err := auth.c.PostForm("https://api.ix.clodo.ru/system/vnc", v)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 || res.Body == nil {
		if res.Body != nil {
			io.Copy(buf, res.Body)
		}
		fmt.Printf("failed to get auth data: code %d body %s\n", res.StatusCode, buf.String())
		defer buf.Reset()
		return fmt.Errorf("failed to get auth data: code %d body %s", res.StatusCode, buf.String())
	}
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return fmt.Errorf("failed to get auth data: %s", err.Error())
	}
	log.Debugf("http auth: %s\n", buf.Bytes())
	res.Body.Close()
	data := strings.Split(buf.String(), " ")
	if len(data) < 2 {
		return fmt.Errorf("failed to get auth data data invalid")
	}
	buf.Reset()

	hostport := string(data[0])
	password := []byte(data[1])

	m.Lock()
	defer m.Unlock()
	cc, inp, out, conns, err := newConn(hostport, password)
	if err != nil {
		return err
	}
	cliconns[hostport] = &Proxy{cc, conns, inp, out}
	srvconns[c] = hostport
	c.SetWidth(cc.Width())
	c.SetHeight(cc.Height())
	return nil
}

func (*AuthVNCHTTP) Type() vnc.SecurityType {
	return vnc.SecTypeVNC
}

func (*AuthVNCHTTP) SubType() vnc.SecuritySubType {
	return vnc.SecSubTypeUnknown
}

func main() {
	go func() {
		log.Info(http.ListenAndServe(":6060", nil))
	}()

	ln, err := net.Listen("tcp", ":6900")
	if err != nil {
		log.Fatalf("Error listen. %v", err)
	}

	schClient := make(chan vnc.ClientMessage)
	schServer := make(chan vnc.ServerMessage)

	scfg := &vnc.ServerConfig{
		SecurityHandlers: []vnc.SecurityHandler{
			&AuthVNCHTTP{c: &http.Client{}},
		},
		Encodings: []vnc.Encoding{
			//		&vnc.TightPngEncoding{},
			&vnc.CopyRectEncoding{},
			&vnc.RawEncoding{},
		},
		PixelFormat:     vnc.PixelFormat32bit,
		ClientMessageCh: schClient,
		ServerMessageCh: schServer,
		//ClientMessages:  vnc.DefaultClientMessages,
		DesktopName: []byte("vnc proxy"),
	}
	scfg.Handlers = append(scfg.Handlers, vnc.DefaultServerHandlers...)
	scfg.Handlers = append(scfg.Handlers[:len(scfg.Handlers)-1], &HijackHandler{})
	vnc.Serve(context.Background(), ln, scfg)
}
