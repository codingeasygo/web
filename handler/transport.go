package handler

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/codingeasygo/util/xnet"
	"github.com/codingeasygo/util/xprop"
	"github.com/codingeasygo/web"
	"golang.org/x/net/websocket"
)

type TransportProxyH struct {
	Remote      string
	Username    string
	Password    string
	server      *websocket.Server
	transporter xnet.Transporter
}

func NewTransportProxyH(remote string) (proxy *TransportProxyH, err error) {
	var transporter xnet.Transporter
	if strings.HasPrefix(remote, "tcp://") {
		transporter = xnet.RawDialerF(net.Dial)
	} else if strings.HasPrefix(remote, "ws://") || strings.HasPrefix(remote, "wss://") {
		transporter = xnet.NewWebsocketDialer()
	} else {
		err = fmt.Errorf("not supported remote %v", remote)
		return
	}
	proxy = &TransportProxyH{
		Remote:      remote,
		transporter: transporter,
	}
	proxy.server = &websocket.Server{Handler: websocket.Handler(proxy.wsHandler)}
	return
}

func (t *TransportProxyH) SrvHTTP(w *web.Session) web.Result {
	if len(t.Username) > 0 {
		havingUsername, havingPassword, ok := w.R.BasicAuth()
		if !ok || t.Username != havingUsername || t.Password != havingPassword {
			web.WarnLog("TransportServerH check basic auth fail with expect(%v:%v),having(%v:%v)", t.Username, t.Password, havingUsername, havingPassword)
			w.W.WriteHeader(401)
			return w.SendPlainText("not acccess")
		}
	}
	t.server.ServeHTTP(w.W, w.R)
	return web.Return
}

func (t *TransportProxyH) wsHandler(ws *websocket.Conn) {
	web.InfoLog("TransportServerH start forward %v to %v", ws.Request().RemoteAddr, t.Remote)
	err := t.transporter.Transport(ws, t.Remote)
	web.InfoLog("TransportServerH forward %v to %v is stopped by %v", ws.Request().RemoteAddr, t.Remote, err)
}

func TransportProxy(conf *xprop.Config, mux *web.SessionMux) {
	enabled := conf.StrDef("0", "/transport/enabled")
	if enabled != "1" {
		return
	}
	username := conf.StrDef("", "/transport/username")
	password := conf.StrDef("", "/transport/password")
	prefix := conf.StrDef("transport", "/transport/prefix")
	conf.Range("transport", func(key string, val interface{}) {
		if key == "enabled" || key == "prefix" || key == "username" || key == "password" {
			return
		}
		proxy, err := NewTransportProxyH(val.(string))
		if err != nil {
			web.ErrorLog("Transport create transport proxy fail with %v=%v", key, val)
			return
		}
		proxy.Username = username
		proxy.Password = password
		pattern := fmt.Sprintf(`^/%v/%v(\?.*)?$`, prefix, key)
		mux.Handle(pattern, proxy)
		web.InfoLog("Transport start transport proxy on %v => %v", pattern, val)
	})
}

type TransportForwardListener struct {
	Local       string
	Remote      string
	listener    net.Listener
	connAll     map[string]net.Conn
	connLock    sync.RWMutex
	waiter      sync.WaitGroup
	transporter xnet.Transporter
}

func NewTransportForwardListener(local, remote string) (forward *TransportForwardListener) {
	forward = &TransportForwardListener{
		Local:       local,
		Remote:      remote,
		connAll:     map[string]net.Conn{},
		connLock:    sync.RWMutex{},
		waiter:      sync.WaitGroup{},
		transporter: xnet.NewWebsocketDialer(),
	}
	return
}

func (t *TransportForwardListener) Serve() (err error) {
	if !strings.HasPrefix(t.Remote, "ws://") && !strings.HasPrefix(t.Remote, "wss://") {
		err = fmt.Errorf("not supported remote %v", t.Remote)
		web.ErrorLog("mapping %v to %v fail with %v", t.Local, t.Remote, err)
		return
	}
	web.InfoLog("TransportForwardH start mapping %v to %v", t.Local, t.Remote)
	t.listener, err = net.Listen("tcp", strings.TrimPrefix(t.Local, "tcp://"))
	if err != nil {
		web.ErrorLog("mapping %v to %v fail with %v", t.Local, t.Remote, err)
		return
	}
	var conn net.Conn
	for {
		conn, err = t.listener.Accept()
		if err != nil {
			break
		}
		t.connLock.Lock()
		t.connAll[fmt.Sprintf("%p", conn)] = conn
		t.connLock.Unlock()
		t.waiter.Add(1)
		go t.proceForward(conn)
	}
	t.connLock.Lock()
	for _, conn := range t.connAll {
		conn.Close()
	}
	t.connLock.Unlock()
	t.waiter.Wait()
	web.InfoLog("mapping %v to %v is done with %v", t.Local, t.Remote, err)
	return
}

func (t *TransportForwardListener) proceForward(conn net.Conn) {
	defer func() {
		t.connLock.Lock()
		delete(t.connAll, fmt.Sprintf("%p", conn))
		t.connLock.Unlock()
		t.waiter.Done()
	}()
	web.InfoLog("start transport %v to %v", conn.RemoteAddr(), t.Remote)
	xerr := t.transporter.Transport(conn, t.Remote)
	web.InfoLog("stop transport %v to %v by %v", conn.RemoteAddr(), t.Remote, xerr)
}

func (t *TransportForwardListener) Close() (err error) {
	if t.listener != nil {
		t.listener.Close()
		t.listener = nil
	}
	return
}

type TransportForwardH struct {
	waiter      sync.WaitGroup
	forwardAll  map[string]*TransportForwardListener
	forwardLock sync.RWMutex
}

func NewTransportForwardH() (forward *TransportForwardH) {
	forward = &TransportForwardH{
		waiter:      sync.WaitGroup{},
		forwardAll:  map[string]*TransportForwardListener{},
		forwardLock: sync.RWMutex{},
	}
	return
}

func (t *TransportForwardH) Start(conf *xprop.Config) (err error) {
	enabled := conf.StrDef("0", "/transport/enabled")
	if enabled != "1" {
		return
	}
	server := conf.StrDef("", "/transport/server")
	if len(server) < 1 {
		err = fmt.Errorf("/transport/server is required")
		return
	}
	conf.Range("transport", func(key string, val interface{}) {
		if key == "enabled" || key == "server" {
			return
		}
		remote := fmt.Sprintf("%v/%v", server, key)
		forward := NewTransportForwardListener(val.(string), remote)
		web.InfoLog("TransportForwardH start transport forward on %v => %v", forward.Local, forward.Remote)
		t.forwardLock.Lock()
		t.forwardAll[fmt.Sprintf("%p", forward)] = forward
		t.forwardLock.Unlock()
		t.waiter.Add(1)
		go t.proceForward(forward)
	})
	return
}

func (t *TransportForwardH) proceForward(forward *TransportForwardListener) {
	defer func() {
		t.forwardLock.Lock()
		delete(t.forwardAll, fmt.Sprintf("%p", forward))
		t.forwardLock.Unlock()
		t.waiter.Done()
	}()
	forward.Serve()
}

func (t *TransportForwardH) Stop() (err error) {
	t.forwardLock.Lock()
	for _, forward := range t.forwardAll {
		forward.Close()
	}
	t.forwardLock.Unlock()
	t.waiter.Wait()
	return
}

func TransportForward(conf *xprop.Config) (forward *TransportForwardH, err error) {
	forward = NewTransportForwardH()
	err = forward.Start(conf)
	return
}
