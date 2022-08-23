package httptest

import (
	"testing"

	"github.com/codingeasygo/util/xmap"
	"github.com/codingeasygo/web"
)

func TestServer(t *testing.T) {
	var err error
	var text string
	ts := NewHandlerFuncServer(func(hs *web.Session) web.Result {
		hs.Printf("%v", hs.Argument("a"))
		return web.Return
	})
	// ts.StartTLS()
	defer ts.Close()
	text, _, err = ts.GetHeaderText(xmap.M{"xx": 1}, "?a=%v", "testing")
	if err != nil || text != "testing" {
		t.Error(err.Error())
		return
	}
}

func TestShould(t *testing.T) {
	ts := NewMuxServer()
	ts.Mux.HandleFunc("/ok", func(s *web.Session) web.Result {
		return s.SendJSON(xmap.M{
			"code": 0,
		})
	})
	ts.Should(t, "code", 0).GetMap("/ok")
	ts.Should(t, "code", 0).GetHeaderMap(nil, "/ok")
	ts.Should(t, "code", 0).PostMap(nil, "/ok")
	ts.Should(t, "code", 0).PostTypeMap("application/json", nil, "/ok")
	ts.Should(t, "code", 0).PostHeaderMap(nil, nil, "/ok")
	ts.Should(t, "code", 0).PostJSONMap(xmap.M{}, "/ok")
	ts.Should(t, "code", 0).MethodMap("POST", nil, nil, "/ok")
	ts.Should(t, "code", 0).PostFormMap(nil, "/ok")
	ts.Should(t, "code", 0).PostMultipartMap(nil, nil, "/ok")
	ts.Should(t, "code", 0).UploadMap(nil, "file", "httptest.go", "/ok")
	ts.ShouldError(t).GetMap("/none")
	ts.Should(t).OnlyLog(true).GetMap("/none")
}
