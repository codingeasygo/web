package web

import (
	"net/http"
	"sync"
	"time"

	"github.com/codingeasygo/util/uuid"
	"github.com/codingeasygo/util/xmap"
)

//MemSession is memory session implement
type MemSession struct {
	xmap.Valuable
	token  string
	latest time.Time
}

//ID return the session id
func (m *MemSession) ID() string {
	return m.token
}

//Flush will flush session latest time
func (m *MemSession) Flush() error {
	m.latest = time.Now()
	return nil
}

//MemSessionBuilder is memory session builder implement
type MemSessionBuilder struct {
	xmap.Valuable
	Domain    string
	Path      string
	Timeout   time.Duration
	CookieKey string //cookie key
	ShowLog   bool
	Event     SessionEventHandler
	//
	delay    time.Duration
	looping  bool
	sessions map[string]*MemSession //key session
	locker   sync.RWMutex
}

//NewMemSessionBuilder will return new MemSessionBuilder
func NewMemSessionBuilder(domain string, path string, cookie string, timeout time.Duration) *MemSessionBuilder {
	sb := MemSessionBuilder{}
	sb.Domain = domain
	sb.Path = path
	sb.Timeout = timeout
	sb.delay = time.Second
	sb.CookieKey = cookie
	sb.Valuable = xmap.New()
	sb.ShowLog = false
	sb.Valuable = xmap.New()
	sb.sessions = map[string]*MemSession{}
	sb.locker = sync.RWMutex{}
	return &sb
}
func (m *MemSessionBuilder) log(f string, args ...interface{}) {
	if m.ShowLog {
		DebugLog(f, args...)
	}
}

//Find will find sesion by tokken
func (m *MemSessionBuilder) Find(id string) (session Sessionable) {
	m.locker.RLock()
	defer m.locker.RUnlock()
	if v, ok := m.sessions[id]; ok {
		session = v
	}
	return
}

//FindSession will find the session by request
func (m *MemSessionBuilder) FindSession(w http.ResponseWriter, r *http.Request) Sessionable {
	c, err := r.Cookie(m.CookieKey)
	ncookie := func() {
		c = &http.Cookie{}
		c.Name = m.CookieKey
		c.Value = uuid.New()
		c.Path = m.Path
		c.Domain = m.Domain
		c.MaxAge = 10 * 24 * 60 * 60
		//
		session := &MemSession{}
		session.token = c.Value
		session.Valuable = xmap.NewSafe()
		session.Flush()
		//
		// s.ks_lck.Lock()
		m.sessions[c.Value] = session
		// s.ks_lck.Unlock()
		http.SetCookie(w, c)
		if m.Event != nil {
			m.Event.OnCreate(session)
		}
		// s.log("setting cookie %v=%v to %v", c.Name, c.Value, r.Host)
	}
	m.locker.Lock()
	defer m.locker.Unlock()
	var ss Sessionable
	if w != nil {
		if err != nil {
			ncookie()
		}
		if _, ok := m.sessions[c.Value]; !ok { //if not found,reset cookie
			ncookie()
		}
		ss = m.sessions[c.Value]
		ss.Flush()
	} else {
		if err == nil {
			ss = m.sessions[c.Value]
		}
	}
	return ss

}

//SetEventHandler will set event handler
func (m *MemSessionBuilder) SetEventHandler(h SessionEventHandler) {
	m.Event = h
}

//StartTimeout will start timeout
func (m *MemSessionBuilder) StartTimeout() {
	if m.Timeout > 0 {
		m.looping = true
		go m.loopTimeout()
	}
}

//StopTimeout will stop timeout
func (m *MemSessionBuilder) StopTimeout() {
	m.looping = false
}

//
func (m *MemSessionBuilder) loopTimeout() {
	for m.looping {
		ary := []string{}
		now := time.Now()
		m.locker.RLock()
		for k, v := range m.sessions {
			delay := now.Sub(v.latest)
			if delay > m.Timeout {
				ary = append(ary, k)
			}
		}
		m.locker.RUnlock()
		if len(ary) > 0 {
			m.log("looping session time out,removing (%v)", ary)
			m.locker.Lock()
			for _, v := range ary {
				s := m.sessions[v]
				delete(m.sessions, v)
				if m.Event != nil {
					m.Event.OnTimeout(s)
				}
			}
			m.locker.Unlock()
		}
		time.Sleep(m.delay)
	}
}
