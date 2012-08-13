package blog

import (
	"github.com/vmihailenco/gforms"
)

type ArticleForm struct {
	*gforms.BaseForm
	Title    *gforms.StringField
	Text     *gforms.StringField
	IsPublic *gforms.BoolField
}

func NewArticleForm(article *Article) *ArticleForm {
	title := gforms.NewStringField()
	title.MinLen = 1
	title.MaxLen = 500

	text := gforms.NewTextareaStringField()
	text.MinLen = 1

	isPublic := gforms.NewBoolField()
	isPublic.IsRequired = false
	isPublic.Label = "Is public?"

	if article != nil {
		title.SetInitial(article.Title)
		text.SetInitial(article.Text())
		isPublic.SetInitial(article.IsPublic)
	}

	f := &ArticleForm{
		BaseForm: &gforms.BaseForm{},
		Title:    title,
		Text:     text,
		IsPublic: isPublic,
	}
	gforms.InitForm(f)

	return f
}
