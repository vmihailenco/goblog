package gforms

import (
	"fmt"
	"html/template"
	"strings"
	tTemplate "text/template"
)

type Widget interface {
	Attrs() *WidgetAttrs
	Render([]string, string) template.HTML
}

type WidgetAttrs struct {
	Attrs [][2]string
}

func (w *WidgetAttrs) Set(name, value string) {
	value = tTemplate.HTMLEscapeString(value)

	exists := false
	for i := range w.Attrs {
		attr := &w.Attrs[i]
		if attr[0] == name {
			exists = true
			attr[1] = value
		}
	}
	if !exists {
		w.Attrs = append(w.Attrs, [...]string{name, value})
	}
}

func (w *WidgetAttrs) Get(name string) (string, bool) {
	for _, attr := range w.Attrs {
		if attr[0] == name {
			return attr[1], true
		}
	}
	return "", false
}

func (w *WidgetAttrs) Names() []string {
	names := make([]string, 0)
	for _, attr := range w.Attrs {
		names = append(names, attr[0])
	}
	return names
}

func (w *WidgetAttrs) String() string {
	attrsArr := make([]string, 0)
	for _, attr := range w.Attrs {
		attrsArr = append(attrsArr, fmt.Sprintf(`%v="%v"`, attr[0], attr[1]))
	}
	if len(attrsArr) > 0 {
		return " " + strings.Join(attrsArr, " ")
	}
	return ""
}

type BaseWidget struct {
	HTML  string
	attrs *WidgetAttrs
}

func (w *BaseWidget) Attrs() *WidgetAttrs {
	return w.attrs
}

func (w *BaseWidget) Render(attrs []string, value string) template.HTML {
	widgetAttrs := &WidgetAttrs{}
	baseAttrs := w.Attrs()
	for _, name := range baseAttrs.Names() {
		value, _ := baseAttrs.Get(name)
		widgetAttrs.Set(name, value)
	}
	for i := 0; i < len(attrs); i += 2 {
		widgetAttrs.Set(attrs[i], attrs[i+1])
	}
	value = tTemplate.HTMLEscapeString(value)
	html := fmt.Sprintf(w.HTML, widgetAttrs.String(), value)
	return template.HTML(html)
}

func NewTextWidget() Widget {
	return &BaseWidget{
		HTML: `<input%v value="%v" />`,
		attrs: &WidgetAttrs{
			Attrs: [][2]string{{"type", "text"}},
		},
	}
}

func NewTextareaWidget() Widget {
	return &BaseWidget{
		HTML: `<textarea%v>%v</textarea>`,
		attrs: &WidgetAttrs{
			Attrs: make([][2]string, 0),
		},
	}
}

func NewCheckboxWidget() Widget {
	return &BaseWidget{
		HTML: `<input%v value="%v" />`,
		attrs: &WidgetAttrs{
			Attrs: [][2]string{{"type", "checkbox"}},
		},
	}
}
