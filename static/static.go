package static

import (
	"embed"
	"net/http"
)

//go:generate npx tailwindcss -o tailwind.css --content "../cmd/web/ui/**/*"

//go:embed *
var files embed.FS

func Server() http.HandlerFunc {
	return http.StripPrefix("/static/", http.FileServer(http.FS(files))).ServeHTTP
}
