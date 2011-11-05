package blog

import (
	"strconv"
	"http"
	"appengine"
	"gorilla.googlecode.com/hg/gorilla/mux"
	"gorilla.googlecode.com/hg/gorilla/schema"
	"core"
	"httputils"
	"tset"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	tset.RenderTemplate(c, w, "about.html", nil)
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

	tset.RenderTemplate(c, w, "blog/article.html", tset.Context{"article": article})
}

func ArticleListHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	q := GetArticleQuery().Order("-CreatedOn")
	articles, err := GetArticles(c, q, 20)
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}
	tset.RenderTemplate(c, w, "blog/article_list.html", tset.Context{"articles": articles})
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

	tset.RenderTemplate(c, w, "blog/article_create.html", nil)
}