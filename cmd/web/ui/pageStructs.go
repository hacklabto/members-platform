package ui

import "html/template"

type Passwd struct {
	Error string
	Token string
}

type Confirmation struct {
	Title   string
	Message template.HTML
}
