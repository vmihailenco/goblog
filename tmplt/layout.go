package tmplt

import (
	"os"
	"bytes"
	"template"
	"sync"
)

type Context map[string]interface{}

type Layout struct {
	basePath         string
	LayoutPath       string
	filenames        []string
	funcMap          template.FuncMap
	templateSetCache map[string]*template.Set
	mutex            sync.RWMutex
}

func NewLayout(basePath string, LayoutPath string) *Layout {
	return &Layout{
		basePath:         basePath,
		LayoutPath:       LayoutPath,
		funcMap:          make(template.FuncMap),
		templateSetCache: make(map[string]*template.Set),
	}
}

func (l *Layout) NewLayout() *Layout {
	return &Layout{
		basePath:         l.basePath,
		LayoutPath:       l.LayoutPath,
		filenames:        l.filenames,
		funcMap:          l.funcMap,
		templateSetCache: make(map[string]*template.Set),
	}
}

func (l *Layout) SetFilenames(filenames ...string) *Layout {
	for i, f := range filenames {
		filenames[i] = l.templatePath(f)
	}
	l.filenames = filenames
	return l
}

func (l *Layout) SetFuncMap(funcMap template.FuncMap) *Layout {
	l.funcMap = funcMap
	return l
}

func (l *Layout) NewTemplateSet() (*template.Set, os.Error) {
	s := &template.Set{}
	s.Funcs(l.funcMap)

	if _, err := s.ParseTemplateFiles(l.templatePath(l.LayoutPath)); err != nil {
		return nil, err
	}

	if _, err := s.ParseFiles(l.filenames...); err != nil {
		return nil, err
	}

	return s, nil
}

func (l *Layout) templateSetFromCache(filename string) *template.Set {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	if s, ok := l.templateSetCache[filename]; ok {
		return s
	}
	return nil
}

func (l *Layout) TemplateSet(filename string) (*template.Set, os.Error) {
	s := l.templateSetFromCache(filename)
	if s != nil {
		return s, nil
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	// cache may be already filled by another goroutine
	if s, ok := l.templateSetCache[filename]; ok {
		return s, nil
	}

	s, err := l.NewTemplateSet()
	if err != nil {
		return nil, err
	}

	if filename != "" {
		if _, err := s.ParseFiles(l.templatePath(filename)); err != nil {
			return nil, err
		}
	}

	l.templateSetCache[filename] = s
	return s, nil
}

func (l *Layout) Render(context Context, filename string) (*bytes.Buffer, os.Error) {
	s, err := l.TemplateSet(filename)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err := s.Execute(buf, l.LayoutPath, context); err != nil {
		return nil, err
	}
	return buf, nil
}

func (l *Layout) templatePath(path string) string {
	if path != "" && path[0:1] == "/" {
		return path
	}
	return l.basePath + "/" + path
}
