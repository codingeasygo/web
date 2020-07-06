package web

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

//SendFile will send file to session
func (s *Session) SendFile(name, filename, contentType string) Result {
	SendFile(s.W, s.R, name, filename, contentType, true)
	return Return
}

//SendBinary will send file to session
func (s *Session) SendBinary(filename, contentType string) Result {
	SendFile(s.W, s.R, "", filename, contentType, false)
	return Return
}

//SendBytes string by target context type.
func (s *Session) SendBytes(data []byte, contentType string) Result {
	header := s.W.Header()
	header.Set("Content-Type", contentType)
	header.Set("Content-Length", fmt.Sprintf("%v", len(data)))
	// header.Set("Content-Transfer-Encoding", "binary")
	header.Set("Expires", "0")
	s.W.Write(data)
	return Return
}

//SendString string by target context type.
func (s *Session) SendString(data string, contentType string) Result {
	s.SendBytes([]byte(data), contentType)
	return Return
}

//SendPlainText will send string by text/plain content type
func (s *Session) SendPlainText(data string) Result {
	return s.SendBytes([]byte(data), ContentTypePlainText)
}

//SendJSON will parse value to json and send it
func (s *Session) SendJSON(v interface{}) Result {
	data, err := json.Marshal(v)
	if err != nil {
		ErrorLog("sending json(%v) fail with %s", v, err.Error())
		http.Error(s.W, err.Error(), 500)
	} else {
		s.SendBytes(data, ContentTypeJSON)
	}
	return Return
}

//SendFile will send file to http response
func SendFile(w http.ResponseWriter, r *http.Request, name, filename, contentType string, attach bool) (err error) {
	defer func() {
		if err != nil {
			ErrorLog("sending file(%v) error:%s", filename, err.Error())
		}
	}()
	src, err := os.Open(filename)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	err = sendf(w, src, filename, contentType, attach)
	return
}

func sendf(w http.ResponseWriter, file *os.File, name, contentType string, attach bool) error {
	fi, err := file.Stat()
	if err == nil {
		fsize := fi.Size()
		header := w.Header()
		if len(contentType) < 1 {
			header.Set("Content-Type", "application/octet-stream")
		} else {
			header.Set("Content-Type", contentType)
		}
		if attach && len(name) > 0 {
			header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url.QueryEscape(name)))
			header.Set("Content-Transfer-Encoding", "binary")
		}
		header.Set("Content-Length", fmt.Sprintf("%v", fsize))
		header.Set("Expires", "0")
		_, err = io.Copy(w, file)
	}
	return err
}

//Printf will printf format string to http response
func (s *Session) Printf(format string, args ...interface{}) Result {
	header := s.W.Header()
	header.Set("Content-Type", ContentTypePlainText)
	fmt.Fprintf(s.W, format, args...)
	return Return
}
