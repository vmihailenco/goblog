package blog

import (
	"net/url"
	"strconv"
	"time"

	"appengine"
	"appengine/datastore"

	"core/entity"
)

const (
	ARTICLE_KIND = "article"
)

func GetArticleQuery() *datastore.Query {
	return datastore.NewQuery(ARTICLE_KIND)
}

func GetArticleById(c appengine.Context, id int64) (*Article, error) {
	key := datastore.NewKey(c, "article", "", id, nil)
	article := &Article{}
	if err := datastore.Get(c, key, article); err != nil {
		return nil, err
	}
	article.SetKey(key)
	return article, nil
}

func GetArticles(c appengine.Context, q *datastore.Query, limit int) (*[]Article, error) {
	q = q.Limit(limit)
	articles := make([]Article, 0, limit)
	keys, err := q.GetAll(c, &articles)
	if err != nil {
		return nil, err
	}
	for i, key := range keys {
		articles[i].SetKey(key)
	}
	return &articles, nil
}

type Article struct {
	*entity.Entity `datastore:"-"`

	Title string
	Text  string

	CreatedOn time.Time
}

func (a *Article) SetKey(key *datastore.Key) error {
	if a.Entity == nil {
		a.Entity = entity.NewEntity(ARTICLE_KIND)
	}
	return a.Entity.SetKey(key)
}

func (a *Article) URL() (*url.URL, error) {
	return Router.GetRoute("article").URL("id", strconv.FormatInt(a.Key().IntID(), 10))
}

func NewArticle(c appengine.Context, title string, text string) (*Article, error) {
	a := &Article{
		Entity:    entity.NewEntity(ARTICLE_KIND),
		Title:     title,
		Text:      text,
		CreatedOn: time.Now(),
	}
	if _, err := entity.PutEntity(c, a); err != nil {
		return nil, err
	}
	return a, nil
}
