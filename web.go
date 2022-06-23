package web

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/codingeasygo/util/attrscan"
	"github.com/codingeasygo/util/attrvalid"
	"github.com/codingeasygo/util/monitor"
	"github.com/codingeasygo/util/xmap"
)

//Result is http handler result
type Result int

const (
	//Continue is http handler return for conntinue next handler
	Continue Result = iota
	//Return is http handler return for request is done
	Return
)

// const (
// 	//the hook name
// 	HK_ROUTING = "ROUTING"

// 	//
// 	//filter begin,
// 	//the hook parameter
// 	// val:nil
// 	// args[]:
// 	//  *HTTTPSession	the HTTTPSession
// 	HK_R_BEG = "R_BEG"

// 	//
// 	//filter end,
// 	//the hook parameter
// 	// val: the HTTTPSession.V, it will be converted by SessionMux.FIND_V returned function.
// 	// args[]:
// 	//  *HTTTPSession	the HTTTPSession
// 	//  matched(bool)	if having filter matched.
// 	HK_R_END = "R_END"

// 	//
// 	//filter begin,
// 	//the hook parameter
// 	// val:nil
// 	// args[]:
// 	//  *HTTTPSession	the HTTTPSession
// 	HK_F_BEG = "F_BEG"

// 	//
// 	//filter end,
// 	//the hook parameter
// 	// val:nil
// 	// args[]:
// 	//  *HTTTPSession	the HTTTPSession
// 	//  matched(bool)	if having filter matched.
// 	//  HResult			the execute result.
// 	HK_F_END = "F_END" //filter end

// 	//
// 	//handler begin,
// 	//the hook parameter
// 	// val:nil
// 	// args[]:
// 	//  *HTTTPSession	the HTTTPSession
// 	HK_H_BEG = "H_BEG"

// 	//
// 	//handler end,
// 	//the hook parameter
// 	// val:nil
// 	// args[]:
// 	//  *HTTTPSession	the HTTTPSession
// 	//  matched(bool)	if having filter matched.
// 	//  HResult			the execute result.
// 	HK_H_END = "H_END"
// )

func (h Result) String() string {
	if h == Continue {
		return "CONTINUE"
	}
	return "RETURN"
}

//SessionEventFunc is SessionEventHandler implement
type SessionEventFunc func(string, Sessionable)

//OnCreate is event handler on session create
func (f SessionEventFunc) OnCreate(s Sessionable) {
	f("CREATE", s)
}

//OnTimeout is event handler on session tiemout
func (f SessionEventFunc) OnTimeout(s Sessionable) {
	f("TIMEOUT", s)
}

//SessionEventHandler is interface to session event handler
type SessionEventHandler interface {
	OnCreate(s Sessionable)
	OnTimeout(s Sessionable)
}

//SessionBuilder is interface to build the session
type SessionBuilder interface {
	Find(id string) Sessionable
	FindSession(w http.ResponseWriter, r *http.Request) Sessionable
	SetEventHandler(h SessionEventHandler)
}

//Sessionable is interface to record session value
type Sessionable interface {
	xmap.Valuable
	ID() string
	Flush() error
}

//Handler is interface to handler http request
type Handler interface {
	SrvHTTP(*Session) Result
}

//HandlerFunc is func implment Handler
type HandlerFunc func(*Session) Result

//SrvHTTP implement Handler
func (h HandlerFunc) SrvHTTP(s *Session) Result {
	return h(s)
}

//NormalHandlerFunc is normal http handler
type NormalHandlerFunc func(w http.ResponseWriter, r *http.Request)

//SrvHTTP implement Handler
func (n NormalHandlerFunc) SrvHTTP(s *Session) Result {
	n(s.W, s.R)
	return -1
}

// type International interface {
// 	SetLocal(session *Session, local string)
// 	LocalVal(session *Session, key string) string
// }

//Session is http session implement
type Session struct {
	Sessionable
	W   http.ResponseWriter
	R   *http.Request
	Mux *SessionMux
	// INT International
	// V interface{} //response value.
}

//SetCookie will set cookie by key/value
func (s *Session) SetCookie(key string, val string) {
	cookie := &http.Cookie{}
	cookie.Name = key
	cookie.Domain = s.Mux.Domain
	cookie.Path = s.Mux.Path
	cookie.Value = val
	cookie.MaxAge = 0
	// if len(val) < 1 {
	// 	cookie.Expires = time.Unix(0, 0*1e6)
	// }
	http.SetCookie(s.W, cookie)
}

