package filter

import (
	"net/url"

	"github.com/codingeasygo/web"
)

func ParseQueryF(s *web.Session) web.Result {
	vals, err := url.ParseQuery(s.R.URL.RawQuery)
	if err == nil {
		s.R.Form = vals
		s.R.PostForm = vals
	}
	return web.Continue
}
