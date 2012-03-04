package gforms

import (
	"net/url"
	"reflect"
)

type FormValuer interface {
	FormValue(key string) string
}

type Form interface {
	SetErrors(map[string]error)
	Errors() map[string]error
}

func InitForm(f Form) {
	s := reflect.ValueOf(f).Elem()
	typeOfForm := s.Type()
	for i := 0; i < s.NumField(); i++ {
		field, ok := s.Field(i).Interface().(Field)
		if !ok {
			continue
		}
		field.SetName(typeOfForm.Field(i).Name)
	}
}

func IsValid(f Form, data map[string]interface{}) bool {
	s := reflect.ValueOf(f).Elem()
	errs := make(map[string]error, 0)
	for i := 0; i < s.NumField(); i++ {
		field, ok := s.Field(i).Interface().(Field)
		if !ok {
			continue
		}

		if fieldErr := field.Validate(data[field.Name()]); fieldErr != nil {
			errs[field.Name()] = fieldErr
		}
	}
	f.SetErrors(errs)

	return len(f.Errors()) == 0
}

func IsFormValid(f Form, data url.Values) bool {
	m := make(map[string]interface{})
	for k, v := range data {
		if len(v) > 0 {
			m[k] = v[0]
		}
	}
	return IsValid(f, m)
}

type BaseForm struct {
	errors map[string]error
}

func (bf *BaseForm) SetErrors(errors map[string]error) {
	bf.errors = errors
}

func (bf *BaseForm) Errors() map[string]error {
	return bf.errors
}
