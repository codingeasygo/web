package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/codingeasygo/web"
)

func main() {
	var listen, single, dir string
	flag.StringVar(&listen, "listen", "127.0.0.1:0", "the listen address")
	flag.StringVar(&single, "single", "", "start single file server")
	flag.StringVar(&dir, "dir", ".", "start dir file server")
	flag.Parse()
	mux := web.NewSessionMux("")
	if len(single) > 0 {
		fmt.Printf("start single file web by %v\n", single)
		mux.HandleFunc("^.*$", func(s *web.Session) web.Result {
			return s.SendFile("", single, "")
		})
	} else if len(dir) > 0 {
		fmt.Printf("start dir web by %v\n", dir)
		mux.HandleNormal("^.*$", http.FileServer(http.Dir(dir)))
	}
	listener, err := net.Listen("tcp", listen)
	if err != nil {
		fmt.Printf("listen on %v fail with %v\n", listen, err)
		os.Exit(1)
		return
	}
	fmt.Printf("listen web on %v\n", listener.Addr())
	server := &http.Server{Handler: mux}
	fmt.Println(server.Serve(listener))
}
