package blog

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"github.com/russross/blackfriday"

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

	Title     string
	TextBytes []byte
	HTMLBytes []byte

	IsPublic  bool
	CreatedOn time.Time
}

func (a *Article) Text() string {
	return string(a.TextBytes)
}

func (a *Article) HTML() string {
	return string(a.HTMLBytes)
}

func (a *Article) SetKey(key *datastore.Key) {
	if a.Entity == nil {
		a.Entity = entity.NewEntity(ARTICLE_KIND)
	}
	a.Entity.SetKey(key)
}

var slugRe = regexp.MustCompile("[^0-9A-Za-z_-]+")

func (a *Article) Slug() string {
	return strings.ToLower(slugRe.ReplaceAllLiteralString(a.Title, "-"))
}

func (a *Article) URL() (*url.URL, error) {
	return Router.GetRoute("article").URL("id", strconv.FormatInt(a.Key().IntID(), 10), "slug", a.Slug())
}

func (a *Article) UpdateURL() (*url.URL, error) {
	return Router.GetRoute("articleUpdate").URL("id", strconv.FormatInt(a.Key().IntID(), 10))
}

func CreateArticle(c appengine.Context, title string, text string, isPublic bool) (*Article, error) {
	a := &Article{
		Entity:    entity.NewEntity(ARTICLE_KIND),
		CreatedOn: time.Now(),
	}
	if err := UpdateArticle(c, a, title, text, isPublic); err != nil {
		return nil, err
	}
	return a, nil
}

func UpdateArticle(c appengine.Context, a *Article, title string, text string, isPublic bool) error {
	textBytes := []byte(text)

	a.Title = title
	a.TextBytes = textBytes
	a.HTMLBytes = blackfriday.MarkdownCommon(textBytes)
	a.IsPublic = isPublic

	if err := entity.Put(c, a); err != nil {
		return err
	}
	return nil
}
