package blog

import (
	"net/http"
	"strconv"

	"appengine"
	"code.google.com/p/gorilla/mux"
	"github.com/russross/blackfriday"
	"github.com/vmihailenco/gforms"

	"auth"
	"core"
	"tmplt"
)

func ArticleHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		core.HandleError(c, w, err)
		return
	}

	article, err := GetArticleById(c, id)
	if err != nil {
		core.HandleNotFound(c, w)
		return
	}

	if !article.IsPublic {
		user := auth.CurrentUser(c)
		if !user.IsAdmin {
			core.HandleAuthRequired(c, w)
			return
		}
	}

	context := tmplt.Context{"article": article}
	core.RenderTemplate(c, w, Layout, context, "templates/blog/article.html")
}

func ArticleListHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	user := auth.CurrentUser(c)

	q := GetArticleQuery().Order("-CreatedOn")
	if !user.IsAdmin {
		q = q.Filter("IsPublic=", true)
	}

	articles, err := GetArticles(c, q, 20)
	if err != nil {
		core.HandleError(c, w, err)
		return
	}

	context := tmplt.Context{"articles": articles}
	core.RenderTemplate(c, w, Layout, context, "templates/blog/articleList.html")
}

func ArticleCreateHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	user := core.AdminUser(c, w)
	if user == nil {
		return
	}

	form := NewArticleForm(nil)

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			core.HandleError(c, w, err)
			return
		}

		if gforms.IsFormValid(form, r.Form) {
			article, err := CreateArticle(c,
				form.Title.Value(),
				form.Text.Value(),
				form.IsPublic.Value(),
			)
			if err != nil {
				core.HandleError(c, w, err)
				return
			}

			redirectTo, err := article.URL()
			if err != nil {
				core.HandleError(c, w, err)
				return
			}
			http.Redirect(w, r, redirectTo.Path, 302)
		}
	}

	context := map[string]interface{}{
		"form": form,
	}
	core.RenderTemplate(c, w, Layout, context, "templates/blog/articleForm.html", "templates/blog/articleCreate.html")
}

func ArticleUpdateHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	user := core.AdminUser(c, w)
	if user == nil {
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		core.HandleError(c, w, err)
		return
	}

	article, err := GetArticleById(c, id)
	if err != nil {
		core.HandleNotFound(c, w)
		return
	}

	form := NewArticleForm(article)

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			core.HandleError(c, w, err)
			return
		}

		if gforms.IsFormValid(form, r.Form) {
			err := UpdateArticle(c, article,
				form.Title.Value(),
				form.Text.Value(),
				form.IsPublic.Value(),
			)
			if err != nil {
				core.HandleError(c, w, err)
				return
			}

			redirectTo, err := article.URL()
			if err != nil {
				core.HandleError(c, w, err)
				return
			}
			http.Redirect(w, r, redirectTo.Path, 302)
		}
	}

	context := map[string]interface{}{
		"article": article,
		"form":    form,
	}
	core.RenderTemplate(c, w, Layout, context, "templates/blog/articleForm.html", "templates/blog/articleUpdate.html")
}

func MarkdownPreviewHandler(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	user := core.AdminUser(c, w)
	if user == nil {
		return
	}

	html := string(blackfriday.MarkdownCommon([]byte(r.FormValue("text"))))
	core.HandleJSON(c, w, map[string]string{"html": html})
}
