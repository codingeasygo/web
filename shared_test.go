package web

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/codingeasygo/util/xhttp"
)

func TestShared(t *testing.T) {
	waiter := sync.WaitGroup{}
	waiter.Add(2)
	go func() {
		ListenAndServe(":12332")
		waiter.Done()
	}()
	go func() {
		HandleSignal()
		waiter.Done()
	}()
	FilterFunc("/f1/h1", func(s *Session) Result {
		s.SetValue("a", "123")
		return Continue
	})
	HandleFunc("/f1/h1", func(s *Session) Result {
		return s.Printf("%v", s.Value("a"))
	})
	text, err := xhttp.GetText("%v/f1/h1", "http://127.0.0.1:12332")
	if err != nil || text != "123" {
		t.Errorf("err:%v,text:%v", err, text)
		return
	}
	//
	close(sigc)
	waiter.Wait()
	//
	err = ListenAndServe(":xxxx")
	if err == nil {
		t.Error(err)
		return
	}
}

func TestListenUnix(t *testing.T) {
	var err error
	os.Remove("/tmp/xtest.unix")
	waiter := sync.WaitGroup{}
	waiter.Add(2)
	go func() {
		fmt.Println(ListenAndServe("/tmp/xtest.unix,0777"))
		waiter.Done()
	}()
	go func() {
		HandleSignal()
		waiter.Done()
	}()
	time.Sleep(100 * time.Millisecond)
	//
	close(sigc)
	waiter.Wait()
	//
	//test error
	err = ListenAndServe("/xx/xx/xtest.unix,0777")
	if err == nil {
		t.Error(err)
		return
	}
	err = ListenAndServe("/tmp/xtest.unix,x0777")
	if err == nil {
		t.Error(err)
		return
	}
}
