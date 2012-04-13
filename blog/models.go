package blog

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/memcache"
	"github.com/russross/blackfriday"

	"core/entity"
)

const (
	ARTICLE_KIND = "article"
)

func articleCacheKey(id int64) string {
	return fmt.Sprintf("blog-article-%v", id)
}

type Article struct {
	*entity.Entity `datastore:"-"`

	Title     string
	TextBytes []byte
	HTMLBytes []byte

	IsPublic  bool
	CreatedOn time.Time
}

func NewArticle() *Article {
	return &Article{
		Entity: entity.NewEntity(ARTICLE_KIND),
	}
}

func NewArticleQuery() *datastore.Query {
	return datastore.NewQuery(ARTICLE_KIND)
}

func GetArticleById(c appengine.Context, id int64) (*Article, error) {
	article := NewArticle()
	articleCacheKey := articleCacheKey(id)

	if item, err := memcache.Get(c, articleCacheKey); err == nil {
		dec := gob.NewDecoder(bytes.NewBuffer(item.Value))
		err = dec.Decode(article)
		if err == nil {
			return article, nil
		} else {
			c.Errorf("error decoding article: %v", err)
		}
	} else if err != memcache.ErrCacheMiss {
		c.Errorf("error getting item: %v", err)
	}

	key := datastore.NewKey(c, "article", "", id, nil)
	if err := datastore.Get(c, key, article); err != nil {
		return nil, err
	}
	article.SetKey(key)

	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(article); err == nil {
		item := &memcache.Item{
			Key:        articleCacheKey,
			Value:      buf.Bytes(),
			Expiration: time.Duration(24) * time.Hour,
		}
		if err := memcache.Set(c, item); err != nil {
			c.Errorf("error setting item: %v", err)
		}
	} else {
		c.Errorf("error encoding article: %v", err)
	}

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
	return Router.GetRoute("article").URL("id", strconv.FormatInt(a.Key().IntID(), 10),
		"slug", a.Slug())
}

func (a *Article) PermaURL() (*url.URL, error) {
	return Router.GetRoute("articlePermaLink").URL("id", strconv.FormatInt(a.Key().IntID(), 10))
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

	memcache.Delete(c, articleCacheKey(a.Key().IntID()))

	return nil
}
