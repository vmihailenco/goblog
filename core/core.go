package core

import (
	"bytes"
	"html/template"
	"io"
	"net/http"
	"path"
	"strings"

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
	Router.HandleFunc("/profile/", TemplateHandler("templates/profile.html", "templates/layout.html")).Name("profile")

	http.Handle("/", NewProfilingHandler(Router))
}

func RenderTemplate(c appengine.Context, w http.ResponseWriter, context tmplt.Context, templateNames ...string) {
	if len(templateNames) == 0 {
		panic("expected at least 1 template, but got 0")
	}

	newFunc := func() (*template.Template, error) {
		var err error

		newT := template.New(path.Base(templateNames[len(templateNames)-1]))
		newT = AddTemplateFuncs(newT)
		newT, err = newT.ParseFiles(templateNames...)
		if err != nil {
			return nil, err
		}
		return newT, nil
	}

	t, err := tmplt.Holder.Get(strings.Join(templateNames, ","), newFunc)
	if err != nil {
		HandleError(c, w, err)
		return
	}

	if context == nil {
		context = tmplt.Context{}
	}

	context["appengineContext"] = c
	context["user"] = auth.CurrentUser(c)

	if w.Header().Get("content-type") == "" {
		w.Header().Add("content-type", "text/html")
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, context)
	if err != nil {
		HandleError(c, w, err)
		return
	}

	io.Copy(w, buf)
}
