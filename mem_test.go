package web

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/util/xmap"
)

type CSrv struct {
	Count int
	Res   Result
}

func (s *CSrv) SrvHTTP(hs *Session) Result {
	s.Count = s.Count + 1
	hs.SetValue("abc", "123456789")
	fmt.Println(hs.Host())
	hs.SetValue("kkk", nil)
	fmt.Println(hs.Value("kkk"))
	//
	var iv int64
	err := hs.ValidFormat("int,R|I,R:50~300", &iv)
	fmt.Println(err, iv)
	if iv != 123 {
		panic("hava error")
	}
	hs.Cookie("key")
	hs.SetCookie("kk", "sfsf")
	hs.Cookie("kk")
	hs.SetCookie("kk", "")
	time.Sleep(10 * time.Millisecond)
	return s.Res
}

func TestMemSession(t *testing.T) {
	builder := NewMemSessionBuilder("", "/", "mtest", 50*time.Millisecond)
	builder.ShowLog = true
	builder.delay = 10 * time.Millisecond
	builder.StartTimeout()
	builder.SetEventHandler(SessionEventFunc(func(key string, s Sessionable) {
	}))
	mux := NewBuilderSessionMux("", builder)
	mux.StartMonitor()
	mux.ShowLog = true
	mux.ShowSlow = time.Millisecond
	mux.HandleFunc("/val/.*", func(s *Session) Result {
		var ival int
		var sval string
		err := s.ValidFormat(`
			ival,r|i,r:0;
			sval,r|s,l:0;
		`, &ival, &sval)
		if err != nil {
			return s.SendJSON(xmap.M{"code": -1, "err": err.Error()})
		}
		return s.SendJSON(xmap.M{
			"code": 0,
			"ival": ival,
			"sval": sval,
			"sid":  s.ID(),
		})
	})
	ts := httptest.NewServer(mux)
	res, err := xhttp.GetMap("%v/val/?ival=100&sval=abc", ts.URL)
	if err != nil {
		t.Errorf("err:%v,res:%v", err, res)
		return
	}
	var code, ival int
	var sval, sid string
	err = res.ValidFormat(`
		code,r|i,o:0;
		ival,r|i,o:100;
		sval,r|s,o:abc;
		sid,r|s,l:0;
	`, &code, &ival, &sval, &sid)
	if err != nil {
		t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
		return
	}
	if xx := builder.Find(sid); xx == nil {
		t.Error(xx)
		return
	}
	time.Sleep(100 * time.Millisecond)
	if xx := builder.Find(sid); xx != nil {
		t.Error(xx)
		return
	}
	res, err = xhttp.GetMap("%v/val/?ival=100&sval=abc", ts.URL)
	if err != nil {
		t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
		return
	}
	builder.StopTimeout()
}
