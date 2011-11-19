package core

import (
	"os"
	"fmt"
	"template"
	"runtime/debug"
	"http"
	"url"

	"appengine"
	"appengine/user"
	"gorilla.googlecode.com/hg/gorilla/mux"

	"httputils"
	"tmplt"
	"auth"
)

var Router = &mux.Router{}

func URLFor(name string, pairs ...string) *url.URL {
	return Router.NamedRoutes[name].URL(pairs...)
}

func TemplateFuncRecover() {
	if err := recover(); err != nil {
		errStr := fmt.Sprint(err)
		fmt.Println(errStr)
		debug.PrintStack()
		panic(os.NewError(errStr))
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
	route, _ := Router.NamedRoutes[name]
	return route.URL(strPairs...).String()
}

func loginUrl(context tmplt.Context, redirectTo string) (string, os.Error) {
	defer TemplateFuncRecover()
	c := context["appengineContext"].(appengine.Context)
	return user.LoginURL(c, redirectTo)
}

func logoutUrl(context tmplt.Context, redirectTo string) (string, os.Error) {
	defer TemplateFuncRecover()
	c := context["appengineContext"].(appengine.Context)
	return user.LogoutURL(c, redirectTo)
}

func isAdmin(context tmplt.Context) bool {
	defer TemplateFuncRecover()
	c := context["appengineContext"].(appengine.Context)
	return user.IsAdmin(c)
}

var Layout = tmplt.NewLayout("templates", "layout.html").
	SetFuncMap(template.FuncMap{
	"urlFor":    urlFor,
	"loginUrl":  loginUrl,
	"logoutUrl": logoutUrl})

func RenderTemplate(
	c appengine.Context,
	w http.ResponseWriter,
	l *tmplt.Layout,
	context tmplt.Context,
	filename string) {
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

	buf, err := l.Render(context, filename)
	httputils.ServeBuffer(c, w, buf, err)
}
