package gforms

import (
	"bytes"
	"html/template"
	"testing"
)

func TestFields(t *testing.T) {
	stringField := NewStringField(true, 0, 0)
	if err := stringField.Validate("bar"); err != nil {
		t.Errorf("StringField did not pass validation: %v", err)
	}
	if v := stringField.Value(); v != "bar" {
		t.Errorf("StringField.Value(): expected bar, got %v", v)
	}
	expected := template.HTML(`<input type="text" value="bar" />`)
	if v := stringField.Render(); v != expected {
		t.Errorf("stringField.Render(): expected %v, got %v", expected, v)
	}

	int64Field := NewInt64Field(true)
	if err := int64Field.Validate("bar"); err.Error() != "bar is not valid integer." {
		t.Errorf("Int64Field.Validate(): expected , got %v", err)
	}
	if int64Field.HasError() != true {
		t.Errorf("int64Field.HasError(): field does not have error")
	}
}

func TestFieldRender(t *testing.T) {
	nameField := NewStringField(true, 0, 0)
	nameField.SetName("Name")

	tmplt, err := template.New("test").Parse("{{.name.Render}}")
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	err = tmplt.Execute(buf, map[string]interface{}{
		"name": nameField,
	})
	if err != nil {
		panic(err)
	}
	expected := `<input type="text" id="Name" name="Name" value="" />`
	if v := buf.String(); v != expected {
		t.Errorf("name.Render(): expected %v, got %v", expected, v)
	}
}

func TestFieldRenderWithAttrs(t *testing.T) {
	nameField := NewStringField(true, 0, 0)
	nameField.SetName("Name")

	tmplt, err := template.New("test").Parse(`{{.name.Render "title" "hi"}}`)
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	err = tmplt.Execute(buf, map[string]interface{}{
		"name": nameField,
	})
	if err != nil {
		panic(err)
	}
	expected := `<input type="text" id="Name" name="Name" title="hi" value="" />`
	if v := buf.String(); v != expected {
		t.Errorf("name.Render(): expected %v, got %v", expected, v)
	}
}
