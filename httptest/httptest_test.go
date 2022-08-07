package httptest

import (
	"testing"

	"github.com/codingeasygo/util/xmap"
	"github.com/codingeasygo/web"
)

func testHandlerFunc(hs *web.Session) web.Result {
	hs.Printf("%v", hs.Argument("a"))
	return web.Return
}

// func NT(w http.ResponseWriter, r *http.Request) {
// 	fmt.Println(c)
// 	c = c + 1
// }

// type T2 struct {
// }

// func (t *T2) SrvHTTP(hs *web.Session) web.Result {
// 	fmt.Println(hs.Argument("a"))
// 	fmt.Println(hs.Argument("b"))
// 	fmt.Println(c)
// 	c = c + 1
// 	hs.W.Write([]byte("{\"OK\":1}"))
// 	return routing.HRES_RETURN
// }

// func (t *T2) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	w.Write([]byte("{\"OK\":1}"))
// }

func TestServer(t *testing.T) {
	var err error
	var text string
	ts := NewHandlerFuncServer(testHandlerFunc)
	// ts.StartTLS()
	defer ts.Close()
	text, _, err = ts.GetHeaderText(xmap.M{"xx": 1}, "?a=%v", "testing")
	if err != nil || text != "testing" {
		t.Error(err.Error())
		return
	}
}
