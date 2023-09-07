package filter

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xhttp"
	"github.com/codingeasygo/web"
)

var regExternLine = regexp.MustCompile(`^<!--[\s]*R:.*-->$`)

// RenderHandler is render handler
type RenderHandler interface {
	LoadData(r *Render, hs *web.Session) (tmpl *Template, data interface{}, err error)
}

// RenderDataHandler is render data handler
type RenderDataHandler interface {
	LoadData(r *Render, hs *web.Session, tmpl *Template, args url.Values, info interface{}) (data interface{}, err error)
}

// RenderDataHandlerFunc is reander data handler by func
type RenderDataHandlerFunc func(r *Render, hs *web.Session, tmpl *Template, args url.Values, info interface{}) (data interface{}, err error)

// LoadData is implement RenderDataHandler
func (f RenderDataHandlerFunc) LoadData(r *Render, hs *web.Session, tmpl *Template, args url.Values, info interface{}) (data interface{}, err error) {
	return f(r, hs, tmpl, args, info)
}

// RenderWebData is render data handler by upstream url
type RenderWebData struct {
	Upstream string //the upstream uri
	Path     string //the path of reponsed map value
}

// NewRenderWebData will create render data by upstream url
func NewRenderWebData(upstream string) *RenderWebData {
	return &RenderWebData{Upstream: upstream}
}

// LoadData will load data from web url
func (r *RenderWebData) LoadData(render *Render, hs *web.Session, tmpl *Template, args url.Values, info interface{}) (data interface{}, err error) {
	var url string
	if strings.Contains(r.Upstream, "?") {
		url = r.Upstream + "&" + args.Encode()
	} else {
		url = r.Upstream + "?" + args.Encode()
	}
	res, err := xhttp.GetMap("%v", url)
	if err == nil {
		if len(r.Path) > 0 {
			data, err = res.ValueVal(r.Path)
		} else {
			data = res
		}
	}
	if err != nil {
		err = fmt.Errorf("RenderWebData do request by url(%v) fail with error(%v)->%v", url, err, converter.JSON(res))
	}
	return data, err
}

// RenderDataNamedHandler is named reader data handler
type RenderDataNamedHandler struct {
	handler map[string]RenderDataHandler
}

// NewRenderDataNamedHandler will return new RenderDataNamedHandler
func NewRenderDataNamedHandler() *RenderDataNamedHandler {
	return &RenderDataNamedHandler{
		handler: map[string]RenderDataHandler{},
	}
}

// AddFunc will register func handler
func (r *RenderDataNamedHandler) AddFunc(key string, f RenderDataHandlerFunc) {
	r.handler[key] = f
}

// AddHandler will register handler
func (r *RenderDataNamedHandler) AddHandler(key string, h RenderDataHandler) {
	r.handler[key] = h
}

// LoadData is implement Handler
func (r *RenderDataNamedHandler) LoadData(reander *Render, hs *web.Session) (tmpl *Template, data interface{}, err error) {
	var args url.Values
	tmpl, args, err = reander.LoadSession(hs)
	if err != nil {
		return
	}
	dataf, ok := r.handler[tmpl.Key]
	if !ok {
		err = fmt.Errorf("the data provider by key(%v) is not found", tmpl.Key)
		return
	}
	data, err = dataf.LoadData(reander, hs, tmpl, args, nil)
	if err != nil {
		err = fmt.Errorf("load provider(%v) data by args(%v) fail with error->%v", tmpl.Key, args.Encode(), err)
	}
	return
}

// Template is reander template
type Template struct {
	Path     string             `json:"path"`
	Text     string             `json:"text"`
	Key      string             `json:"key"`
	URL      *url.URL           `json:"-"`
	Template *template.Template `json:"-"`
}

// Render is http web page render on server
type Render struct {
	Dir       string
	Handler   RenderHandler
	ErrorPage string
	Funcs     template.FuncMap
	CacheErr  bool
	CacheDir  string
	latest    map[string][]byte
	cacheLck  sync.RWMutex
}

// NewRender will create reander by handler
func NewRender(dir string, h RenderHandler) *Render {
	return &Render{
		Dir:      dir,
		Handler:  h,
		CacheErr: true,
		CacheDir: os.TempDir(),
		latest:   map[string][]byte{},
		cacheLck: sync.RWMutex{},
	}
}

