package httputils

import (
	"html/template"
	"net/http"

	"appengine"

	"tmplt"
)

var Layout *template.Template

func init() {
	var err error
	Layout, err = template.ParseFiles("templates/500.html")
	if err != nil {
		panic(err)
	}
}

func HandleError(c appengine.Context, w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)

	err2 := Layout.Execute(w, tmplt.Context{"err": err})
	if err2 != nil {
		c.Criticalf("Got error %v while serving %v", err2, err)
		return
	}
}
