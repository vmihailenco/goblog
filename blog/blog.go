package blog

import (
	"html/template"
	"net/http"

	appengineSessions "code.google.com/p/gorilla/appengine/sessions"
	"code.google.com/p/gorilla/sessions"

	"core"
)

var Router = core.Router
var Layout *template.Template

func init() {
	var err error
	Layout, err = core.Layout.Clone()
	if err != nil {
		panic(err)
	}

	Router.HandleFunc("/article/create/", ArticleCreateHandler).Name("articleCreate")
	Router.HandleFunc("/article/update/{id:[0-9]+}/", ArticleUpdateHandler).Name("articleUpdate")
	Router.HandleFunc("/article/list/", ArticleListHandler).Name("articleList")
	Router.HandleFunc("/articles/{id:[0-9]+}/{slug:[0-9A-Za-z_-]+}/", ArticleHandler).Name("article")
	Router.HandleFunc("/markdown-preview/", MarkdownPreviewHandler).Name("markdownPreview")
	Router.HandleFunc("/about/", core.TemplateHandler(Layout, "templates/about.html")).Name("about")
	Router.HandleFunc("/", ArticleListHandler).Name("home")

	http.Handle("/", Router)

	// Register the datastore and memcache session stores.
	sessions.SetStore("datastore", new(appengineSessions.DatastoreSessionStore))
	sessions.SetStore("memcache", new(appengineSessions.MemcacheSessionStore))

	// Set secret keys for the session stores.
	sessions.SetStoreKeys("datastore",
		[]byte("my-secret-key"),
		[]byte("1234567890123456"))
	sessions.SetStoreKeys("memcache",
		[]byte("my-secret-key"),
		[]byte("1234567890123456"))
}
