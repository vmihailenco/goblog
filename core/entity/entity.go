package entity

import (
	"errors"

	"appengine"
	"appengine/datastore"
)

type Putable interface {
	Kind() string
	SetKey(*datastore.Key) error
}

type Entity struct {
	key  *datastore.Key
	kind string
}

func NewEntity(kind string) *Entity {
	return &Entity{kind: kind}
}

func (e *Entity) SetKey(key *datastore.Key) error {
	if e.key != nil {
		return errors.New("entity already has a key")
	}
	e.key = key
	return nil
}

func (e *Entity) Key() *datastore.Key {
	return e.key
}

func (e *Entity) Kind() string {
	return e.kind
}

func PutEntity(c appengine.Context, e Putable) (interface{}, error) {
	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, e.Kind(), nil), e)
	if err != nil {
		return nil, err
	}
	if err = e.SetKey(key); err != nil {
		return nil, err
	}
	return e, nil
}
