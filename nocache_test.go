package web

import (
	"fmt"
	"regexp"
	"testing"
)

func TestNoCache(t *testing.T) {
	NewAllNoCacheDir("www")
	nnd := NewNoCacheDir("test")
	nnd.ShowLog = true
	nnd.Add(regexp.MustCompile(`^test\.html(\?.*)?$`))
	f, _ := nnd.Open("test.html")
	fi, _ := f.Stat()
	fmt.Println(fi.ModTime())
	nnd.Open("test2.html")
	nnd.Open("tt.html")
}
