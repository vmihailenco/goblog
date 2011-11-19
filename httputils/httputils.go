package httputils

import (
	"os"
	"bytes"
	"http"

	"appengine"

	"tmplt"
)

var layout = tmplt.NewLayout("templates", "500.html")

func ServeBuffer(c appengine.Context, w http.ResponseWriter, buf *bytes.Buffer, err os.Error) {
	if err != nil {
		HandleError(c, w, err)
		return
	}
	buf.WriteTo(w)
}

func HandleError(c appengine.Context, w http.ResponseWriter, err os.Error) {
	buf, err := layout.Render(tmplt.Context{"err": err}, "")
	if err != nil {
		c.Criticalf("%v", err)
	}
	ServeBuffer(c, w, buf, nil)
}
