package web

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/codingeasygo/util/converter"
	"github.com/codingeasygo/util/xio"
)

//MultipartFile is recevied file result
type MultipartFile struct {
	Name     string
	Filename string
	SavePath string
	Length   int64
	SHA1     []byte
	MD5      []byte
}

//MultipartValues is recevied result
type MultipartValues struct {
	Files  []*MultipartFile
	Values map[string][][]byte
}

//RecvMultipart will recv http body as multi part
func (s *Session) RecvMultipart(enableSHA1, enableMD5 bool, savePathFunc func(*multipart.Part) (filename string, mode os.FileMode, external []io.Writer, err error)) (*MultipartValues, error) {
	mr, err := s.R.MultipartReader()
	if err != nil {
		return nil, fmt.Errorf("MultipartReader err(%v)", err.Error())
	}
	vals := &MultipartValues{
		Values: map[string][][]byte{},
	}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			return vals, fmt.Errorf("NextPart err(%v)", err.Error())
		}
		if len(part.FileName()) < 1 {
			bys, err := ioutil.ReadAll(part)
			if err != nil {
				part.Close()
				return vals, err
			}
			part.Close()
			vals.Values[part.FormName()] = append(vals.Values[part.FormName()], bys)
			continue
		}
		filename, filemode, external, err := savePathFunc(part)
		if err != nil {
			part.Close()
			return vals, err
		}
		_, fn := filepath.Split(part.FileName())
		if strings.HasSuffix(filename, "/") {
			filename = filename + fn
		}
		dir, _ := filepath.Split(filename)
		os.MkdirAll(dir, filemode)
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, filemode)
		if err != nil {
			part.Close()
			return vals, err
		}
		external = append(external, file)
		var md5h, shah hash.Hash
		if enableMD5 {
			md5h = md5.New()
			external = append(external, md5h)
		}
		if enableSHA1 {
			shah = sha1.New()
			external = append(external, shah)
		}
		w, err := xio.CopyMulti(external, part)
		part.Close()
		file.Close()
		if err != nil {
			return vals, err
		}
		vf := &MultipartFile{
			Name:     part.FormName(),
			Filename: part.FileName(),
			SavePath: filename,
			Length:   w,
		}
		if md5h != nil {
			vf.MD5 = md5h.Sum(nil)
		}
		if shah != nil {
			vf.SHA1 = shah.Sum(nil)
		}
		vals.Files = append(vals.Files, vf)
	}
	return vals, nil
}

//RecvFile will receive form file and save to filename
func (s *Session) RecvFile(enableSHA1, enableMD5 bool, name, filename string) (*MultipartValues, error) {
	vals, err := s.RecvMultipart(enableSHA1, enableMD5, func(part *multipart.Part) (fn string, mode os.FileMode, external []io.Writer, err error) {
		fn = filename
		mode = os.ModePerm
		if len(part.FileName()) > 0 && part.FormName() != name {
			err = fmt.Errorf("file form name is %v, expect %v", part.FormName(), name)
		}
		return
	})
	if err == nil && len(vals.Files) < 1 {
		err = fmt.Errorf("file form name by %v is not exists", name)
	}
	return vals, err
}

//RecvFileBytes will receive body to bytes
func (s *Session) RecvFileBytes(name string, maxMemory int64) (data []byte, err error) {
	// err = s.R.ParseMultipartForm(maxMemory)
	// if err != nil {
	// 	return
	// }
	src, _, err := s.R.FormFile(name)
	if err != nil {
		return nil, err
	}
	defer src.Close()
	buffer := bytes.NewBuffer(nil)
	_, err = xio.CopyMax(buffer, src, maxMemory)
	if err == nil || err == io.EOF {
		err = nil
		data = buffer.Bytes()
	}
	return
}

//RecvBody will receive body and parse to json object
func (s *Session) RecvBody() (data []byte, err error) {
	data, err = ioutil.ReadAll(s.R.Body)
	return
}

//RecvJSON will receive body and parse to json object
func (s *Session) RecvJSON(v interface{}) (data []byte, err error) {
	data, err = converter.UnmarshalJSON(s.R.Body, v)
	return
}

//RecvXML will receive body and parse to xml object
func (s *Session) RecvXML(v interface{}) (data []byte, err error) {
	data, err = converter.UnmarshalXML(s.R.Body, v)
	return
}

type fsSizable interface {
	Size() int64
}

type fsStatable interface {
	Stat() (os.FileInfo, error)
}

// type fsNamable interface {
// 	Name() string
// }

func formFileSzie(src interface{}) int64 {
	var fsize int64 = 0
	if statInterface, ok := src.(fsStatable); ok {
		fileInfo, _ := statInterface.Stat()
		fsize = fileInfo.Size()
	}
	if sizeInterface, ok := src.(fsSizable); ok {
		fsize = sizeInterface.Size()
	}
	return fsize
}

//FormFileInfo will return form file info
func (s *Session) FormFileInfo(name string) (filesize int64, filename string, err error) {
	src, fh, err := s.R.FormFile(name)
	if err == nil {
		filesize = formFileSzie(src)
		filename = fh.Filename
	}
	return
}
