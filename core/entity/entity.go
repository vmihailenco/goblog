package entity

import (
	"appengine"
	"appengine/datastore"
)

type Putable interface {
	Kind() string
	SetKey(*datastore.Key)
	Key() *datastore.Key
}

type Entity struct {
	// key is public to simplify gob encoding
	DsKey *datastore.Key
	kind  string
}

func NewEntity(kind string) *Entity {
	return &Entity{kind: kind}
}

func (e *Entity) SetKey(key *datastore.Key) {
	if e.DsKey != nil {
		panic("Entity already has a key.")
	}
	e.DsKey = key
}

func (e *Entity) Key() *datastore.Key {
	return e.DsKey
}

func (e *Entity) Kind() string {
	if e.DsKey != nil {
		return e.DsKey.Kind()
	}
	return e.kind
}

func Put(c appengine.Context, entity Putable) error {
	key := entity.Key()
	if key == nil {
		key = datastore.NewIncompleteKey(c, entity.Kind(), nil)
	}

	key, err := datastore.Put(c, key, entity)
	if err != nil {
		return err
	}

	if entity.Key() == nil {
		entity.SetKey(key)
	}

	return nil
}
