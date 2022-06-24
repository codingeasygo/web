package web

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/util/xmap"
)

func init() {
	xhttp.EnableCookie()
}

type AbcXML struct {
	XMLName xml.Name `xml:"abc"`
	A       string   `xml:"a" valid:"a,r|s,l:0;"`
}

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
		mux.HandleFunc("/post/a", func(s *Session) Result {
			var a string
			err := s.ValidFormat(`a,r|s,l:0`, &a)
			if err != nil {
				return s.Printf("%v", err.Error())
			}
			return s.Printf("%v", a)
		})
		mux.HandleFunc("/post/b", func(s *Session) Result {
			var args struct {
				A string `json:"a" valid:"a,r|s,l:0;"`
			}
			err := s.Valid(&args, "#all")
			if err != nil {
				return s.Printf("%v", err.Error())
			}
			return s.Printf("%v", args.A)
		})
		mux.HandleFunc("/post/c", func(s *Session) Result {
			var args struct {
				A string `json:"a" valid:"a,r|s,l:0;"`
			}
			var b string
			err := s.Valid(&args, "#all", `b,r|s,l:0`, &b)
			if err != nil {
				return s.Printf("%v", err.Error())
			}
			return s.Printf("%v", args.A+b)
		})
		mux.HandleFunc("/post/d", func(s *Session) Result {
			var args struct {
				A string `json:"a" valid:"a,r|s,l:0;"`
			}
			_, err := s.RecvValidJSON(&args, "#all")
			if err != nil {
				return s.Printf("%v", err.Error())
			}
			return s.Printf("%v", args.A)
		})
		mux.HandleFunc("/post/e", func(s *Session) Result {
			var args AbcXML
			_, err := s.RecvValideXML(&args, "#all")
			if err != nil {
				return s.Printf("%v", err.Error())
			}
			return s.Printf("%v", args.A)
		})
		text, err = xhttp.PostFormText(xmap.M{"a": "123"}, "%v/post/a", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
		text, err = xhttp.PostFormText(xmap.M{"a": "123"}, "%v/post/b", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
		text, err = xhttp.PostFormText(xmap.M{"a": "12", "b": "3"}, "%v/post/c", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
		text, _, err = xhttp.PostHeaderText(nil, bytes.NewBufferString(converter.JSON(xmap.M{"a": "123"})), "%v/post/d", ts.URL)
		if err != nil || text != "123" {
			t.Errorf("err:%v,text:%v", err, text)
			return
		}
		fmt.Println(converter.XML(AbcXML{A: "123"}))
		text, _, err = xhttp.PostHeaderText(nil, bytes.NewBufferString(converter.XML(AbcXML{A: "123"})), "%v/post/e", ts.URL)
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
