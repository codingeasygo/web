package httptest

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/util/xmap"
	"github.com/codingeasygo/web"
)

//Server is httptest server
type Server struct {
	*xhttp.Client
	URL string
	S   *httptest.Server
	TLS *httptest.Server
	Mux *web.SessionMux
}

//NewServer will return session mux httptest server
func NewServer(mux *web.SessionMux) *Server {
	srv := &Server{Mux: mux}
	srv.S = httptest.NewServer(mux)
	srv.Client = &xhttp.Client{Raw: srv.rawRequest}
	srv.URL = srv.S.URL
	return srv
}

//NewMuxServer will return httptest server
func NewMuxServer() *Server {
	sb := web.NewMemSessionBuilder("", "/", "httptest", 60*time.Second)
	mux := web.NewBuilderSessionMux("", sb)
	return NewServer(mux)
}

//NewHandlerServer will return httptest server for web.Handler
func NewHandlerServer(f web.Handler) *Server {
	srv := NewMuxServer()
	srv.Mux.Handle("^.*$", f)
	return srv
}

//NewHandlerFuncServer will return httptest server for web.HandlerFunc
func NewHandlerFuncServer(f web.HandlerFunc) *Server {
	return NewHandlerServer(f)
}

//Close will close the httptest server
func (s *Server) Close() {
	s.S.Close()
	if s.TLS != nil {
		s.TLS.Close()
	}
}

//StartTLS will enable tls
func (s *Server) StartTLS() {
	s.TLS = httptest.NewTLSServer(s.Mux)
	s.URL = s.TLS.URL
}

func (s *Server) rawRequest(method, uri string, header xmap.M, body io.Reader) (req *http.Request, res *http.Response, err error) {
	remote := fmt.Sprintf("%v%v", s.URL, uri)
	req, err = http.NewRequest(method, remote, body)
	if err == nil {
		for k, v := range header {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
		res, err = xhttp.DefaultClient.Do(req)
	}
	return
}

func (s *Server) Should(t *testing.T, args ...interface{}) *xhttp.ShouldClient {
	c := xhttp.NewShouldClient()
	c.Client = s.Client
	return c.Should(t, args...)
}

func (s *Server) ShouldError(t *testing.T) *xhttp.ShouldClient {
	c := xhttp.NewShouldClient()
	c.Client = s.Client
	return c.ShouldError(t)
}
