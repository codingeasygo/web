package filter

// import (
// 	"fmt"

// 	"github.com/Centny/gwf/routing"
// 	"github.com/Centny/gwf/util"
// 	"github.com/codingeasygo/util/converter"
// )

// type S2js struct {
// 	Key   string
// 	Value interface{}
// }

// func NewS2js(key string, vals []string) *S2js {
// 	return &S2js{
// 		Key:  key,
// 		Vals: vals,
// 	}
// }

// func (s *S2js) SrvHTTP(hs *routing.HTTPSession) routing.HResult {
// 	hs.W.Header().Set("Content-Type", "application/javascript;charset=utf-8")
// 	var vals = util.Map{}
// 	for _, val := range s.Vals {
// 		vals[val] = hs.Val(val)
// 	}
// 	fmt.Fprintf(hs.W, "var %v = %v;", s.Key, converter.JSON(vals))
// 	return routing.HRES_RETURN
// }
