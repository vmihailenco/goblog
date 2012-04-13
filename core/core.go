package core

import (
	"html/template"
	"net/http"
	"path"

	"appengine"
	"code.google.com/p/gorilla/mux"

	"auth"
	"tmplt"
)

const (
	LAYOUT = "templates/layout.html"
)

var (
	Router = &mux.Router{}
)

func init() {
	Router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	Router.HandleFunc("/500.html", InternalErrorHandler).Name("internalError")
	Router.HandleFunc("/profile/", TemplateHandler("templates/layout.html", "templates/profile.html")).Name("profile")

	http.Handle("/", NewProfilingHandler(Router))
}

func RenderTemplate(c appengine.Context, w http.ResponseWriter, context tmplt.Context, templateNames ...string) {
	var (
		t   *template.Template
		err error
	)

	if len(templateNames) == 0 {
		panic("expected at least 1 template, but got 0")
	}

	for _, name := range templateNames {
		newFunc := func(filename string) (*template.Template, error) {
			var newT *template.Template
			if t == nil {
				newT = AddTemplateFuncs(template.New(path.Base(filename)))
			} else {
				newT, err = t.Clone()
				if err != nil {
					return nil, err
				}
			}
			newT, err = newT.ParseFiles(filename)
			if err != nil {
				return nil, err
			}
			return newT, nil
		}

		t, err = tmplt.Holder.Get(name, newFunc)
		if err != nil {
			HandleError(c, w, err)
			return
		}
	}

	if context == nil {
		context = tmplt.Context{}
	}

	context["appengineContext"] = c
	context["user"] = auth.CurrentUser(c)

	if w.Header().Get("content-type") == "" {
		w.Header().Add("content-type", "text/html")
	}

	err = t.Execute(w, context)
	if err != nil {
		HandleError(c, w, err)
		return
	}
}
