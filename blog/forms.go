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
	title.SetMinLen(1)
	title.SetMaxLen(500)

	text := gforms.NewTextareaStringField()
	text.SetMinLen(1)

	isPublic := gforms.NewBoolField()
	isPublic.SetIsRequired(false)
	isPublic.SetLabel("Is public?")

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