//Cookie will return cookie value by key
func (s *Session) Cookie(key string) (val string) {
	c, err := s.R.Cookie(key)
	if err == nil && c != nil {
		val = c.Value
	}
	return
}

//Redirect will send redirect to url
func (s *Session) Redirect(url string) Result {
	http.Redirect(s.W, s.R, url, http.StatusTemporaryRedirect)
	return Return
}

//Argument will get argument by key from form or post form
func (s *Session) Argument(key string) (sval string) {
	if len(s.R.Form) < 1 && len(s.R.PostForm) < 1 {
		s.R.ParseForm()
	}
	sval = s.R.Form.Get(key)
	if len(sval) < 1 {
		sval = s.R.PostFormValue(key)
	}
	return
}

//Get is implement for attrvalid
func (s *Session) Get(key string) (val interface{}, err error) {
	if len(s.R.Form) < 1 && len(s.R.PostForm) < 1 {
		err = s.R.ParseForm()
	}
	sval := s.R.Form.Get(key)
	if len(sval) < 1 {
		sval = s.R.PostFormValue(key)
	}
	val = sval
	return
}

//ValidFormat is implement for attrvalid
func (s *Session) ValidFormat(format string, args ...interface{}) (err error) {
	err = attrvalid.ValidAttrFormat(format, s, true, args...)
	return
}

var Valider = attrvalid.Valider{
	Scanner: attrscan.Scanner{
		Tag: "json",
		NameConv: func(on, name string, field reflect.StructField) string {
			return name
		},
	},
}

//Valid is implement for valid object and argument
func (s *Session) Valid(target interface{}, filter string, args ...interface{}) (err error) {
	format, args := Valider.ValidArgs(target, filter, args...)
	err = s.ValidFormat(format, args...)
	return
}

//LocalValue will return local value.
func (s *Session) LocalValue(key string) string {
	return key
}

// func http_res(code int, data interface{}, msg string, dmsg string) util.Map {
// 	res := make(util.Map)
// 	res["code"] = code
// 	if len(msg) > 0 {
// 		res["msg"] = msg
// 	}
// 	if data != nil {
// 		res["data"] = data
// 	}
// 	if len(dmsg) > 0 {
// 		res["dmsg"] = dmsg
// 	}
// 	return res
// }
// func http_res_ext(code int, data interface{}, msg string, dmsg string, ext interface{}, pa interface{}) util.Map {
// 	res := make(util.Map)
// 	res["code"] = code
// 	if len(msg) > 0 {
// 		res["msg"] = msg
// 	}
// 	if data != nil {
// 		res["data"] = data
// 	}
// 	if len(dmsg) > 0 {
// 		res["dmsg"] = dmsg
// 	}
// 	if ext != nil {
// 		res["ext"] = ext
// 	}
// 	if pa != nil {
// 		res["pa"] = pa
// 	}
// 	return res
// }

// func json_res(code int, data interface{}, msg string, dmsg string) []byte {
// 	res := http_res(code, data, msg, dmsg)
// 	dbys, _ := json.Marshal(res)
// 	return dbys
// }

// /* International */
// func (s *Session) SetLocal(local string) {
// 	if h.INT != nil {
// 		h.INT.SetLocal(h, local)
// 	}
// }
// func (s *Session) LocalVal(key string) string {
// 	if h.INT != nil {
// 		return h.INT.LocalVal(h, key)
// 	} else {
// 		return ""
// 	}
// }

//Host will return request host
func (s *Session) Host() string {
	return s.R.Host
}

// /* --------------- Access-Language --------------- */
// type LangQ struct {
// 	Lang string
// 	Q    float64
// }
// type LangQes []LangQ

// func (l LangQes) Len() int           { return len(l) }
// func (l LangQes) Swap(i, j int)      { l[i], l[j] = l[j], l[i] }
// func (l LangQes) Less(i, j int) bool { return l[i].Q > l[j].Q }

// func (s *Session) AcceptLanguages() LangQes {
// 	if len(h.R.Header["Accept-Language"]) < 1 {
// 		return LangQes{}
// 	}
// 	lstr := h.R.Header["Accept-Language"][0]
// 	var als LangQes = LangQes{} //all access languages.
// 	regexp.MustCompile("[^;]*;q?[^,]*").ReplaceAllStringFunc(lstr, func(src string) string {
// 		src = strings.Trim(src, "\t \n,")
// 		lq := strings.Split(src, ";")
// 		qua, err := strconv.ParseFloat(strings.Replace(lq[1], "q=", "", -1), 64)
// 		if err != nil {
// 			DebugLog("invalid Accept-Language q:%s", src)
// 			return src
// 		}
// 		for _, lan := range strings.Split(lq[0], ",") {
// 			als = append(als, LangQ{
// 				Lang: lan,
// 				Q:    qua,
// 			})
// 		}
// 		return src
// 	})
// 	sort.Sort(als)
// 	return als
// }

