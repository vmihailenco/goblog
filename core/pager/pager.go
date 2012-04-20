package pager

import (
	"errors"
	"fmt"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/memcache"
)

type Pager struct {
	Page     int
	PageSize int

	context     appengine.Context
	cachePrefix string
	query       *datastore.Query
	hasMore     bool
}

func NewPager(c appengine.Context, cachePrefix string, q *datastore.Query, page int, pageSize int) *Pager {
	if page < 1 {
		page = 1
	}
	return &Pager{
		Page:        page,
		PageSize:    pageSize,
		context:     c,
		cachePrefix: cachePrefix,
		query:       q,
		hasMore:     true,
	}
}

func (p *Pager) cacheKey(page int) string {
	return p.cachePrefix + fmt.Sprintf("-%d-%d", page, p.PageSize)
}

func (p *Pager) HasPrev() bool {
	return p.Page > 1
}

func (p *Pager) PrevPage() int {
	if p.HasPrev() {
		return p.Page - 1
	}
	return 1
}

func (p *Pager) HasNext() bool {
	return p.hasMore
}

func (p *Pager) NextPage() int {
	if p.HasNext() {
		return p.Page + 1
	}
	return p.Page
}

func (p *Pager) Update(cursor datastore.Cursor, hasMore bool) {
	p.hasMore = hasMore
	item := &memcache.Item{
		Key:        p.cacheKey(p.Page + 1),
		Value:      []byte(cursor.String()),
		Expiration: time.Duration(24) * time.Hour,
	}
	if err := memcache.Set(p.context, item); err != nil {
		p.context.Errorf("error setting item: %v", err)
	}
}

func (p *Pager) Cursor() (*datastore.Cursor, error) {
	if item, err := memcache.Get(p.context, p.cacheKey(p.Page)); err == nil {
		cursor, err := datastore.DecodeCursor(string(item.Value))
		if err != nil {
			return nil, err
		}
		return &cursor, nil
	}
	return nil, errors.New("cursor not found")
}

func (p *Pager) Query() *datastore.Query {
	// there is no cursor for first page
	if p.Page != 1 {
		if c, err := p.Cursor(); err == nil {
			return p.query.Start(*c)
		}
	}
	return p.query.Offset(p.PageSize * (p.Page - 1)).Limit(p.PageSize)
}
