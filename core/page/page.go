package page

// https://gist.github.com/2370781

import (
	"reflect"

	"appengine"
	"appengine/datastore"
)

var (
	typeOfPropertyLoadSaver = reflect.TypeOf((*datastore.PropertyLoadSaver)(nil)).Elem()
	typeOfPropertyList      = reflect.TypeOf(datastore.PropertyList(nil))
)

type multiArgType int

const (
	multiArgTypeInvalid multiArgType = iota
	multiArgTypePropertyLoadSaver
	multiArgTypeStruct
	multiArgTypeStructPtr
	multiArgTypeInterface
)

// checkMultiArg checks that v has type []S, []*S, []I, or []P, for some struct
// type S, for some interface type I, or some non-interface non-pointer type P
// such that P or *P implements PropertyLoadSaver.
//
// It returns what category the slice's elements are, and the reflect.Type
// that represents S, I or P.
//
// As a special case, PropertyList is an invalid type for v.
func checkMultiArg(v reflect.Value) (m multiArgType, elemType reflect.Type) {
	if v.Kind() != reflect.Slice {
		return multiArgTypeInvalid, nil
	}
	if v.Type() == typeOfPropertyList {
		return multiArgTypeInvalid, nil
	}
	elemType = v.Type().Elem()
	if reflect.PtrTo(elemType).Implements(typeOfPropertyLoadSaver) {
		return multiArgTypePropertyLoadSaver, elemType
	}
	switch elemType.Kind() {
	case reflect.Struct:
		return multiArgTypeStruct, elemType
	case reflect.Interface:
		return multiArgTypeInterface, elemType
	case reflect.Ptr:
		elemType = elemType.Elem()
		if elemType.Kind() == reflect.Struct {
			return multiArgTypeStructPtr, elemType
		}
	}
	return multiArgTypeInvalid, nil
}

// ----------------------------------------------------------------------------

type Page struct {
	Keys  []*datastore.Key
	Start datastore.Cursor
	More  bool
}

func GetPage(c appengine.Context, query *datastore.Query, limit int, keysOnly bool, dst interface{}) (*Page, error) {
	var dv reflect.Value
	var mat multiArgType
	var elemType reflect.Type

	if !keysOnly {
		dv = reflect.ValueOf(dst)
		if dv.Kind() != reflect.Ptr || dv.IsNil() {
			return nil, datastore.ErrInvalidEntityType
		}
		dv = dv.Elem()
		mat, elemType = checkMultiArg(dv)
		if mat == multiArgTypeInvalid || mat == multiArgTypeInterface {
			return nil, datastore.ErrInvalidEntityType
		}
	}

	var keys []*datastore.Key
	var cursor datastore.Cursor

	query = query.Limit(limit + 1)
	if keysOnly {
		query = query.KeysOnly()
	}
	t := query.Run(c)
	more := true
	for i := 0; i < limit; i++ {
		var ev reflect.Value
		if !keysOnly {
			ev = reflect.New(elemType)
		}
		k, err := t.Next(ev.Interface())
		if err == datastore.Done {
			more = false
			break
		}
		if err != nil {
			return nil, err
		}
		if !keysOnly {
			if mat != multiArgTypeStructPtr {
				ev = ev.Elem()
			}
			dv.Set(reflect.Append(dv, ev))
		}
		keys = append(keys, k)
	}

	if more {
		var err error
		cursor, err = t.Cursor()
		if err != nil {
			return nil, err
		}
		var ei interface{}
		if !keysOnly {
			ei = reflect.New(elemType).Interface()
		}
		_, err = t.Next(ei)
		if err == datastore.Done {
			more = false
		}
	}

	return &Page{
		Keys:  keys,
		Start: cursor,
		More:  more,
	}, nil
}
