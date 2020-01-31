package templates

import (
	"html/template"
	"log"
)

func mustParse(name, body string) *template.Template {
	tmpl, err := template.New(name).Parse(body)
	if err != nil {
		log.Panic(err)
	}

	return tmpl
}
