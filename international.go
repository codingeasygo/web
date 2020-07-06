package web

// import (
// 	"encoding/json"
// 	"io/ioutil"
// 	"path/filepath"
// 	"sync"
// )

// type JSONINT struct {
// 	Path    string
// 	Default string
// 	Local   string
// 	Kvs     map[string]map[string]string
// 	lock    sync.RWMutex
// }

// func (j *JSONINT) SetLocal(hs *HTTPSession, local string) {
// 	j.Local = local
// }

// func (j *JSONINT) LocalVal(hs *HTTPSession, key string) string {
// 	if len(j.Local) > 0 {
// 		val := j.LangVal(j.Local, key)
// 		if len(val) > 0 {
// 			return val
// 		}
// 	}
// 	als := hs.AcceptLanguages()
// 	if len(als) < 1 {
// 		return j.LangVal(j.Default, key)
// 	} else {
// 		return j.LangQesVal(als, key)
// 	}
// }

// func (j *JSONINT) LangQesVal(als LangQes, key string) string {
// 	for _, al := range als {
// 		val := j.LangVal(al.Lang, key)
// 		if len(val) > 0 {
// 			return val
// 		}
// 	}
// 	return j.LangVal(j.Default, key)
// }

// func (j *JSONINT) LangVal(lang string, key string) string {
// 	if _, ok := j.Kvs[lang]; !ok {
// 		j.lock.Lock()
// 		err := j.LoadJson(lang)
// 		if err != nil { //load error or not found,marked loaded.
// 			j.Kvs[lang] = map[string]string{}
// 		}
// 		j.lock.Unlock()
// 	}
// 	return j.Kvs[lang][key]
// }

// func (j *JSONINT) LoadJson(lang string) error {
// 	fpath := filepath.Join(j.Path, lang+".json")
// 	bys, err := ioutil.ReadFile(fpath)
// 	if err != nil {
// 		WarnLog("reading file(%v) error(%v)", fpath, err.Error())
// 		return err
// 	}
// 	mv := map[string]string{}
// 	err = json.Unmarshal(bys, &mv)
// 	if err != nil {
// 		WarnLog("load the lang(%s) json file(%s) error:%s", lang, fpath, err.Error())
// 		return err
// 	}
// 	j.Kvs[lang] = mv
// 	return nil
// }

// func NewJSONINT(path string) (*JSONINT, error) {
// 	ji := &JSONINT{}
// 	ji.Path = path
// 	ji.Kvs = map[string]map[string]string{}
// 	ji.lock = sync.RWMutex{}
// 	return ji, nil
// }
