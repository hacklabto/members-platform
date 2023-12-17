package ui

import (
	"encoding/base64"
	"log"
	"members-platform/internal/auth"
	"members-platform/static"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Router() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(auth.AuthenticateHTTP)

	registerStaticRoutes(r)
	registerStaticPages(r)

	// todo: this needs to be POST with CSRF
	r.Get("/logout/", func(rw http.ResponseWriter, r *http.Request) {
		http.SetCookie(rw, &http.Cookie{
			Name:    "HL-Session",
			Value:   "",
			Path:    "/",
			Expires: time.Now().UTC(),
		})
		http.Redirect(rw, r, "/", http.StatusTemporaryRedirect)
	})

	r.Post("/login/", func(rw http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			log.Println(err)
		}
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		ok, err := auth.AuthenticateUser(username, password)
		if err != nil {
			MaybeHtmxComponent(rw, r, "login", shit{Error: err.Error()})
			return
		}
		if !ok {
			MaybeHtmxComponent(rw, r, "login", shit{Error: "Invalid username or password"})
			return
		}

		http.SetCookie(rw, &http.Cookie{
			Name:     "HL-Session",
			Value:    "Basic " + base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{username, password}, ":"))),
			HttpOnly: true,
			Path:     "/",
		})

		// htmx this is stupid
		if r.Header.Get("HX-Request") == "true" {
			rw.Header().Set("HX-Location", "/")
		} else {
			http.Redirect(rw, r, "/", http.StatusFound)
		}
	})

	r.Post("/passwd/", func(rw http.ResponseWriter, r *http.Request) {
		token := auth.CreateResetToken("lillian")
		_ = auth.SendResetEmail("lillian@hacklab.to", "lillian", token)

		// todo: don't, obviously
		err := auth.DoChangePassword(
			"uid=lilliantest,ou=people,dc=hacklab,dc=to",
			"NotAPassword!!",
			"uid=lilliantest,ou=people,dc=hacklab,dc=to",
			"newpass1234",
		)
		data := shit{Error: "ok"}
		if err != nil {
			data.Error = err.Error()
		}
		MaybeHtmxComponent(rw, r, "passwd-reset", data)
	})

	return r
}

func registerStaticRoutes(r chi.Router) {
	r.Get("/favicon.ico", func(rw http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/static/favicon.ico"
		static.Server().ServeHTTP(rw, r)
	})

	r.Get("/robots.txt", func(rw http.ResponseWriter, _ *http.Request) {
		rw.Write([]byte("User-Agent: *\nDisallow: *\nDisallow: /ban-me/admin.php"))
	})

	r.Get("/static/*", static.Server())
}

func registerStaticPages(r chi.Router) {
	pathPages := map[string]string{
		"/":        "index",
		"/login/":  "login",
		"/passwd/": "passwd",
		"/apply/":  "apply",
	}

	for k, v := range pathPages {
		// ???
		p := k
		q := v
		r.Get(p, func(rw http.ResponseWriter, r *http.Request) {
			if err := PageWithShell(r.Context(), rw, q, nil); err != nil {
				log.Println(err)
			}
		})
	}
}

type shit struct {
	Error string
}
