package tmplt

import (
	"html/template"
	"sync"
)

type Context map[string]interface{}

type TmpltHolder struct {
	templateCache map[string]*template.Template
	mutex         sync.Mutex
}

func NewTmpltHolder() *TmpltHolder {
	return &TmpltHolder{templateCache: make(map[string]*template.Template)}
}

var Holder = NewTmpltHolder()

type NewFunc func() (*template.Template, error)

func (h *TmpltHolder) Get(key string, newFunc NewFunc) (*template.Template, error) {
	if t, ok := h.templateCache[key]; ok {
		return t, nil
	}

	h.mutex.Lock()
	defer h.mutex.Unlock()

	if t, ok := h.templateCache[key]; ok {
		return t, nil
	}

	t, err := newFunc()
	if err != nil {
		return nil, err
	}

	h.templateCache[key] = t

	return t, nil
}
