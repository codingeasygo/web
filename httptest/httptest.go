package httptest

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/util/xmap"
	"github.com/codingeasygo/web"
)

var Log = log.New(os.Stderr, "    ", log.Llongfile)

type Client struct {
	Client     *xhttp.Client
	shouldT    *testing.T
	shouldErr  bool
	shouldArgs []interface{}
}

func (c *Client) callError(err error) {
	if c.shouldT == nil {
		panic(err)
	} else {
		Log.Output(5, err.Error())
		c.shouldT.Fail()
		c.shouldT.SkipNow()
	}
}

func (c *Client) validError(res xmap.M, err error) bool {
	if err != nil {
		c.callError(fmt.Errorf("%v, res is %v", err, converter.JSON(res)))
		return false
	}
	return true
}

func (c *Client) validShould(res xmap.M, err error) bool {
	if len(c.shouldArgs) < 1 {
		return true
	}
	xerr := res.Should(c.shouldArgs...)
	if xerr != nil {
		c.callError(fmt.Errorf("%v, res is %v", xerr, converter.JSON(res)))
		return false
	}
	return true
}

func (c *Client) validResult(res xmap.M, err error) bool {
	if c.shouldErr {
		if err == nil {
			c.callError(fmt.Errorf("err is nil, res is %v", converter.JSON(res)))
			return false
		}
	} else {
		if !c.validError(res, err) {
			return false
		}
		if !c.validShould(res, err) {
			return false
		}
	}
	return true
}

//Should will assert by xmap.M.Should
func (c *Client) Should(t *testing.T, args ...interface{}) *Client {
	c.shouldT, c.shouldArgs = t, args
	return c
}

//Should will assert by xmap.M.Should
func (c *Client) ShouldError(t *testing.T) *Client {
	c.shouldT, c.shouldErr = t, true
	return c
}

//GetMap will get map from remote
func (c *Client) GetMap(format string, args ...interface{}) (data xmap.M, err error) {
	data, err = c.Client.GetMap(format, args...)
	c.validResult(data, err)
	return
}

//GetHeaderMap will get map from remote
func (c *Client) GetHeaderMap(header xmap.M, format string, args ...interface{}) (data xmap.M, res *http.Response, err error) {
	data, res, err = c.Client.GetHeaderMap(header, format, args...)
	c.validResult(data, err)
	return
}

//PostMap will get map from remote
func (c *Client) PostMap(body io.Reader, format string, args ...interface{}) (data xmap.M, err error) {
	data, err = c.Client.PostMap(body, format, args...)
	c.validResult(data, err)
	return
}

//PostTypeMap will get map from remote
func (c *Client) PostTypeMap(contentType string, body io.Reader, format string, args ...interface{}) (data xmap.M, err error) {
	data, err = c.Client.PostTypeMap(contentType, body, format, args...)
	c.validResult(data, err)
	return
}

//PostHeaderMap will get map from remote
func (c *Client) PostHeaderMap(header xmap.M, body io.Reader, format string, args ...interface{}) (data xmap.M, res *http.Response, err error) {
	data, res, err = c.Client.PostHeaderMap(header, body, format, args...)
	c.validResult(data, err)
	return
}

//PostJSONMap will get map from remote
func (c *Client) PostJSONMap(body interface{}, format string, args ...interface{}) (data xmap.M, err error) {
	data, err = c.Client.PostJSONMap(body, format, args...)
	c.validResult(data, err)
	return
}

//MethodBytes will do http request, read reponse and parse to map
func (c *Client) MethodMap(method string, header xmap.M, body io.Reader, format string, args ...interface{}) (data xmap.M, res *http.Response, err error) {
	data, res, err = c.Client.MethodMap(method, header, body, format, args...)
	c.validResult(data, err)
	return
}

//PostFormMap will get map from remote
func (c *Client) PostFormMap(form xmap.M, format string, args ...interface{}) (data xmap.M, err error) {
	data, err = c.Client.PostFormMap(form, format, args...)
	c.validResult(data, err)
	return
}

//PostMultipartMap will get map from remote
func (c *Client) PostMultipartMap(header, fields xmap.M, format string, args ...interface{}) (data xmap.M, err error) {
	data, err = c.Client.PostMultipartMap(header, fields, format, args...)
	c.validResult(data, err)
	return
}

//UploadMap will get map from remote
func (c *Client) UploadMap(fields xmap.M, filekey, filename, format string, args ...interface{}) (data xmap.M, err error) {
	data, err = c.Client.UploadMap(fields, filekey, filename, format, args...)
	c.validResult(data, err)
	return
}

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

func (s *Server) Should(t *testing.T, args ...interface{}) *Client {
	c := &Client{
		Client: &xhttp.Client{Raw: s.rawRequest},
	}
	return c.Should(t, args...)
}

func (s *Server) ShouldError(t *testing.T) *Client {
	c := &Client{
		Client: &xhttp.Client{Raw: s.rawRequest},
	}
	return c.ShouldError(t)
}
