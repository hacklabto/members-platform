package ui

import (
	"log"
	"members-platform/internal/auth"
	"net/http"
)

// todo: integrate this with auth
// todo: check HX-Current-URL and add title to component-only request
func MaybeHtmxComponent(rw http.ResponseWriter, r *http.Request, page string, pageData any) {
	if r.Header.Get("HX-Request") == "true" {
		ctx := r.Context()
		data := TmplData{
			Ctx: TmplContext{
				AuthLevel:       ctx.Value(auth.Ctx__AuthLevel).(auth.AuthLevel),
				CurrentUsername: ctx.Value(auth.Ctx__AuthenticatedUser).(string),
			},
			Data: pageData,
		}
		if err := Page(rw, page, data); err != nil {
			log.Println(err)
		}
		return
	}

	if err := PageWithShell(r.Context(), rw, page, pageData); err != nil {
		log.Println(err)
	}
}
