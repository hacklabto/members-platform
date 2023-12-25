package main

import (
	"log"
	"members-platform/cmd/web/api"
	"members-platform/cmd/web/ui"
	"members-platform/internal/db"
	"net/http"
	"os"
)

func main() {
	isPasswdWeb := os.Getenv("PASSWDWEB") == "true"

	r := ui.Router(isPasswdWeb)
	r.Mount("/rpc", api.RpcRouter())
	r.Mount("/rest", api.RestRouter())

	if err := db.ConnectPG(true); err != nil {
		log.Fatalln(err)
	}

	if isPasswdWeb {
		if err := db.ConnectRedis(); err != nil {
			log.Fatalln(err)
		}
	}

	log.Println("starting web server")
	port := ":18884"
	if isPasswdWeb {
		port = ":18885"
	}
	http.ListenAndServe(port, r)
}
