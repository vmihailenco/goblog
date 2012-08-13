package blog

import (
	"core"
)

const (
	LAYOUT = "templates/layout.html"
)

var (
	Router = core.Router
)

func init() {
	Router.HandleFunc("/article/create/", ArticleCreateHandler).Name("articleCreate")
	Router.HandleFunc("/article/update/{id:[0-9]+}/", ArticleUpdateHandler).Name("articleUpdate")
	Router.HandleFunc("/article/delete/{id:[0-9]+}/", ArticleDeleteHandler).Name("articleDelete")
	Router.HandleFunc("/article/page/{page:[0-9]+}/", ArticlePageHandler).Name("articlePage")
	Router.HandleFunc("/articles/{id:[0-9]+}/", ArticlePermaLinkHandler).Name("articlePermaLink")
	Router.HandleFunc("/articles/{id:[0-9]+}/{slug:[0-9A-Za-z_-]+}/", ArticleHandler).Name("article")
	Router.HandleFunc("/feed/", ArticleFeedHandler).Name("articleFeed")
	Router.HandleFunc("/markdown-preview/", MarkdownPreviewHandler).Name("markdownPreview")
	Router.HandleFunc("/about/", core.TemplateHandler("templates/layout.html", "templates/about.html")).Name("about")
	Router.HandleFunc("/", ArticlePageHandler).Name("home")

	Router.HandleFunc("/image-upload/url/", ImageUploadURLHandler).Name("imageUploadURL")
	Router.HandleFunc("/image-upload/", ImageUploadHandler).Name("imageUpload")
}
