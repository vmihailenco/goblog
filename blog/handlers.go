package blog

import (
	"strconv"
	"http"

	"appengine"
	"gorilla.googlecode.com/hg/gorilla/mux"
	"gorilla.googlecode.com/hg/gorilla/schema"

	"core"
	"tmplt"
	"httputils"
)

var Layout = core.Layout.NewLayout().SetFilenames("blog/base.html")

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	core.RenderTemplate(c, w, nil, "about.html")
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

	context := tmplt.Context{"article": article}
	core.RenderTemplate(c, w, context, "blog/article.html")
}

func ArticleListHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	q := GetArticleQuery().Order("-CreatedOn")
	articles, err := GetArticles(c, q, 20)
	if err != nil {
		httputils.HandleError(c, w, err)
		return
	}

	context := tmplt.Context{"articles": articles}
	core.RenderTemplate(c, w, context, "blog/article_list.html")
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

		redirect_to := core.URLFor(
			"article", "id", strconv.Itoa64(a.Key().IntID()))
		http.Redirect(w, r, redirect_to.Path, 302)
	}

	core.RenderTemplate(c, w, nil, "blog/article_create.html")
}
