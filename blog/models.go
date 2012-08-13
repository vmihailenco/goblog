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
	"core/page"
	"core/pager"
)

const (
	ARTICLE_KIND = "article"
)

func articleCacheKey(id int64) string {
	return fmt.Sprintf("blog-article-%v", id)
}

func NewArticlePager(c appengine.Context, q *datastore.Query, page int) *pager.Pager {
	return pager.NewPager(c, ARTICLE_KIND, q, page, 10)
}

type Article struct {
	*entity.Entity `datastore:"-"`

	Title     string
	TextBytes []byte
	HTMLBytes []byte

	ViewsCount int
	IsPublic   bool
	CreatedOn  time.Time
}

func NewArticle() *Article {
	return &Article{
		Entity: entity.NewEntity(ARTICLE_KIND),
	}
}

func NewArticleQuery() *datastore.Query {
	return datastore.NewQuery(ARTICLE_KIND)
}

func GetArticleById(c appengine.Context, id int64, useCache bool) (*Article, error) {
	article := NewArticle()
	articleCacheKey := articleCacheKey(id)

	if useCache {
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

func GetArticles(c appengine.Context, p *pager.Pager) ([]*Article, error) {
	q := p.Query()

	articles := make([]*Article, 0, p.PageSize)
	page, err := page.GetPage(c, q, p.PageSize, false, &articles)
	if err != nil {
		return nil, err
	}
	for i, key := range page.Keys {
		articles[i].SetKey(key)
	}
	p.Update(page.Start, page.More)
	return articles, nil
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
	return Router.GetRoute("article").URL(
		"id",
		strconv.FormatInt(a.Key().IntID(), 10),
		"slug",
		a.Slug(),
	)
}

func (a *Article) PermaURL() (*url.URL, error) {
	return Router.GetRoute("articlePermaLink").URL(
		"id",
		strconv.FormatInt(a.Key().IntID(), 10),
	)
}

func (a *Article) UpdateURL() (*url.URL, error) {
	return Router.GetRoute("articleUpdate").URL(
		"id",
		strconv.FormatInt(a.Key().IntID(), 10),
	)
}

func (a *Article) DeleteURL() (*url.URL, error) {
	return Router.GetRoute("articleDelete").URL(
		"id",
		strconv.FormatInt(a.Key().IntID(), 10),
	)
}

func ChangeArticleViewsCount(c appengine.Context, key *datastore.Key, delta int) error {
	article := NewArticle()
	return datastore.RunInTransaction(c, func(c appengine.Context) error {
		if err := datastore.Get(c, key, article); err != nil {
			return err
		}
		article.ViewsCount += delta
		if _, err := datastore.Put(c, key, article); err != nil {
			return err
		}
		return nil
	}, nil)
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

func UpdateArticle(c appengine.Context, article *Article, title string, text string, isPublic bool) error {
	textBytes := []byte(text)

	article.Title = title
	article.TextBytes = textBytes
	article.HTMLBytes = blackfriday.MarkdownCommon(textBytes)
	article.IsPublic = isPublic

	if err := entity.Put(c, article); err != nil {
		return err
	}

	memcache.Delete(c, articleCacheKey(article.Key().IntID()))

	return nil
}

func DeleteArticle(c appengine.Context, article *Article) error {
	err := datastore.Delete(c, article.Key())
	if err != nil {
		return err
	}
	err = memcache.Delete(c, articleCacheKey(article.Key().IntID()))
	if err != nil {
		return err
	}
	return nil
}
