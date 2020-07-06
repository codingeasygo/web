package web

import (
	"net/http/httptest"
	"testing"

	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/util/xmap"
)

func TestSendFile(t *testing.T) {
	mux := NewSessionMux("")
	var err error
	var text string
	ts := httptest.NewServer(mux)
	mux.HandleFunc("/send1", func(s *Session) Result {
		return s.SendFile("send.go", "send.go", "")
	})
	mux.HandleFunc("/send2", func(s *Session) Result {
		return s.SendBinary("send.go", ContentTypePlainText)
	})
	mux.HandleFunc("/send3", func(s *Session) Result {
		return s.SendString("abc", ContentTypePlainText)
	})
	mux.HandleFunc("/send4", func(s *Session) Result {
		return s.SendPlainText("abc")
	})
	mux.HandleFunc("/send5", func(s *Session) Result {
		return s.SendJSON(xmap.M{"data": "abc"})
	})
	mux.HandleFunc("/send6", func(s *Session) Result {
		return s.Printf("%v", "abc")
	})
	mux.HandleFunc("/send_err1", func(s *Session) Result {
		return s.SendBinary("xxsend.go", ContentTypePlainText)
	})
	mux.HandleFunc("/send_err2", func(s *Session) Result {
		return s.SendJSON(xmap.M{"data": TestSendFile})
	})
	text, err = xhttp.GetText("%v/send1", ts.URL)
	if err != nil || len(text) < 1 {
		t.Error(err)
		return
	}
	text, err = xhttp.GetText("%v/send2", ts.URL)
	if err != nil || len(text) < 1 {
		t.Error(err)
		return
	}
	text, err = xhttp.GetText("%v/send3", ts.URL)
	if err != nil || len(text) < 1 {
		t.Error(err)
		return
	}
	text, err = xhttp.GetText("%v/send4", ts.URL)
	if err != nil || len(text) < 1 {
		t.Error(err)
		return
	}
	text, err = xhttp.GetText("%v/send5", ts.URL)
	if err != nil || len(text) < 1 {
		t.Error(err)
		return
	}
	text, err = xhttp.GetText("%v/send6", ts.URL)
	if err != nil || len(text) < 1 {
		t.Error(err)
		return
	}
	//
	//error
	_, err = xhttp.GetText("%v/send_err1", ts.URL)
	if err == nil {
		t.Error(err)
		return
	}
	_, err = xhttp.GetText("%v/send_err2", ts.URL)
	if err == nil {
		t.Error(err)
		return
	}
}
