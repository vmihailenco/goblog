package core

import (
	"errors"
	"html/template"
	"net/http"

	"appengine"
)

func TemplateHandler(layout *template.Template, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		RenderTemplate(c, w, layout, nil, name)
	}
}

func InternalErrorHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	HandleError(c, w, errors.New("empty"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	w.WriteHeader(http.StatusNotFound)
	RenderTemplate(c, w, Layout, nil, "templates/404.html")
}
