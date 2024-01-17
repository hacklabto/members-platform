package main

import (
	"log"
	"members-platform/cmd/web/api"
	"members-platform/cmd/web/ui"
	"members-platform/internal/db"
	"net/http"
)

func main() {

	r := ui.Router()
	r.Mount("/rpc", api.RpcRouter())
	r.Mount("/rest", api.RestRouter())

	if err := db.ConnectPG(true); err != nil {
		log.Fatalln(err)
	}
	if err := db.ConnectRedis(); err != nil {
		log.Fatalln(err)
	}

	port := ":18884"
	log.Println("starting web server at", port)
	http.ListenAndServe(port, r)
}