// LoadTemplate will create load template from path
func (r *Render) LoadTemplate(path string) (tmpl *Template, err error) {
	tmpl = &Template{}
	tmpl.Path = path
	filename := filepath.Join(r.Dir, path)
	bys, err := ioutil.ReadFile(filename)
	if err != nil {
		err = fmt.Errorf("read template file(%v) fail with error->%v", filename, err)
		return nil, err
	}
	tmpl.Text = string(bys)
	ext := strings.SplitN(tmpl.Text, "\n", 2)[0]
	ext = strings.TrimSpace(ext)
	if regExternLine.MatchString(ext) {
		ext = strings.TrimPrefix(ext, "<!--")
		ext = strings.TrimSuffix(ext, "-->")
		ext = strings.TrimSpace(ext)
		ext = strings.TrimPrefix(ext, "R:")
		tmpl.URL, err = url.Parse(ext)
		if err != nil {
			err = fmt.Errorf("parsing extern line(%v) on file(%v) fail with error->%v", ext, filename, err)
			return nil, err
		}
		tmpl.Key = tmpl.URL.Path
	}
	stdtmpl := template.New(tmpl.Path)
	if r.Funcs != nil {
		stdtmpl = stdtmpl.Funcs(r.Funcs)
	}
	tmpl.Template, err = stdtmpl.Parse(tmpl.Text)
	return
}

// LoadSession will load template by http session.
func (r *Render) LoadSession(hs *web.Session) (*Template, url.Values, error) {
	path := strings.TrimSpace(hs.R.URL.Path)
	path = strings.Trim(path, "/ \t")
	if len(path) < 1 {
		path = "index.html"
	}
	tmpl, err := r.LoadTemplate(path)
	if err != nil {
		return nil, nil, fmt.Errorf("loading template fail->%v", err)
	}
	targs := hs.R.URL.Query()
	if tmpl.URL != nil {
		for key, vals := range tmpl.URL.Query() {
			targs[key] = vals
		}
	}
	return tmpl, targs, nil
}

// SrvHTTP is implement for web.Handler
func (r *Render) SrvHTTP(hs *web.Session) web.Result {
	web.DebugLog("Render doing %v", hs.R.URL.Path)
	if hs.R.URL.Query().Get("_data_") == "1" {
		_, data, err := r.Handler.LoadData(r, hs)
		if err == nil {
			return hs.SendJSON(data)
		}
		return hs.Printf("load data fail with %v", err)
	}
	buffer := bytes.NewBuffer(nil)
	_, _, err := r.prepareResponseData(buffer, hs)
	if err == nil {
		cache := buffer.Bytes()
		hs.W.Write(cache)
		err = r.storeCacheData(hs, cache)
		if err != nil {
			web.ErrorLog("Render store cache data fail with %v", err)
		}
	} else {
		web.ErrorLog("Render prepare response data fail with %v", err)
		cache, lerr := r.loadCacheData(hs)
		if lerr == nil && len(cache) > 0 {
			web.ErrorLog("Render prepare response data fail with %v, and using cache(%v)", err, len(cache))
			hs.W.Write(cache)
		} else if len(r.ErrorPage) > 0 {
			hs.SendBinary(filepath.Join(r.Dir, r.ErrorPage), "text/html")
		} else {
			web.ErrorLog("Render prepare response data fail with %v, and load cache fail with len(%v),%v", err, len(cache), lerr)
			hs.Printf("Render prepare response data fail with %v, and load cache fail with len(%v),%v", err, len(cache), lerr)
		}
	}
	return web.Return
}

func (r *Render) prepareResponseData(w io.Writer, hs *web.Session) (tmpl *Template, data interface{}, err error) {
	tmpl, data, err = r.Handler.LoadData(r, hs)
	if err == nil {
		err = tmpl.Template.Execute(w, data)
	}
	return
}

func (r *Render) cacheFilename(hs *web.Session) (name string) {
	path := strings.TrimSpace(hs.R.URL.Path)
	path = strings.Trim(path, "/ \t")
	if len(path) < 1 {
		path = "index.html"
	}
	name = strings.Replace(strings.Replace(path, "/", "_", -1), "\\", "_", -1) + ".cache"
	return
}

func (r *Render) loadCacheData(hs *web.Session) (cache []byte, err error) {
	if !r.CacheErr {
		return
	}
	filename := r.cacheFilename(hs)
	r.cacheLck.Lock()
	defer r.cacheLck.Unlock()
	cache = r.latest[filename]
	if len(cache) > 0 {
		return
	}
	if len(r.CacheDir) > 0 {
		cacheFile := filepath.Join(r.CacheDir, r.cacheFilename(hs))
		cache, err = ioutil.ReadFile(cacheFile)
		if err == nil {
			r.latest[filename] = cache
			web.DebugLog("Render read cache from %v success", cacheFile)
		}
	}
	return
}

func (r *Render) storeCacheData(hs *web.Session, cache []byte) (err error) {
	if !r.CacheErr {
		return
	}
	filename := r.cacheFilename(hs)
	r.cacheLck.Lock()
	defer r.cacheLck.Unlock()
	having := r.latest[filename]
	if len(having) > 0 && bytes.Equal(having, cache) {
		return
	}
	r.latest[filename] = cache
	if len(r.CacheDir) > 0 {
		cacheFile := filepath.Join(r.CacheDir, r.cacheFilename(hs))
		web.DebugLog("Redner saving cache to file %v", cacheFile)
		err = os.Remove(cacheFile)
		if err == nil || os.IsNotExist(err) {
			err = ioutil.WriteFile(cacheFile, cache, os.ModePerm)
		}
	}
	return
}
