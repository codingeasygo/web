package handler

import "github.com/codingeasygo/web"

type Map map[string]interface{}

func (m Map) SrvHTTP(s *web.Session) web.Result {
	return s.SendJSON(m)
}
