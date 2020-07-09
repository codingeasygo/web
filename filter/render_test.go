package filter

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/codingeasygo/util/xmap"
	"github.com/codingeasygo/util/xos"
	"github.com/codingeasygo/web/httptest"

	"github.com/codingeasygo/web"
)

func assertGet(ts *httptest.Server, expect string, trim bool, f string, args ...interface{}) {
	data, err := ts.GetText(f, args...)
	if err != nil {
		panic(err)
	}
	if trim {
		data = strings.Trim(data, "\r\n\t ")
	}
	if data != expect {
		panic(fmt.Sprintf("expect %v, but %v", []byte(expect), []byte(data)))
	}
}

func assertGetLike(ts *httptest.Server, expect string, f string, args ...interface{}) {
	data, err := ts.GetText(f, args...)
	if err != nil {
		panic(err)
	}
	if !strings.Contains(data, expect) {
		panic(fmt.Sprintf("expect %v, but %v", expect, data))
	}
}

func TestReander(t *testing.T) {
	webTS := httptest.NewHandlerFuncServer(func(hs *web.Session) web.Result {
		var vals []xmap.M
		json.Unmarshal([]byte(hs.Argument("keys")), &vals)
		return hs.SendJSON(xmap.M{
			"data": xmap.M{
				"name": vals,
			},
		})
	})
	web.SetLogLevel(web.LogLevelDebug)
	xos.Run("rm -rf " + os.TempDir() + "/render_test*")
	var rn = NewRenderDataNamedHandler()
	var r = NewRender(".", rn)
	var ts = httptest.NewHandlerServer(r)
	var abcVal = xmap.M{"name": "abc"}
	rn.AddFunc("/abc", func(r *Render, hs *web.Session, tmpl *Template, args url.Values, info interface{}) (interface{}, error) {
		return abcVal, nil
	})
	rn.AddFunc("", func(r *Render, hs *web.Session, tmpl *Template, args url.Values, info interface{}) (interface{}, error) {
		return xmap.M{"name": "default"}, nil
	})
	webdata := NewRenderWebData(webTS.URL)
	webdata.Path = "/data"
	rn.AddHandler("/web", webdata)
	assertGet(ts, "abc", true, "/render_test1.html")
	assertGet(ts, "abc", true, "/render_test1.html")
	assertGet(ts, "default", true, "/render_test2.html")
	assertGetLike(ts, "render_test3.html", "/render_test3.html")
	assertGetLike(ts, "fail with loading template", "/render_test4.html")
	assertGetLike(ts, "testa,testb", "/render_test5.html")
	assertGet(ts, `{"name":"abc"}`, true, "/render_test1.html?_data_=1")
	//
	//test cache error
	assertGetLike(ts, "render_test6.html", "/render_test6.html")
	abcVal = xmap.M{"name": []string{"abc"}}
	assertGet(ts, "abc", true, "/render_test6.html")
	//using memory cache
	abcVal = xmap.M{"name": "abc"}
	assertGet(ts, "abc", true, "/render_test6.html")
	//using file cache
	r.latest = map[string][]byte{} //clear cache
	assertGet(ts, "abc", true, "/render_test6.html")
	//
	fmt.Printf("test normal done...\n\n\n")
	//
	//
	r = NewRender(".", rn)
	ts = httptest.NewHandlerServer(r)
	rn.AddFunc("/abc", func(r *Render, hs *web.Session, tmpl *Template, args url.Values, info interface{}) (interface{}, error) {
		return nil, fmt.Errorf("error")
	})
	fmt.Println(ts.GetText("/render_test1.html"))
	fmt.Println(ts.GetText("/render_test2.html"))
	fmt.Println(ts.GetText(""))
	r.ErrorPage = "render_test1.html"
	fmt.Println(ts.GetText(""))
	//
}
