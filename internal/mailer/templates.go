package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed templates/*
var tmplFS embed.FS

var emails *template.Template

const footer = "\n--------------\n\nThis was an automated email sent by members.hacklab.to"

func init() {
	emails = template.Must(template.ParseFS(tmplFS, "templates/*.txt"))
}

func ExecuteTemplate(tmpl string, data any) (string, error) {
	w := bytes.NewBuffer([]byte{})

	if err := emails.ExecuteTemplate(w, tmpl+".txt", data); err != nil {
		return "", fmt.Errorf("execute email template: %w", err)
	}
	if _, err := w.Write([]byte(footer)); err != nil {
		return "", fmt.Errorf("write email footer: %w", err)
	}

	return w.String(), nil
}

func ExecuteTemplateWithoutFooter(tmpl string, data any) (string, error) {
	w := bytes.NewBuffer([]byte{})

	if err := emails.ExecuteTemplate(w, tmpl+".txt", data); err != nil {
		return "", fmt.Errorf("execute email template: %w", err)
	}

	return w.String(), nil
}
