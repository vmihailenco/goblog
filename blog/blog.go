package blog

import (
	"html/template"

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
}
