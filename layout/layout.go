package layout

import (
	"os"
	"fmt"
	"runtime/debug"
	"bytes"
	"template"

	"core"
)

type Context map[string]interface{}

type Layout struct {
	basePath string
	layoutPath string
	filenames []string
	templateSetCache map[string]*template.Set
}

func NewLayout(basePath string, layoutPath string, filenames ...string) *Layout {
	l := &Layout{basePath, layoutPath, nil, make(map[string]*template.Set)}
	for i, f := range filenames {
		filenames[i] = l.templatePath(f)
	}
	l.filenames = filenames
	return l
}

func (l *Layout) NewTemplateSet() (*template.Set, os.Error) {
	s := &template.Set{}
	s.Funcs(map[string]interface{}{"urlFor": urlFor})

	if 	_, err := s.ParseTemplateFiles(l.templatePath(l.layoutPath)); err != nil {
		return nil, err
	}

	if _, err := s.ParseFiles(l.filenames...); err != nil {
		return nil, err
	}

	return s, nil
}

func (l *Layout) TemplateSet(filename string) (*template.Set, os.Error) {
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

func (l *Layout) templatePath(path string) string {
	return l.basePath + "/" + path
}

func (l *Layout) Render(context Context, filename string) (*bytes.Buffer, os.Error) {
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

func TemplateFuncRecover() {
	if err := recover(); err != nil {
		errStr := fmt.Sprint(err)
		fmt.Println(errStr)
		debug.PrintStack()
		panic(os.NewError(errStr))
	}
}

func urlFor(name string, pairs ...interface{}) string {
	defer TemplateFuncRecover()

	size := len(pairs)
	strPairs := make([]string, size)
	for i := 0; i < size; i++ {
		if v, ok := pairs[i].(string); ok {
			strPairs[i] = v
		} else {
			strPairs[i] = fmt.Sprint(pairs[i])
		}
	}
	route, _ := core.Router.NamedRoutes[name]
	return route.URL(strPairs...).String()
}
