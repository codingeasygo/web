package web

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/util/xmap"
)

func TestRecvMultipart(t *testing.T) {
	os.Remove("/tmp/recv.go")
	mux := NewSessionMux("")
	var err error
	var res xmap.Valuable
	ts := httptest.NewServer(mux)
	//
	{ //recv file
		mux.HandleFunc("/file1", func(s *Session) Result {
			vals, err := s.RecvFile(true, true, "file", "/tmp/")
			if err != nil {
				return s.SendJSON(xmap.M{"code": -1, "err": err.Error()})
			}
			file := vals.Files[0]
			return s.SendJSON(xmap.M{
				"code": 0,
				"a":    string(vals.Values["a"][0]),
				"sha1": fmt.Sprintf("%x", file.SHA1),
				"md5":  fmt.Sprintf("%x", file.MD5),
			})
		})
		mux.HandleFunc("/file2", func(s *Session) Result {
			_, err := s.RecvMultipart(true, true, func(*multipart.Part) (filename string, mode os.FileMode, external []io.Writer, err error) {
				filename = "/xx/tmp/"
				mode = os.ModePerm
				return
			})
			if err == nil {
				return s.SendJSON(xmap.M{"code": -1, "err": "nil"})
			}
			return s.SendJSON(xmap.M{"code": 0, "err": err.Error()})
		})
		res, err = xhttp.UploadMap(xmap.M{"a": "abc"}, "file", "recv.go", "%v/file1", ts.URL)
		if err != nil || res.IntDef(-1, "code") != 0 || res.Str("a") != "abc" {
			t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
			return
		}
		//
		res, err = xhttp.UploadMap(xmap.M{"a": "abc"}, "xx", "recv.go", "%v/file1", ts.URL)
		if err != nil || res.IntDef(-1, "code") == 0 {
			t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
			return
		}
		//
		res, err = xhttp.UploadMap(xmap.M{"a": "abc"}, "xx", "recv.go", "%v/file2", ts.URL)
		if err != nil || res.IntDef(-1, "code") != 0 {
			t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
			return
		}
		fmt.Printf("err:%v,res:%v\n", err, converter.JSON(res.Raw()))
		//
		res, err = xhttp.PostMultipartMap(nil, xmap.M{"a": "abc"}, "%v/file1", ts.URL)
		if err != nil || res.IntDef(-1, "code") == 0 {
			t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
			return
		}
		fmt.Printf("err:%v,res:%v\n", err, converter.JSON(res.Raw()))
		//
		res, err = xhttp.PostFormMap(xmap.M{"a": "abc"}, "%v/file1", ts.URL)
		if err != nil || res.IntDef(-1, "code") == 0 {
			t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
			return
		}
		fmt.Printf("err:%v,res:%v\n", err, converter.JSON(res.Raw()))
	}
	{ //recv file bytes
		mux.HandleFunc("/file3", func(s *Session) Result {
			_, err := s.RecvFileBytes("file", 102400)
			if err == nil {
				return s.SendJSON(xmap.M{"code": 0, "err": "nil"})
			}
			return s.SendJSON(xmap.M{"code": 1, "err": err.Error()})
		})
		mux.HandleFunc("/file4", func(s *Session) Result {
			_, err := s.RecvFileBytes("file", 10)
			if err == nil {
				return s.SendJSON(xmap.M{"code": -1, "err": "nil"})
			}
			return s.SendJSON(xmap.M{"code": 0, "err": err.Error()})
		})
		res, err = xhttp.UploadMap(xmap.M{"a": "abc"}, "file", "recv.go", "%v/file3", ts.URL)
		if err != nil || res.IntDef(-1, "code") != 0 {
			t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
			return
		}
		fmt.Printf("err:%v,res:%v\n", err, converter.JSON(res.Raw()))
		//
		res, err = xhttp.UploadMap(xmap.M{"a": "abc"}, "file", "recv.go", "%v/file4", ts.URL)
		if err != nil || res.IntDef(-1, "code") != 0 {
			t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
			return
		}
		fmt.Printf("err:%v,res:%v\n", err, converter.JSON(res.Raw()))
		//
		res, err = xhttp.UploadMap(xmap.M{"a": "abc"}, "xxx", "recv.go", "%v/file3", ts.URL)
		if err != nil || res.IntDef(-1, "code") == 0 {
			t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
			return
		}
		fmt.Printf("err:%v,res:%v\n", err, converter.JSON(res.Raw()))
	}
	{
		mux.HandleFunc("/file5", func(s *Session) Result {
			s.FormFileInfo("file")
			_, err := s.RecvFileBytes("file", 10)
			if err == nil {
				return s.SendJSON(xmap.M{"code": -1, "err": "nil"})
			}
			return s.SendJSON(xmap.M{"code": 0, "err": err.Error()})
		})
		res, err = xhttp.UploadMap(xmap.M{"a": "abc"}, "file", "recv.go", "%v/file5", ts.URL)
		if err != nil || res.IntDef(-1, "code") != 0 {
			t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
			return
		}
		fmt.Printf("err:%v,res:%v\n", err, converter.JSON(res.Raw()))
	}
}

type xmlObj struct {
}

func TestRecvBody(t *testing.T) {
	mux := NewSessionMux("")
	var err error
	var res xmap.Valuable
	ts := httptest.NewServer(mux)
	mux.HandleFunc("/bytes", func(s *Session) Result {
		bys, err := s.RecvBody()
		if err != nil {
			return s.SendJSON(xmap.M{"code": 1, "err": err.Error()})
		}
		return s.SendJSON(xmap.M{
			"code": 0,
			"len":  len(bys),
		})
	})
	mux.HandleFunc("/json", func(s *Session) Result {
		m := xmap.M{}
		bys, err := s.RecvJSON(&m)
		if err != nil {
			return s.SendJSON(xmap.M{"code": 1, "err": err.Error()})
		}
		return s.SendJSON(xmap.M{
			"code": 0,
			"len":  len(bys),
		})
	})
	mux.HandleFunc("/xml", func(s *Session) Result {
		m := &xmlObj{}
		bys, err := s.RecvXML(&m)
		if err != nil {
			return s.SendJSON(xmap.M{"code": 1, "err": err.Error()})
		}
		return s.SendJSON(xmap.M{
			"code": 0,
			"len":  len(bys),
		})
	})
	res, err = xhttp.PostJSONMap(xmap.M{}, "%v/bytes", ts.URL)
	if err != nil || res.IntDef(-1, "code") != 0 {
		t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
		return
	}
	res, err = xhttp.PostJSONMap(xmap.M{}, "%v/json", ts.URL)
	if err != nil || res.IntDef(-1, "code") != 0 {
		t.Errorf("err:%v,res:%v", err, converter.JSON(res.Raw()))
		return
	}
	text, err := xhttp.PostXMLText(xmlObj{}, "%v/xml", ts.URL)
	if err != nil || len(text) < 1 {
		t.Errorf("err:%v,res:%v", err, err)
		return
	}
}

type fsstate struct {
}

func (f *fsstate) Stat() (info os.FileInfo, err error) {
	return f, nil
}

func (f *fsstate) Name() string {
	return "xx"
}
func (f *fsstate) Size() int64 {
	return 1
}
func (f *fsstate) Mode() os.FileMode {
	return os.ModePerm
}
func (f *fsstate) ModTime() time.Time {
	return time.Now()
}
func (f *fsstate) IsDir() bool {
	return false
}
func (f *fsstate) Sys() interface{} {
	return nil
}

func TestFormFileSzie(t *testing.T) {
	if formFileSzie(&fsstate{}) != 1 {
		t.Error("error")
	}
}
