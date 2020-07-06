package filter

import (
	"net/http/httptest"
	"testing"

	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/web"
)

func TestFavicon(t *testing.T) {
	ico := FaviconH("favicon.ico")
	mux := web.NewSessionMux("")
	mux.Handle("^.*$", ico)
	ts := httptest.NewServer(mux)
	xhttp.GetBytes("%v", ts.URL)
	ts2 := httptest.NewServer(ico)
	xhttp.GetBytes("%v", ts2.URL)
}
