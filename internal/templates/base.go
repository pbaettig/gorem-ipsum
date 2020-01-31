package templates

import (
	"html/template"
	"io"
)

// BaseData ...
type BaseData struct {
	Body string
}

// BaseTemplate ...
type BaseTemplate struct{ *template.Template }

// Render ...
func (t BaseTemplate) Render(data BaseData, w io.Writer) {
	t.Execute(w, data)
}

var (
	// Base ...
	Base BaseTemplate
)

func init() {
	Base = BaseTemplate{mustParse("base", "<p>{{ .Body }}</p>")}
}
