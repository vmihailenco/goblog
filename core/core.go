package core

import (
	"os"
	"fmt"
	"template"
	"runtime/debug"

	"gorilla.googlecode.com/hg/gorilla/mux"

	"tmplt"
)

var Router = &mux.Router{}

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

var Layout = tmplt.NewLayout("templates", "layout.html").SetFuncMap(template.FuncMap{"urlFor": urlFor})
