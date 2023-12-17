package main

import (
	"log"
	"members-platform/cmd/web/api"
	"members-platform/cmd/web/ui"
	"members-platform/internal/db"
	"net/http"
)

func main() {
	if err := db.Connect(true); err != nil {
		log.Fatalln(err)
	}
	r := ui.Router()
	r.Mount("/rpc", api.RpcRouter())
	r.Mount("/rest", api.RestRouter())
	log.Println("starting web server")
	http.ListenAndServe(":18884", r)
}
