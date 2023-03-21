package examples

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/mfeilen/go-i18n"
)

// renderTemplate using Golang templating + making i18n available in template
// call within template with eg. {{ i18n "module.function.title" }}
func renderTemplate(tplPath, tplName string, data interface{}) (string, error) {

	tplFile := tplPath + tplName
	funcMap := template.FuncMap{ // register func to template
		"i18n": i18n.Get,
	}

	t, err := template.New(tplFile).Funcs(funcMap).ParseFiles(tplFile)
	if err != nil {
		return ``, fmt.Errorf(`template %s could not be loaded, because %v`, tplFile, err)
	}

	var buf bytes.Buffer
	err = t.ExecuteTemplate(&buf, tplName, data)
	if err != nil {
		return ``, fmt.Errorf(`template %s could not be parsed, because %v`, tplName, err)
	}

	return string(buf.String()), nil
}
