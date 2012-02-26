package blog

import (
	"html/template"
	"net/http"
	"strconv"

	"appengine"
	"code.google.com/p/gorilla/mux"
	"code.google.com/p/gorilla/schema"

	"core"
	"httputils"
	"tmplt"
)

var Layout *template.Template

func init() {
	var err error
	Layout, err = core.Layout.Clone()
	if err != nil {
		panic(err)
	}

	Layout, err = Layout.ParseFiles("templates/blog/base.html")
	if err != nil {
		panic(err)
	}
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	core.RenderTemplate(c, w, Layout, nil, "templates/about.html")
}

func ArticleHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
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
	core.RenderTemplate(c, w, Layout, context, "templates/blog/article.html")
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
	core.RenderTemplate(c, w, Layout, context, "templates/blog/articleList.html")
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
		if err := schema.NewDecoder().Decode(form, r.Form); err != nil {
			httputils.HandleError(c, w, err)
			return
		}

		a, err := NewArticle(c, form.Title, form.Text)
		if err != nil {
			httputils.HandleError(c, w, err)
			return
		}

		redirect_to := core.URLFor(
			"article", "id", strconv.FormatInt(a.Key().IntID(), 10))
		http.Redirect(w, r, redirect_to.Path, 302)
	}

	core.RenderTemplate(c, w, Layout, nil, "templates/blog/articleCreate.html")
}
