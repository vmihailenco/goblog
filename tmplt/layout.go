package tmplt

import (
	"os"
	"bytes"
	"template"
)

type Context map[string]interface{}

type layout struct {
	basePath         string
	layoutPath       string
	filenames        []string
	funcMap          template.FuncMap
	templateSetCache map[string]*template.Set
}

func NewLayout(basePath string, layoutPath string) *layout {
	return &layout{
		basePath,
		layoutPath,
		nil,
		make(template.FuncMap),
		make(map[string]*template.Set),
	}
}

func (l *layout) NewLayout() *layout {
	return &layout{
		l.basePath,
		l.layoutPath,
		l.filenames,
		l.funcMap,
		make(map[string]*template.Set),
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

func (l *layout) TemplateSet(filename string) (*template.Set, os.Error) {
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
