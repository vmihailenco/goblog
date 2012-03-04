package blog

import (
	"gforms"
)

type ArticleForm struct {
	*gforms.BaseForm
	Title *gforms.StringField
	Text  *gforms.StringField
}

func NewArticleForm() *ArticleForm {
	f := &ArticleForm{
		BaseForm: &gforms.BaseForm{},
		Title:    gforms.NewStringField(true, 1, 500),
		Text:     gforms.NewTextareaField(true, 1, 0),
	}
	gforms.InitForm(f)
	return f
}
