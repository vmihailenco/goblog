package blog

import (
	"os"
	"time"
	"strconv"
	"appengine"
	"appengine/datastore"
	"core"
)

const (
	ARTICLE_KIND = "article"
)

type Article struct {
	key       *datastore.Key
	Title     string
	Text      string
	CreatedOn datastore.Time
}

func (a *Article) Key() *datastore.Key {
	return a.key
}

func (a *Article) SetKey(key *datastore.Key) {
	a.key = key
}

func (a *Article) URLString() string {
	return core.Router.NamedRoutes["article"].URL("id", strconv.Itoa64(a.Key().IntID())).String()
}

func NewArticle(c appengine.Context, title string, text string) (*Article, os.Error) {
	a := Article{
		key:       datastore.NewIncompleteKey(c, ARTICLE_KIND, nil),
		Title:     title,
		Text:      text,
		CreatedOn: datastore.SecondsToTime(time.Seconds()),
	}
	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "article", nil), &a)
	if err != nil {
		return nil, err
	}

	a.key = key
	return &a, nil
}

func GetArticleQuery() *datastore.Query {
	return datastore.NewQuery(ARTICLE_KIND)
}

func GetArticleById(c appengine.Context, id int64) (*Article, os.Error) {
	key := datastore.NewKey(c, "article", "", id, nil)
	article := &Article{}
	if err := datastore.Get(c, key, article); err != nil {
		return nil, err
	}
	return article, nil
}

func GetArticles(c appengine.Context, q *datastore.Query, limit int) ([]Article, os.Error) {
	q = q.Limit(limit)
	articles := make([]Article, 0, limit)
	for i, t := 0, q.Run(c); ; i++ {
		article := Article{}
		key, err := t.Next(&article)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		article.SetKey(key)
		articles = articles[0:i+1]
		articles[i] = article
	}
	return articles, nil
}