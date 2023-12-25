package ui

import (
	"log"
	"net/http"
)

// todo: integrate this with auth
// todo: check HX-Current-URL and add title to component-only request
func MaybeHtmxComponent(rw http.ResponseWriter, r *http.Request, page string, data any) {
	if r.Header.Get("HX-Request") == "true" {
		if err := Page(rw, page, data); err != nil {
			log.Println(err)
		}
		return
	}

	if err := PageWithShell(r.Context(), rw, page, data); err != nil {
		log.Println(err)
	}
}