// /* --------------- Access-Language --------------- */

//SessionMux session mux implement
type SessionMux struct {
	xmap.Valuable
	Pre     string
	Domain  string
	Path    string
	Builder SessionBuilder
	//
	FilterEnable   bool
	HandleEnable   bool
	Filters        map[*regexp.Regexp]Handler
	Handlers       map[*regexp.Regexp]Handler
	regexFilterQ   []*regexp.Regexp
	regexFilterM   map[*regexp.Regexp]int
	regexHandlerQ  []*regexp.Regexp
	regexHandlerM  map[*regexp.Regexp]int
	regexMethodM   map[*regexp.Regexp]string
	sessions       map[*http.Request]*Session //request to session
	locker         sync.RWMutex
	CompressLevel  int
	compressRouter map[*regexp.Regexp]int
	//
	// INT           International
	//
	ShowLog  bool
	ShowSlow time.Duration
	M        *monitor.Monitor
}

//NewSessionMux will return new SessionMux
func NewSessionMux(pre string) *SessionMux {
	return NewBuilderSessionMux(pre, NewDefaultSessionBuilder())
}

//NewBuilderSessionMux will create new SessionMux by session builder
func NewBuilderSessionMux(pre string, sb SessionBuilder) *SessionMux {
	mux := SessionMux{}
	mux.Pre = pre
	mux.Domain = ""
	mux.Path = "/"
	mux.Builder = sb
	mux.Filters = map[*regexp.Regexp]Handler{}
	mux.Handlers = map[*regexp.Regexp]Handler{}
	mux.regexFilterM = map[*regexp.Regexp]int{}
	mux.regexFilterQ = []*regexp.Regexp{}
	mux.regexHandlerM = map[*regexp.Regexp]int{}
	mux.regexHandlerQ = []*regexp.Regexp{}
	mux.regexMethodM = map[*regexp.Regexp]string{}
	mux.sessions = map[*http.Request]*Session{}
	mux.Valuable = xmap.New()
	mux.FilterEnable = true
	mux.HandleEnable = true
	mux.ShowLog = false
	// mux.INT = nil
	mux.M = nil
	mux.CompressLevel = gzip.BestSpeed
	mux.compressRouter = map[*regexp.Regexp]int{}
	return &mux
}

// //SetCompress will enable comppress response by pattern
// func (s *SessionMux) SetCompress(pattern string) {
// 	reg := regexp.MustCompile(pattern)
// 	s.compressRouter[reg] = 1
// }

//RequestSession will return sesion by request
func (s *SessionMux) RequestSession(r *http.Request) *Session {
	s.locker.RLock()
	defer s.locker.RUnlock()
	return s.sessions[r]
}

//Filter will register filter
func (s *SessionMux) Filter(pattern string, h Handler) {
	s.FilterMethod(pattern, h, "*")
}

//FilterMethod will register filter
func (s *SessionMux) FilterMethod(pattern string, h Handler, m string) {
	reg := regexp.MustCompile(pattern)
	s.Filters[reg] = h
	s.regexFilterM[reg] = 1
	s.regexFilterQ = append(s.regexFilterQ, reg)
	s.regexMethodM[reg] = m
}

//FilterFunc will register filter by func
func (s *SessionMux) FilterFunc(pattern string, h HandlerFunc) {
	s.Filter(pattern, h)
}

//FilterMethodFunc will register filter by func
func (s *SessionMux) FilterMethodFunc(pattern string, h HandlerFunc, m string) {
	s.FilterMethod(pattern, h, m)
}

//Handle will register handler
func (s *SessionMux) Handle(pattern string, h Handler) {
	s.HandleMethod(pattern, h, "*")
}

//HandleMethod will register handler
func (s *SessionMux) HandleMethod(pattern string, h Handler, m string) {
	reg := regexp.MustCompile(pattern)
	s.Handlers[reg] = h
	s.regexHandlerM[reg] = 1
	s.regexHandlerQ = append(s.regexHandlerQ, reg)
	s.regexMethodM[reg] = m
}

