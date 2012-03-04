package gforms

import (
	"html/template"
	"testing"
)

type FooForm struct {
	*BaseForm
	Name *StringField
	Age  *Int64Field
}

func TestForm(t *testing.T) {
	fooForm := &FooForm{
		BaseForm: &BaseForm{},
		Name:     NewStringField(true, 0, 0),
		Age:      NewInt64Field(true),
	}
	InitForm(fooForm)

	data := map[string]interface{}{
		"Name": "bar",
		"Age":  "23",
	}
	if !IsValid(fooForm, data) {
		t.Errorf("fooForm did not pass validation: %v", fooForm.Errors())
	}
	if v := fooForm.Name.Value(); v != "bar" {
		t.Errorf("fooForm.Name.Value(): expected bar, got %v", v)
	}
	expected := template.HTML(`<input type="text" id="Name" name="Name" value="bar" />`)
	if v := fooForm.Name.Render(); v != expected {
		t.Errorf("foorForm.Name.Render(): expected %v, got %v", expected, v)
	}
	if v := fooForm.Age.Value(); v != 23 {
		t.Errorf("fooForm.Age.Value(): expected 23, got %v", v)
	}
}

func TestFormRequiredField(t *testing.T) {
	fooForm := &FooForm{
		BaseForm: &BaseForm{},
		Name:     NewStringField(true, 0, 0),
		Age:      NewInt64Field(true),
	}
	InitForm(fooForm)

	data := map[string]interface{}{}
	if IsValid(fooForm, data) {
		t.Errorf("fooForm passed validation")
	}

	expected := "This field is required."
	if v := fooForm.Name.Error().Error(); v != expected {
		t.Errorf("fooForm.Name.Error(): expected %v, got %v", expected, v)
	}
}
