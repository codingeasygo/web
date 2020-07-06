package web

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/util/xmap"
)

func TestFilterHandler(t *testing.T) {
	SetLogLevel(LogLevelDebug)
	mux := NewSessionMux("")
	mux.ShowLog = true
	mux.ShowSlow = 100 * time.Millisecond
	mux.StartMonitor()
	var err error
	var text string
	ts := httptest.NewServer(mux)
	{ //filter and handler
		mux.FilterFunc("/f1/.*", func(s *Session) Result {
			s.SetValue("a", "123")
			return Continue
		})
		mux.HandleFunc("/f1/h1", func(s *Session) Result {
			return s.Printf("%v", s.Value("a"))
		})
		mux.HandleNormalFunc("/f1/h2", func(w http.ResponseWriter, r *http.Request) {
			s := mux.Builder.FindSession(w, r)
			fmt.Fprintf(w, "%v", s.Value("a"))
		})
		text, err = xhttp.GetText("%v/f1/h1", ts.URL)
		if err != nil || text != "123" {
			t.Error(err)
			return
		}
		text, err = xhttp.GetText("%v/f1/h2", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //handler continue
		mux.HandleFunc("/f2/h1", func(s *Session) Result {
			s.SetValue("a", "123")
			return Continue
		})
		mux.HandleFunc("/f2/h1", func(s *Session) Result {
			return s.Printf("%v", s.Value("a"))
		})
		text, err = xhttp.GetText("%v/f2/h1", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //normal handler continue
		mux.HandleMethodNormalFunc("/f2/h2", func(w http.ResponseWriter, r *http.Request) {
			s := mux.Builder.FindSession(w, r)
			s.SetValue("a", "123")
		}, "GET,:"+Continue.String())
		mux.HandleMethodNormalFunc("/f2/h2", func(w http.ResponseWriter, r *http.Request) {
			s := mux.Builder.FindSession(w, r)
			fmt.Fprintf(w, "%v", s.Value("a"))
		}, "GET,:"+Return.String())
		text, err = xhttp.GetText("%v/f2/h2", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //filter return
		mux.FilterFunc("/f3/h1", func(s *Session) Result {
			return s.Printf("%v", "123")
		})
		text, err = xhttp.GetText("%v/f3/h1", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //not method
		mux.FilterMethodFunc("/notm/", func(s *Session) Result {
			return s.Printf("%v", "123")
		}, "POST")
		mux.HandleMethodFunc("/notm/", func(s *Session) Result {
			return s.Printf("%v", "123")
		}, "POST")
		text, err = xhttp.GetText("%v/notm/", ts.URL)
		if err == nil {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //not found
		text, err = xhttp.GetText("%v/not", ts.URL)
		if err == nil {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //info
		mux.HandleFunc("/info/", func(s *Session) Result {
			if mux.RequestSession(s.R) == nil {
				panic(nil)
			}
			s.Host()
			return s.Printf("%v", "ok")
		})
		text, err = xhttp.GetText("%v/info/", ts.URL)
		if err != nil || text != "ok" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //show slow
		mux.HandleFunc("/slow/", func(s *Session) Result {
			time.Sleep(150 * time.Millisecond)
			return s.Printf("%v", "ok")
		})
		text, err = xhttp.GetText("%v/slow/", ts.URL)
		if err != nil || text != "ok" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //redirect
		mux.HandleFunc("/redirect/", func(s *Session) Result {
			return s.Redirect("/abc/")
		})
		mux.HandleFunc("/abc/", func(s *Session) Result {
			return s.Printf("%v", "123")
		})
		text, err = xhttp.GetText("%v/redirect/", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //cookie
		mux.HandleFunc("/cookie/set/", func(s *Session) Result {
			s.SetCookie("a", "123")
			return s.Printf("%v", "ok")
		})
		mux.HandleFunc("/cookie/get/", func(s *Session) Result {
			return s.Printf("%v", s.Cookie("a"))
		})
		text, err = xhttp.GetText("%v/cookie/set/", ts.URL)
		if err != nil || text != "ok" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
		text, err = xhttp.GetText("%v/cookie/get/", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{ //valid
		mux.FilterFunc("/post/", func(s *Session) Result {
			s.R.ParseForm()
			s.R.PostForm = s.R.Form
			s.R.Form = nil
			return Continue
		})
		mux.HandleFunc("/post/", func(s *Session) Result {
			var a string
			err := s.ValidFormat(`a,r|s,l:0`, &a)
			if err != nil {
				return s.Printf("%v", err.Error())
			}
			return s.Printf("%v", a)
		})
		text, err = xhttp.PostFormText(xmap.M{"a": "123"}, "%v/post/", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
	}
	{
		mux.Print()
		mux.State()
	}
}
