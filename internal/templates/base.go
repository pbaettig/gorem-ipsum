package templates

import (
	"html/template"
	"io"
)

// BaseData ...
type BaseData struct {
	Body string
}

// baseTemplate ...
type baseTemplate struct{ *template.Template }

// Render ...
func (t baseTemplate) Render(data BaseData, w io.Writer) {
	t.Execute(w, data)
}

var (
	// Base ...
	Base baseTemplate
)

func init() {
	Base = baseTemplate{mustParse("base", "<p>{{ .Body }}</p>")}
}
