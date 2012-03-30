package core

import (
	"encoding/json"
	"html/template"
	"net/http"

	"appengine"

	"tmplt"
)

var ErrLayout *template.Template

func init() {
	var err error
	ErrLayout, err = template.ParseFiles("templates/500.html")
	if err != nil {
		panic(err)
	}
}

func HandleNotFound(c appengine.Context, w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	RenderTemplate(c, w, Layout, nil, "templates/404.html")
}

func HandleError(c appengine.Context, w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)

	err2 := ErrLayout.Execute(w, tmplt.Context{"err": err})
	if err2 != nil {
		c.Criticalf("Got error %v while serving %v", err2, err)
		return
	}
}

func HandleAuthRequired(c appengine.Context, w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	RenderTemplate(c, w, Layout, nil, "templates/401.html")
}

func HandleJSON(c appengine.Context, w http.ResponseWriter, value interface{}) {
	w.Header().Add("content-type", "application/json")
	json.NewEncoder(w).Encode(value)
}
