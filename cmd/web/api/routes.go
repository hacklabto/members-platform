package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// todo: service (ldap/db) auth

func RpcRouter() chi.Router {
	r := chi.NewRouter()

	r.Get("/verify_card", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	})

	return r
}

func RestRouter() chi.Router {
	r := chi.NewRouter()

	// db crud I guess

	return r
}
