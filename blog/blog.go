package blog

import (
	"http"
	appengineSessions "gorilla.googlecode.com/hg/gorilla/appengine/sessions"
	"gorilla.googlecode.com/hg/gorilla/sessions"
	"core"
)

var Router = core.Router

func init() {
	// Register a couple of routes.
	Router.HandleFunc("/article/create/", ArticleCreateHandler).Name("articleCreate")
	Router.HandleFunc("/article/list/", ArticleListHandler).Name("articleList")
	Router.HandleFunc("/articles/{id:[0-9]+}/", ArticleHandler).Name("article")
	Router.HandleFunc("/about/", AboutHandler).Name("about")
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
