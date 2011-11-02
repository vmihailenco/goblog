package httputils

import (
	"os"
	"http"
	"appengine"
)

func HandleError(c appengine.Context, w http.ResponseWriter, err os.Error) {
	http.Error(w, err.String(), http.StatusBadRequest)
}
