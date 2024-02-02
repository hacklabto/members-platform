package main

import (
	"fmt"
	"log"
	"members-platform/internal/db"

	"github.com/emersion/go-smtp"
)

func main() {
	if err := db.ConnectRedis(); err != nil {
		log.Fatalln(fmt.Errorf("connect redis: %w", err))
	}

	s := smtp.NewServer(&recvBackend{})

	s.Addr = "100.80.182.53:2525"
	s.Domain = "lists.hacklab.to"

	log.Println("Starting SMTP server at", s.Addr)
	log.Fatal(s.ListenAndServe())
}
