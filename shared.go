package web

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Shared is the shared session mux
var Shared = NewSessionMux("")

// Filter will register the handler to shared
func Filter(pattern string, h Handler) {
	Shared.Filter(pattern, h)
}

// FilterFunc will register the handler to shared
func FilterFunc(pattern string, h HandlerFunc) {
	Filter(pattern, h)
}

// Handle will register the handler to shared
func Handle(pattern string, h Handler) {
	Shared.Handle(pattern, h)
}

// HandleFunc will register the handler to shared
func HandleFunc(pattern string, h HandlerFunc) {
	Handle(pattern, h)
}

// Server is shared http server
var Server *http.Server

// Listener is shared listener
var Listener net.Listener

// ListenAndServe will listen the shared server
func ListenAndServe(addr string) (err error) {
	Server = &http.Server{Handler: Shared}
	if strings.HasPrefix(addr, "/") {
		addrs := strings.SplitN(addr, ",", 2)
		Server.Addr = addr
		Listener, err = net.Listen("unix", addrs[0])
		if err != nil {
			return
		}
		defer Listener.Close()
		var mod uint64
		mod, err = strconv.ParseUint(addrs[1], 8, 32)
		if err != nil {
			return
		}
		os.Chmod(addrs[0], os.FileMode(mod))
	} else {
		Server.Addr = addr
		Listener, err = net.Listen("tcp", addr)
		if err != nil {
			return
		}
		Listener = &tcpKeepAliveListener{TCPListener: Listener.(*net.TCPListener)}
		defer Listener.Close()
	}
	return Server.Serve(Listener)
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

var sigc chan os.Signal

// HandleSignal will handle the kill signal and stop the server
func HandleSignal() error {
	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-sigc
	return Listener.Close()
}
