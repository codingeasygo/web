package web

import (
	"net/http"
	"os"
	"regexp"
	"time"
)

//Dir is implement http file system for ignore cache
type Dir struct {
	http.Dir
	Inc     []*regexp.Regexp
	ShowLog bool
}

//File is implement http file system for ignore cache
type File struct {
	http.File
}

//FileInfo is implement http file system for ignore cache
type FileInfo struct {
	os.FileInfo
}

func (d *Dir) log(f string, args ...interface{}) {
	if d.ShowLog {
		DebugLog(f, args...)
	}
}

//Add will add regexp for enable ignore cache
func (d *Dir) Add(m *regexp.Regexp) {
	d.Inc = append(d.Inc, m)
}

//Open will opoen the file by name
func (d *Dir) Open(name string) (http.File, error) {
	rf, err := d.Dir.Open(name)
	if err != nil {
		return rf, err
	}
	for _, inc := range d.Inc {
		if inc.MatchString(name) {
			d.log("not cahce for path:%v", name)
			return &File{File: rf}, nil
		}
	}
	d.log("using normal file system for path:%v", name)
	return rf, nil
}

//Stat will return file info
func (f *File) Stat() (os.FileInfo, error) {
	d, err := f.File.Stat()
	return &FileInfo{FileInfo: d}, err
}

//ModTime will return the file mode file time
func (f *FileInfo) ModTime() time.Time {
	return time.Now()
}

//NewNoCacheDir will return new not cahche dir
func NewNoCacheDir(path string) *Dir {
	return &Dir{
		Dir: http.Dir(path),
		Inc: []*regexp.Regexp{},
	}
}

//NewAllNoCacheDir will return new all not cache dir
func NewAllNoCacheDir(path string) *Dir {
	return &Dir{
		Dir: http.Dir(path),
		Inc: []*regexp.Regexp{regexp.MustCompile("^.*$")},
	}
}
