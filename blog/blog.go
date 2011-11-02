package blog

import (
	"http"
	appengineSessions "gorilla.googlecode.com/hg/gorilla/appengine/sessions"
	"gorilla.googlecode.com/hg/gorilla/sessions"
	"gorilla.googlecode.com/hg/gorilla/mux"
)

func init() {
	// Register a couple of routes.
	mux.HandleFunc("/article/create/", ArticleCreateHandler).Name("articleCreate")
	mux.HandleFunc("/article/list/", ArticleListHandler).Name("articleList")
	mux.HandleFunc("/articles/{id:[0-9]+}/", ArticleHandler).Name("article")
	mux.HandleFunc("/about/", AboutHandler).Name("about")
	mux.HandleFunc("/", ArticleListHandler).Name("home")

	// Send all incoming requests to mux.DefaultRouter.
	http.Handle("/", mux.DefaultRouter)

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
