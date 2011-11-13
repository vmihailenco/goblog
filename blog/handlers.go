package blog

import (
	"strconv"
	"http"
	"appengine"
	"gorilla.googlecode.com/hg/gorilla/mux"
	"gorilla.googlecode.com/hg/gorilla/schema"
	"core"
	"httputils"
	"layout"
)

var l = layout.NewLayout("templates", "layout.html")

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	buf, err := l.Render(nil, "about.html")
	httputils.ServeBuffer(c, w, buf, err)
}

func ArticleHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	vars := mux.Vars(r)
	id, err := strconv.Atoi64(vars["id"])
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}

	article, err := GetArticleById(c, id)
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}

	buf, err := l.Render(layout.Context{"article": article}, "blog/article.html")
	httputils.ServeBuffer(c, w, buf, err)
}

func ArticleListHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	q := GetArticleQuery().Order("-CreatedOn")
	articles, err := GetArticles(c, q, 20)
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}
	buf, err := l.Render(layout.Context{"articles": articles}, "blog/article_list.html")
	httputils.ServeBuffer(c, w, buf, err)
}

type ArticleForm struct {
	Title string
	Text  string
}

func ArticleCreateHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			httputils.HandleError(c, w, err)
			return
		}

		form := &ArticleForm{}
		if err := schema.Load(form, r.Form); err != nil {
			httputils.HandleError(c, w, err)
			return
		}

		a, err := NewArticle(c, form.Title, form.Text)
		if err != nil {
			httputils.HandleError(c, w, err)
			return
		}

		redirect_to := core.Router.NamedRoutes["article"].URL("id", strconv.Itoa64(a.Key().IntID()))
		http.Redirect(w, r, redirect_to.Path, 302)
	}

	buf, err := l.Render(nil, "blog/article_create.html")
	httputils.ServeBuffer(c, w, buf, err)
}
