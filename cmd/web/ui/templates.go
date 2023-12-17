package ui

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"members-platform/internal/auth"
	"strconv"
	"time"
)

//go:embed */*.html
var tmplFS embed.FS

//go:embed shell.html
var shellTmpl string

var pages *template.Template
var shell *template.Template

var pageTitles = map[string]string{
	"login": "Log In",
}

var funcs template.FuncMap = map[string]any{
	// auth
	"IsLoggedOut": func(a auth.AuthLevel) bool {
		return a == auth.AuthLevel_LoggedOut
	},
	"IsApplicantLoggedIn": func(a auth.AuthLevel) bool {
		return a == auth.AuthLevel_Applicant
	},
	"IsMemberLoggedIn": func(a auth.AuthLevel) bool {
		return a >= auth.AuthLevel_Member
	},
}

func init() {
	pages = template.Must(template.New("").Funcs(funcs).ParseFS(tmplFS, "pages/*.html"))
	shell = template.Must(template.New("shell.html").Funcs(funcs).Parse(shellTmpl))
}

type TmplContext struct {
	// TODO: csrf token
	AuthLevel       auth.AuthLevel
	CurrentUsername string
}

type TmplData struct {
	Ctx  TmplContext
	Data any
}

type ShellData struct {
	Title                   string
	CurrentYearForCopyright string

	UnsafeInnerHTML template.HTML
}

func Page(w io.Writer, tmpl string, pageData any) error {
	return pages.ExecuteTemplate(w, tmpl+".html", pageData)
}

func PageWithShell(ctx context.Context, w io.Writer, page string, pageData any) error {
	pw := bytes.NewBuffer([]byte{})
	data := TmplData{
		Ctx: TmplContext{
			AuthLevel:       ctx.Value(auth.Ctx__AuthLevel).(auth.AuthLevel),
			CurrentUsername: ctx.Value(auth.Ctx__AuthenticatedUser).(string),
		},
		Data: pageData,
	}
	if err := Page(pw, page, data); err != nil {
		return fmt.Errorf("executing page template: %w", err)
	}

	title := ""
	if v, ok := pageTitles[page]; ok {
		title += v
		title += " - "
	}
	title += "Hacklab Members Portal"

	shellData := ShellData{
		UnsafeInnerHTML:         template.HTML(pw.Bytes()),
		CurrentYearForCopyright: strconv.Itoa(time.Now().UTC().Year()),
		Title:                   title,
	}
	data.Data = shellData
	if err := shell.Execute(w, data); err != nil {
		return fmt.Errorf("executing shell template: %w", err)
	}
	return nil
}
