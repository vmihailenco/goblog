package tset

import (
	"os"
	"http"
	"template"
	"httputils"
	"appengine"
	"gorilla.googlecode.com/hg/gorilla/mux"
)

func urlFor(name string, vars ...string) string {
	route, _ := mux.NamedRoutes[name]
	url := route.URL(vars...)
	return url.Path
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
	s.Funcs(map[string]interface{}{"urlFor": urlFor})

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
