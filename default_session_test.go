package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codingeasygo/util/xhttp"
)

func TestDefaultSessionMux(t *testing.T) {
	mux := NewSessionMux("/t")
	mux.HandleFunc("^.*$", func(hs *Session) Result {
		hs.SetValue("abc", "123")
		hs.StrVal("abc")
		hs.SetValue("abc", nil)
		hs.StrVal("abc")
		hs.Flush()
		hs.Printf("%v", hs.ID())
		return Return
	})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux.ServeHTTP(w, r)
	}))
	sid, err := xhttp.GetText("%s/t", ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	if mux.Builder.Find(sid) == nil {
		t.Error("error")
		return
	}
}
