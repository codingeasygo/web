package handler

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/codingeasygo/util/xnet"
	"github.com/codingeasygo/util/xprop"
	"github.com/codingeasygo/web"
	"golang.org/x/net/websocket"
)

const transportProxyConfig = `
[transport]
enabled=1
username=test
password=123
prefix=ss
a=tcp://127.0.0.1:10010
b=ws://127.0.0.1:10020
c=xx
`
const transportFowradConfig = `
[transport]
enabled=1
server=ws://test:123@127.0.0.1:10000/ss
a=tcp://127.0.0.1:10030
b=tcp://127.0.0.1:10040
`

func TestTransport(t *testing.T) {
	tcpBackend, err := net.Listen("tcp", ":10010")
	if err != nil {
		t.Error(err)
		return
	}
	defer tcpBackend.Close()
	go func() {
		for {
			conn, err := tcpBackend.Accept()
			if err != nil {
				break
			}
			go io.Copy(conn, conn)
		}
	}()
	wsBackend := http.Server{
		Addr: ":10020",
		Handler: websocket.Server{
			Handler: func(conn *websocket.Conn) {
				io.Copy(conn, conn)
			},
		},
	}
	defer wsBackend.Close()
	go func() {
		wsBackend.ListenAndServe()
	}()
	proxyConf := xprop.NewConfig()
	proxyConf.LoadPropString(transportProxyConfig)
	proxyMux := web.NewBuilderSessionMux("", web.NewMemSessionBuilder("", "/", "httptest", 60*time.Second))
	proxy := http.Server{Addr: ":10000", Handler: proxyMux}
	defer proxy.Close()
	go func() {
		proxy.ListenAndServe()
	}()
	TransportProxy(proxyConf, proxyMux)

	{ //direct
		dialer := xnet.NewWebsocketDialer()
		buf := make([]byte, 16)

		tcpConn, err := dialer.Dial("ws://test:123@127.0.0.1:10000/ss/a")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Fprintf(tcpConn, "abc")
		n, err := tcpConn.Read(buf)
		if err != nil || string(buf[0:n]) != "abc" {
			t.Error(err)
			return
		}
		tcpConn.Close()

		wsConn, err := dialer.Dial("ws://test:123@127.0.0.1:10000/ss/b")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Fprintf(wsConn, "abc")
		n, err = wsConn.Read(buf)
		if err != nil || string(buf[0:n]) != "abc" {
			t.Error(err)
			return
		}
		wsConn.Close()

		_, err = dialer.Dial("ws://test:xxx@127.0.0.1:10000/ss/b")
		if err == nil {
			t.Error(err)
			return
		}
	}

	{ //forward
		forwardConf := xprop.NewConfig()
		forwardConf.LoadPropString(transportFowradConfig)
		forward, err := TransportForward(forwardConf)
		if err != nil {
			t.Error(err)
			return
		}
		defer forward.Stop()
		buf := make([]byte, 16)
		time.Sleep(500 * time.Millisecond)

		tcpConn, err := net.Dial("tcp", "127.0.0.1:10030")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Fprintf(tcpConn, "abc")
		n, err := tcpConn.Read(buf)
		if err != nil || string(buf[0:n]) != "abc" {
			t.Error(err)
			return
		}
		tcpConn.Close()

		wsConn, err := net.Dial("tcp", "127.0.0.1:10040")
		if err != nil {
			t.Error(err)
			return
		}
		fmt.Fprintf(wsConn, "abc")
		n, err = wsConn.Read(buf)
		if err != nil || string(buf[0:n]) != "abc" {
			t.Error(err)
			return
		}
		wsConn.Close()
	}

	{ //error
		conf := xprop.NewConfig()
		TransportProxy(conf, proxyMux)
		TransportForward(conf)
		conf.SetValue("/transport/enabled", "1")
		TransportForward(conf)

		NewTransportForwardListener("", "xxx://").Serve()
		NewTransportForwardListener("tcp://xx:xx", "ws://localhost").Serve()
	}

}
