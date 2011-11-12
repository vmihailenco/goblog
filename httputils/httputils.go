package httputils

import (
	"os"
	"bytes"
	"http"

	"appengine"

	"layout"
)

var l = layout.NewLayout("templates", "500.html")

func HandleError(c appengine.Context, w http.ResponseWriter, err os.Error) {
	buf, err := l.Render(layout.Context{"err": err}, "")
	if err != nil {
		c.Criticalf("%v", err)
	}
	ServeBuffer(c, w, buf, nil)
}

func ServeBuffer(c appengine.Context, w http.ResponseWriter, buf *bytes.Buffer, err os.Error) {
	if err != nil {
		HandleError(c, w, err)
		return
	}
	buf.WriteTo(w)
}
