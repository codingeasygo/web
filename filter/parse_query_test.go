package filter

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/web"
)

func testRecv(hs *web.Session) web.Result {
	var a string
	err := hs.ValidFormat(`
		a,R|S,L:0;
		`, &a)
	if err != nil {
		return hs.Printf("%v", err)
	}
	fmt.Println("A->", a)
	_, err = hs.RecvFile(false, false, "file", "abc2.txt")
	if err == nil {
		return hs.Printf("%v", "ok")
	}
	return hs.Printf("%v", err)
}

func TestParseQuery(t *testing.T) {
	mux := web.NewSessionMux("")
	mux.FilterFunc("^.*$", ParseQueryF)
	mux.HandleFunc("^.*$", testRecv)
	ts := httptest.NewServer(mux)
	ioutil.WriteFile("abc.txt", []byte("123"), os.ModePerm)
	text, err := xhttp.UploadText(nil, "file", "abc.txt", "%v?a=1", ts.URL)
	if err != nil || text != "ok" {
		t.Errorf("err:%v,text:%v", err, text)
		return
	}
}
