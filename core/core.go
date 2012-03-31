package core

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"runtime/debug"

	"appengine"
	"appengine/user"
	"code.google.com/p/gorilla/mux"
	"github.com/vmihailenco/gforms"

	"auth"
	"tmplt"
)

var Router = &mux.Router{}

var Layout *template.Template

func init() {
	Layout = template.New("layout.html")
	Layout = Layout.Funcs(template.FuncMap{
		"htmlSafe": htmlSafe,

		"urlFor":    urlFor,
		"loginUrl":  loginUrl,
		"logoutUrl": logoutUrl,

		"render":      gforms.Render,
		"renderLabel": gforms.RenderLabel,
		"renderError": gforms.RenderError,
	})

	var err error
	Layout, err = Layout.ParseFiles("templates/layout.html")
	if err != nil {
		panic(err)
	}

	Router.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	Router.HandleFunc("/500.html", InternalErrorHandler).Name("internalError")
	Router.HandleFunc("/profile/", TemplateHandler(Layout, "templates/profile.html")).Name("profile")

	http.Handle("/", NewProfilingHandler(Router))
}

func TemplateFuncRecover() {
	if err := recover(); err != nil {
		errStr := fmt.Sprint(err)
		fmt.Println(errStr)
		debug.PrintStack()
		panic(errors.New(errStr))
	}
}

func htmlSafe(text string) template.HTML {
	return template.HTML(text)
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
	url, err := Router.GetRoute(name).URL(strPairs...)
	if err != nil {
		return err.Error()
	}
	return url.String()
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

func RenderTemplate(c appengine.Context, w http.ResponseWriter, base *template.Template, context tmplt.Context, templateName ...string) {
	t := base
	var err error
	for _, name := range templateName {
		t, err = tmplt.Holder.Get(name, t)
		if err != nil {
			HandleError(c, w, err)
			return
		}
	}

	if context == nil {
		context = tmplt.Context{}
	}

	context["appengineContext"] = c
	context["user"] = auth.CurrentUser(c)

	w.Header().Add("content-type", "text/html")

	err = t.Execute(w, context)
	if err != nil {
		HandleError(c, w, err)
		return
	}
}
