package filter

import (
	"net/http"

	"github.com/codingeasygo/web"
)

type FaviconH string

func (f FaviconH) SrvHTTP(s *web.Session) web.Result {
	s.SendFile("favicon.ico", string(f), "image/x-icon")
	return web.Return
}

func (f FaviconH) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	web.SendFile(w, r, "favicon.ico", string(f), "image/x-icon", false)
}
