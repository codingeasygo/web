package filter

import (
	"time"

	"github.com/codingeasygo/web"
)

//NoCacheF will set the not cache to reponse
func NoCacheF(hs *web.Session) web.Result {
	hs.W.Header().Set("Expires", "Tue, 01 Jan 1980 1:00:00 GMT")
	hs.W.Header().Set("Last-Modified", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT"))
	hs.W.Header().Set("Cache-Control", "no-stroe,no-cache,must-revalidate,post-check=0,pre-check=0")
	hs.W.Header().Set("Pragma", "no-cache")
	return web.Continue
}
