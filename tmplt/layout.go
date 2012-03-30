package tmplt

import (
	"html/template"
	"sync"
)

type Context map[string]interface{}

type TmpltHolder struct {
	templateCache map[string]*template.Template
	mutex         sync.RWMutex
}

func NewTmpltHolder() *TmpltHolder {
	return &TmpltHolder{templateCache: make(map[string]*template.Template)}
}

var Holder = NewTmpltHolder()

func (h *TmpltHolder) Get(filename string, base *template.Template) (*template.Template, error) {
	if t, ok := h.templateCache[filename]; ok {
		return t, nil
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	if t, ok := h.templateCache[filename]; ok {
		return t, nil
	}

	t, err := base.Clone()
	if err != nil {
		return nil, err
	}

	t, err = t.ParseFiles(filename)
	if err != nil {
		return nil, err
	}

	h.templateCache[filename] = t

	return t, nil
}
