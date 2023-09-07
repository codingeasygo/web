package web

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/codingeasygo/util/xmap"
)

// DefaultSession is web memory sesson implement
type DefaultSession struct {
	*xmap.SafeM
	id string
}

// ID will return the session id
func (d *DefaultSession) ID() string {
	return d.id
}

// Flush will flush value
func (d *DefaultSession) Flush() error {
	return nil
}

// DefaultSessionBuilder is session builder to create default session
type DefaultSessionBuilder struct {
	sessions map[string]*DefaultSession
	locker   sync.RWMutex
}

// NewDefaultSessionBuilder will return new builder
func NewDefaultSessionBuilder() *DefaultSessionBuilder {
	return &DefaultSessionBuilder{
		sessions: map[string]*DefaultSession{},
		locker:   sync.RWMutex{},
	}
}

// Find will return session by id
func (s *DefaultSessionBuilder) Find(id string) Sessionable {
	s.locker.RLock()
	defer s.locker.RUnlock()
	return s.sessions[id]
}

// FindSession will find session by http
func (s *DefaultSessionBuilder) FindSession(w http.ResponseWriter, r *http.Request) Sessionable {
	sid := fmt.Sprintf("%p", r)
	s.locker.Lock()
	session, ok := s.sessions[sid]
	if !ok {
		session = &DefaultSession{id: sid, SafeM: xmap.NewSafe()}
		s.sessions[sid] = session
	}
	s.locker.Unlock()
	return session
}

// SetEventHandler will set the session event handler
func (s *DefaultSessionBuilder) SetEventHandler(h SessionEventHandler) {
}
