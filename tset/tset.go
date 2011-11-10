package tset

import (
	"os"
	"fmt"
	"strconv"
	"http"
	"template"

	"appengine"

	"httputils"
	"core"
)

type Context map[string]interface{}

func urlFor(name string, pairs ...interface{}) string {
	size := len(pairs)
	strPairs := make([]string, size, size)
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

func TemplatePath(name string) string {
	return "templates/" + name
}

var templateSetCache = make(map[string]*template.Set)

func TemplateSet(name string) (*template.Set, os.Error) {
	if s, ok := templateSetCache[name]; ok {
		return s, nil
	}

	s := &template.Set{}
	s.Funcs(map[string]interface{}{"urlFor": urlFor, "Itoa64": strconv.Itoa64})

	s, err := s.ParseFiles(TemplatePath(name))
	if err != nil {
		return nil, err
	}

	_, err = s.ParseTemplateFiles(TemplatePath("layout.html"))
	if err != nil {
		return nil, err
	}

	templateSetCache[name] = s
	return s, nil
}

func RenderTemplate(c appengine.Context, w http.ResponseWriter, name string, vars interface{}) {
	s, err := TemplateSet(name)
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}

	err = s.Execute(w, "layout.html", vars)
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}
}
