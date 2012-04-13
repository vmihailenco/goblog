package core

import (
	"fmt"
	"html/template"
	"time"

	"appengine"
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

		"urlFor":    urlFor,
		"loginUrl":  loginUrl,
		"logoutUrl": logoutUrl,

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

func loginUrl(context tmplt.Context, redirectTo string) (string, error) {
	c := context["appengineContext"].(appengine.Context)
	return user.LoginURL(c, redirectTo)
}

func logoutUrl(context tmplt.Context, redirectTo string) (string, error) {
	c := context["appengineContext"].(appengine.Context)
	return user.LogoutURL(c, redirectTo)
}

func isAdmin(context tmplt.Context) bool {
	c := context["appengineContext"].(appengine.Context)
	return user.IsAdmin(c)
}