//HandleFunc will register func as handler
func (s *SessionMux) HandleFunc(pattern string, h HandlerFunc) {
	s.Handle(pattern, h)
}

//HandleMethodFunc will register func as handler
func (s *SessionMux) HandleMethodFunc(pattern string, h HandlerFunc, method string) {
	s.HandleMethod(pattern, h, method)
}

//HandleNormal will register normal handler as handler
func (s *SessionMux) HandleNormal(pattern string, h http.Handler) {
	s.HandleNormalMethod(pattern, h, "*")
}

//HandleNormalMethod will register normal handler as handler
func (s *SessionMux) HandleNormalMethod(pattern string, h http.Handler, method string) {
	reg := regexp.MustCompile(pattern)
	s.Handlers[reg] = NormalHandlerFunc(h.ServeHTTP)
	s.regexHandlerM[reg] = 3
	s.regexHandlerQ = append(s.regexHandlerQ, reg)
	// if ret {
	method = fmt.Sprintf("%s,:"+Return.String(), method)
	// } else {
	// 	m = fmt.Sprintf("%s,:CONTINUE", m)
	// }
	s.regexMethodM[reg] = method
}

//HandleNormalFunc will register normal func as handler
func (s *SessionMux) HandleNormalFunc(pattern string, h http.HandlerFunc) {
	s.HandleNormal(pattern, h)
}

//HandleMethodNormalFunc will register normal func as handler, m is http method, * is for all method
func (s *SessionMux) HandleMethodNormalFunc(pattern string, h http.HandlerFunc, method string) {
	s.HandleNormalMethod(pattern, h, method)
}

func (s *SessionMux) slog(fmt string, args ...interface{}) {
	if s.ShowLog {
		DebugLog(fmt, args...)
	}
}
func (s *SessionMux) checkMethod(reg *regexp.Regexp, m string) bool {
	tm, ok := s.regexMethodM[reg]
	return ok && strings.Contains(tm, "*") || strings.Contains(tm, m)
}

func (s *SessionMux) checkContinue(reg *regexp.Regexp) bool {
	tm, ok := s.regexMethodM[reg]
	return ok && strings.Contains(tm, ":"+Continue.String())
}

func (s *SessionMux) execFilter(hs *Session) (bool, Result) {
	url := hs.R.URL.Path
	var matched bool = false
	for _, k := range s.regexFilterQ {
		if !k.MatchString(url) {
			continue
		}
		if !s.checkMethod(k, hs.R.Method) {
			s.slog("not mathced method %v to %v", hs.R.Method, s.regexMethodM[k])
			continue
		}
		var mid = ""
		matched = true
		if s.M != nil {
			mid = s.M.Start(fmt.Sprintf("F_%v", k.String()))
		}
		rv := s.Filters[k]
		res := rv.SrvHTTP(hs)
		if s.M != nil {
			s.M.Done(mid)
		}
		s.slog("mathced filter %v to %v (%v)", k, hs.R.URL.Path, res.String())
		if res == Return {
			return matched, res
		}
	}
	return matched, Continue
}

func (s *SessionMux) execHandler(hs *Session) (bool, Result) {
	url := hs.R.URL.Path
	var matched bool = false
	for _, k := range s.regexHandlerQ {
		if !k.MatchString(url) {
			continue
		}
		if !s.checkMethod(k, hs.R.Method) {
			s.slog("not mathced method %v to %v", hs.R.Method, s.regexMethodM[k])
			continue
		}
		var mid = ""
		matched = true
		switch s.regexHandlerM[k] {
		case 1:
			fallthrough
		case 2:
			if s.M != nil {
				mid = s.M.Start(fmt.Sprintf("H_%v", k.String()))
			}
			rv := s.Handlers[k]
			res := rv.SrvHTTP(hs)
			if s.M != nil {
				s.M.Done(mid)
			}
			s.slog("mathced handler %v to %v (%v)", k, hs.R.URL.Path, res.String())
			if res == Return {
				return matched, res
			}
		case 3:
			fallthrough
		case 4:
			if s.M != nil {
				mid = s.M.Start(fmt.Sprintf("H_%v", k.String()))
			}
			rv := s.Handlers[k]
			rv.SrvHTTP(hs)
			if s.M != nil {
				s.M.Done(mid)
			}
			if s.checkContinue(k) {
				s.slog("mathced normal handler %v to %v (%v)", k, hs.R.URL.Path, Continue.String())
				continue
			} else {
				s.slog("mathced normal handler %v to %v (%v)", k, hs.R.URL.Path, Return.String())
				return matched, Return
			}
		}
	}
	return matched, Continue
}

