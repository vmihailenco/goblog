package core

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"appengine"
)

func TemplateHandler(templateNames ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := appengine.NewContext(r)
		RenderTemplate(c, w, nil, templateNames...)
	}
}

func InternalErrorHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	HandleError(c, w, errors.New("empty"))
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	w.WriteHeader(http.StatusNotFound)
	RenderTemplate(c, w, nil, "templates/404.html", LAYOUT)
}

type ProfilingHandler struct {
	handler http.Handler
}

func NewProfilingHandler(handler http.Handler) *ProfilingHandler {
	return &ProfilingHandler{handler}
}

func (r *ProfilingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	t0 := time.Now()
	r.handler.ServeHTTP(w, req)
	t1 := time.Now()

	if w.Header().Get("content-type") == "text/html" {
		fmt.Fprintf(w, "<!-- %v -->", t1.Sub(t0))
	}
}
