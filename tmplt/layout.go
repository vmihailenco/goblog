package tmplt

import (
	"os"
	"bytes"
	"template"
	"sync"
)

type Context map[string]interface{}

type layout struct {
	basePath         string
	layoutPath       string
	filenames        []string
	funcMap          template.FuncMap
	templateSetCache map[string]*template.Set
	mutex            sync.RWMutex
}

func NewLayout(basePath string, layoutPath string) *layout {
	return &layout{
		basePath:         basePath,
		layoutPath:       layoutPath,
		funcMap:          make(template.FuncMap),
		templateSetCache: make(map[string]*template.Set),
	}
}

func (l *layout) NewLayout() *layout {
	return &layout{
		basePath:         l.basePath,
		layoutPath:       l.layoutPath,
		filenames:        l.filenames,
		funcMap:          l.funcMap,
		templateSetCache: make(map[string]*template.Set),
	}
}

func (l *layout) SetFilenames(filenames ...string) *layout {
	for i, f := range filenames {
		filenames[i] = l.templatePath(f)
	}
	l.filenames = filenames
	return l
}

func (l *layout) SetFuncMap(funcMap template.FuncMap) *layout {
	l.funcMap = funcMap
	return l
}

func (l *layout) NewTemplateSet() (*template.Set, os.Error) {
	s := &template.Set{}
	s.Funcs(l.funcMap)

	if _, err := s.ParseTemplateFiles(l.templatePath(l.layoutPath)); err != nil {
		return nil, err
	}

	if _, err := s.ParseFiles(l.filenames...); err != nil {
		return nil, err
	}

	return s, nil
}

func (l *layout) templateSetFromCache(filename string) *template.Set {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	if s, ok := l.templateSetCache[filename]; ok {
		return s
	}
	return nil
}

func (l *layout) TemplateSet(filename string) (*template.Set, os.Error) {
	s := l.templateSetFromCache(filename)
	if s != nil {
		return s, nil
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

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

func (l *layout) Render(context Context, filename string) (*bytes.Buffer, os.Error) {
	s, err := l.TemplateSet(filename)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if err := s.Execute(buf, l.layoutPath, context); err != nil {
		return nil, err
	}
	return buf, nil
}

func (l *layout) templatePath(path string) string {
	if path != "" && path[0:1] == "/" {
		return path
	}
	return l.basePath + "/" + path
}
