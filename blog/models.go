package blog

import (
	"os"
	"time"
	"strconv"
	"url"

	"appengine"
	"appengine/datastore"

	"core"
	"core/entity"
)

const (
	ARTICLE_KIND = "article"
)

func GetArticleQuery() *datastore.Query {
	return datastore.NewQuery(ARTICLE_KIND)
}

func GetArticleById(c appengine.Context, id int64) (*Article, os.Error) {
	key := datastore.NewKey(c, "article", "", id, nil)
	article := &Article{}
	if err := datastore.Get(c, key, article); err != nil {
		return nil, err
	}
	article.SetKey(key)
	return article, nil
}

func GetArticles(c appengine.Context, q *datastore.Query, limit int) (*[]Article, os.Error) {
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

	CreatedOn datastore.Time
}

func (a *Article) SetKey(key *datastore.Key) os.Error {
	if a.Entity == nil {
		a.Entity = entity.NewEntity(ARTICLE_KIND)
	}
	return a.Entity.SetKey(key)
}

func (a *Article) URL() *url.URL {
	return core.URLFor("article", "id", strconv.Itoa64(a.Key().IntID()))
}

func NewArticle(c appengine.Context, title string, text string) (*Article, os.Error) {
	a := &Article{
		Entity:    entity.NewEntity(ARTICLE_KIND),
		Title:     title,
		Text:      text,
		CreatedOn: datastore.SecondsToTime(time.Seconds()),
	}
	if _, err := entity.PutEntity(c, a); err != nil {
		return nil, err
	}
	return a, nil
}
