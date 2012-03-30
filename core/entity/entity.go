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
	key  *datastore.Key
	kind string
}

func NewEntity(kind string) *Entity {
	return &Entity{kind: kind}
}

func (e *Entity) SetKey(key *datastore.Key) {
	if e.key != nil {
		panic("Entity already has a key.")
	}
	e.key = key
}

func (e *Entity) Key() *datastore.Key {
	return e.key
}

func (e *Entity) Kind() string {
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
