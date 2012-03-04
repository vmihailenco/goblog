package gforms

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
)

var FieldTemplate *template.Template
var emptyHTML = template.HTML("")

func init() {
	FieldTemplate = template.New("field.html")
	FieldTemplate = FieldTemplate.Funcs(template.FuncMap{
		"renderLabel": RenderLabel,
		"renderError": RenderError,
	})
	var err error
	FieldTemplate, err = FieldTemplate.ParseFiles("templates/gforms/field.html")
	if err != nil {
		panic(err)
	}
}

func field(fIntrfc interface{}) (Field, error) {
	f, ok := fIntrfc.(Field)
	if !ok {
		return nil, errors.New("Expected Field")
	}
	return f, nil
}

func Render(f Field) (template.HTML, error) {
	buf := &bytes.Buffer{}
	if err := FieldTemplate.Execute(buf, map[string]interface{}{"f": f}); err != nil {
		return emptyHTML, err
	}
	return template.HTML(buf.String()), nil
}

func RenderError(fIntrfc interface{}) (template.HTML, error) {
	f, err := field(fIntrfc)
	if err != nil {
		return emptyHTML, err
	}

	if !f.HasError() {
		return emptyHTML, nil
	}

	return template.HTML(fmt.Sprintf(`<span class="help-inline">%v</span>`, f.Error())), nil
}

func RenderLabel(fIntrfc interface{}) (template.HTML, error) {
	f, err := field(fIntrfc)
	if err != nil {
		return emptyHTML, err
	}
	return template.HTML(fmt.Sprintf(`<label for="%v">%v</label>`, f.Name(), f.Name())), nil
}
