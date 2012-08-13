package core

import (
	"fmt"
	"html/template"
	"time"

	"appengine"
	"appengine/blobstore"
	"appengine/user"
	"github.com/vmihailenco/gforms"

	"tmplt"
)

func AddTemplateFuncs(t *template.Template) *template.Template {
	return t.Funcs(template.FuncMap{
		"now":           now,
		"formatTime":    formatTime,
		"formatRFC3339": formatRFC3339,

		"htmlSafe": htmlSafe,

		"urlFor":             urlFor,
		"loginURL":           loginURL,
		"logoutURL":          logoutURL,
		"blobstoreUploadURL": blobstoreUploadURL,

		"render":      gforms.Render,
		"renderLabel": gforms.RenderLabel,
		"renderError": gforms.RenderError,
	})
}

func now() time.Time {
	return time.Now()
}

func formatTime(format string, tm time.Time) string {
	return tm.Format(format)
}

func formatRFC3339(tm time.Time) string {
	return tm.Format(time.RFC3339)
}

func htmlSafe(text string) template.HTML {
	return template.HTML(text)
}

func urlFor(name string, pairs ...interface{}) string {
	size := len(pairs)
	strPairs := make([]string, size)
	for i := 0; i < size; i++ {
		if v, ok := pairs[i].(string); ok {
			strPairs[i] = v
		} else {
			strPairs[i] = fmt.Sprint(pairs[i])
		}
	}
	url, err := Router.GetRoute(name).URL(strPairs...)
	if err != nil {
		return err.Error()
	}
	return url.String()
}

func loginURL(context tmplt.Context, redirectTo string) (string, error) {
	c := context["appengineContext"].(appengine.Context)
	return user.LoginURL(c, redirectTo)
}

func logoutURL(context tmplt.Context, redirectTo string) (string, error) {
	c := context["appengineContext"].(appengine.Context)
	return user.LogoutURL(c, redirectTo)
}

func blobstoreUploadURL(context tmplt.Context, url string) (string, error) {
	c := context["appengineContext"].(appengine.Context)
	uploadURL, err := blobstore.UploadURL(c, url, nil)
	if err != nil {
		return "", err
	}
	return uploadURL.Path, nil
}