// func (s *SessionMux) isCompress(w http.ResponseWriter, r *http.Request) bool {
// 	if s.CompressContent {
// 		for reg := range s.compressRouter {
// 			if reg.MatchString(r.URL.Path) {
// 				return true
// 			}
// 		}
// 	}
// 	return false
// }

//
func (s *SessionMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.slog("receive %v request by %v", r.Method, r.URL)
	beg := time.Now()
	r.URL.Path = strings.TrimPrefix(r.URL.Path, s.Pre)
	session := s.Builder.FindSession(w, r)
	hs := &Session{
		W:           w,
		R:           r,
		Sessionable: session,
		Mux:         s,
	}
	s.locker.Lock()
	s.sessions[r] = hs
	s.locker.Unlock()
	defer func() {
		s.locker.Lock()
		delete(s.sessions, r) //remove the http session object.
		s.locker.Unlock()
		used := time.Since(beg)
		if s.ShowSlow > 0 && used > s.ShowSlow {
			WarnLog("SessionMux slow request found->%v", r.URL.String())
		}
	}()
	//
	// if s.isCompress(w, r) {
	// 	writer, err := NewGzipResponseWriter(hs.W, s.CompressLevel)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	w.Header().Set("Content-Encoding", "gzip")
	// 	writer.Header().Set("Content-Encoding", "gzip")
	// 	fmt.Println(w.Header())
	// 	hs.W = writer
	// }
	//
	var matched bool = false
	//
	defer func() {
		if !matched { //if not matched
			s.slog("not matchd any filter:%s", r.URL.Path)
			http.NotFound(w, r)
		}
		// var tv interface{} = hs.V
		// if s.FIND_V != nil {
		// 	if fv := s.FIND_V(hs); fv != nil {
		// 		tv = fv(hs.V)
		// 	}
		// }
		// hooks.Call(HK_ROUTING, HK_R_END, tv, hs, matched)
		// if gz, ok := hs.W.(*GzipResponseWriter); ok {
		// 	gz.Writer.Close()
		// }
	}()
	// hooks.Call(HK_ROUTING, HK_R_BEG, nil, hs)
	//match filter.
	if s.FilterEnable {
		// hooks.Call(HK_ROUTING, HK_F_BEG, nil, hs)
		mrv, res := s.execFilter(hs)
		matched = mrv
		// hooks.Call(HK_ROUTING, HK_F_END, nil, hs, mrv, res)
		if res == Return {
			return
		}
	}
	//match handle
	if s.HandleEnable {
		// hooks.Call(HK_ROUTING, HK_H_BEG, nil, hs)
		mrv, _ := s.execHandler(hs)
		matched = matched || mrv
		// hooks.Call(HK_ROUTING, HK_H_END, nil, hs, mrv, res)
	}
}

//Print will show all current handler info
func (s *SessionMux) Print() {
	if len(s.Filters) > 0 {
		fmt.Println(" >Filters---->")
		for reg, h := range s.Filters {
			fmt.Printf("\t%v->%p\n", reg.String(), h)
		}
	}
	if len(s.Handlers) > 0 {
		fmt.Println(" >Handlers---->")
		for reg, h := range s.Handlers {
			fmt.Printf("\t%v->%p\n", reg.String(), h)
		}
	}
}

//StartMonitor will start monitor
func (s *SessionMux) StartMonitor() {
	s.M = monitor.New()
}

//State for implement Statable for get current state
func (s *SessionMux) State() (state interface{}, err error) {
	if s.M != nil {
		state, err = s.M.State()
	}
	return
}

// //GzipResponseWriter is response compress writer
// type GzipResponseWriter struct {
// 	http.ResponseWriter
// 	Writer       *gzip.Writer
// 	headerSetted uint32
// }

// //NewGzipResponseWriter will return new response writer
// func NewGzipResponseWriter(w http.ResponseWriter, level int) (writer *GzipResponseWriter, err error) {
// 	writer = &GzipResponseWriter{}
// 	writer.ResponseWriter = w
// 	writer.Writer, err = gzip.NewWriterLevel(w, level)
// 	return
// }

// func (g *GzipResponseWriter) Write(p []byte) (n int, err error) {
// 	n, err = g.Writer.Write(p)
// 	return
// }
