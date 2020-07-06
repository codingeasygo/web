package filter

import (
	"net/http/httptest"
	"testing"

	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/web"
)

func TestNoCache(t *testing.T) {
	mux := web.NewSessionMux("")
	mux.FilterFunc("^.*$", NoCacheF)
	mux.HandleFunc("^.*$", func(hs *web.Session) web.Result {
		return web.Return
	})
	ts := httptest.NewServer(mux)
	xhttp.GetText("%v", ts.URL)
}
