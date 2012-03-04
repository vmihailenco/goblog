package gforms

import (
	"errors"
	"fmt"
	"html/template"
	"strconv"
)

type Field interface {
	SetName(string)
	Name() string
	IsRequired() bool
	Validate(data interface{}) error
	Render(attrs ...string) template.HTML
	Error() error
	HasError() bool
}

type BaseField struct {
	name        string
	Widget      Widget
	Required    bool
	data        interface{}
	error       error
	isValidated bool
}

func (bf *BaseField) SetName(name string) {
	bf.name = name
	bf.Widget.Attrs().Set("id", name)
	bf.Widget.Attrs().Set("name", name)
}

func (bf *BaseField) Name() string {
	return bf.name
}

func (bf *BaseField) IsRequired() bool {
	return bf.Required
}

func (bf *BaseField) Validate(data interface{}) error {
	bf.isValidated = true
	bf.data = data
	if bf.Required && bf.data == nil {
		bf.error = errors.New("This field is required.")
		return bf.error
	}
	return nil
}

func (bf *BaseField) Error() error {
	if !bf.isValidated {
		panic("Trying to get error on non-validated field.")
	}
	return bf.error
}

func (bf *BaseField) HasError() bool {
	return bf.error != nil
}

type StringField struct {
	*BaseField
	minLength, maxLength int
	value                string
}

func (f *StringField) Value() string {
	if !f.isValidated {
		panic("Trying to get error on non-validated field.")
	}
	return f.value
}

func (f *StringField) Validate(data interface{}) error {
	if err := f.BaseField.Validate(data); err != nil {
		return err
	}

	value := fmt.Sprint(data)
	valueLen := len(value)
	if f.minLength > 0 && valueLen < f.minLength {
		f.error = fmt.Errorf("This field should have at least %d symbols.", f.minLength)
		return f.error
	}
	if f.maxLength > 0 && valueLen > f.maxLength {
		f.error = fmt.Errorf("This field should have less than %d symbols.", f.maxLength)
		return f.error
	}

	f.value = value

	return nil
}

func (f *StringField) SetInitial(initial string) *StringField {
	f.value = initial
	return f
}

func (f *StringField) Render(attrs ...string) template.HTML {
	var s string
	if f.HasError() {
		s = fmt.Sprint(f.data)
	} else {
		s = f.value
	}
	return f.Widget.Render(attrs, s)
}

func NewStringField(required bool, minLength, maxLength int) *StringField {
	return &StringField{
		minLength: minLength,
		maxLength: maxLength,
		BaseField: &BaseField{
			Widget:   NewTextWidget(),
			Required: required,
		},
	}
}

func NewTextareaField(required bool, minLength, maxLength int) *StringField {
	return &StringField{
		minLength: minLength,
		maxLength: maxLength,
		BaseField: &BaseField{
			Widget:   NewTextareaWidget(),
			Required: required,
		},
	}
}

type Int64Field struct {
	*BaseField
	value int64
}

func (f *Int64Field) Value() int64 {
	if !f.isValidated {
		panic("trying to get value on non-validated field")
	}
	return f.value
}

func (f *Int64Field) Validate(data interface{}) error {
	if err := f.BaseField.Validate(data); err != nil {
		return err
	}

	value, err := strconv.ParseInt(fmt.Sprint(data), 10, 64)
	if err != nil {
		f.error = fmt.Errorf("%v is not valid integer.", data)
		return f.error
	}
	f.value = value

	return nil
}

func (f *Int64Field) SetInitial(initial int64) *Int64Field {
	f.value = initial
	return f
}

func (f *Int64Field) Render(attrs ...string) template.HTML {
	var s string
	if f.HasError() {
		s = fmt.Sprint(f.data)
	} else {
		s = string(f.value)
	}
	return f.Widget.Render(attrs, s)
}

func NewInt64Field(required bool) *Int64Field {
	return &Int64Field{
		BaseField: &BaseField{
			Widget:   NewTextWidget(),
			Required: required,
		},
	}
}
