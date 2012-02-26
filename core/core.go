package core

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"runtime/debug"

	"appengine"
	"appengine/user"
	"code.google.com/p/gorilla/mux"

	"auth"
	"httputils"
	"tmplt"
)

var Router = &mux.Router{}

var Layout *template.Template

func init() {
	Layout = template.New("layout.html")
	Layout = Layout.Funcs(template.FuncMap{
		"urlFor":    urlFor,
		"loginUrl":  loginUrl,
		"logoutUrl": logoutUrl})

	var err error
	Layout, err = Layout.ParseFiles("templates/layout.html")
	if err != nil {
		panic(err)
	}
}

func URLFor(name string, pairs ...string) *url.URL {
	url, _ := Router.GetRoute(name).URL(pairs...)
	return url
}

func TemplateFuncRecover() {
	if err := recover(); err != nil {
		errStr := fmt.Sprint(err)
		fmt.Println(errStr)
		debug.PrintStack()
		panic(errors.New(errStr))
	}
}

func urlFor(name string, pairs ...interface{}) string {
	defer TemplateFuncRecover()

	size := len(pairs)
	strPairs := make([]string, size)
	for i := 0; i < size; i++ {
		if v, ok := pairs[i].(string); ok {
			strPairs[i] = v
		} else {
			strPairs[i] = fmt.Sprint(pairs[i])
		}
	}
	return URLFor(name, strPairs...).String()
}

func loginUrl(context tmplt.Context, redirectTo string) (string, error) {
	defer TemplateFuncRecover()
	c := context["appengineContext"].(appengine.Context)
	return user.LoginURL(c, redirectTo)
}

func logoutUrl(context tmplt.Context, redirectTo string) (string, error) {
	defer TemplateFuncRecover()
	c := context["appengineContext"].(appengine.Context)
	return user.LogoutURL(c, redirectTo)
}

func isAdmin(context tmplt.Context) bool {
	defer TemplateFuncRecover()
	c := context["appengineContext"].(appengine.Context)
	return user.IsAdmin(c)
}

func RenderTemplate(c appengine.Context, w http.ResponseWriter, base *template.Template, context tmplt.Context, filename string) {
	t, err := tmplt.Holder.Lookup(filename, base)
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}

	if context == nil {
		context = tmplt.Context{}
	}
	context["appengineContext"] = c

	u, err := auth.CurrentUser(c)
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}
	context["user"] = u

	err = t.Execute(w, context)
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}
}
