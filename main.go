package main

import (
	"log"
	"net/http"
	"os"

	"github.com/tron-us/dingobot/github"
)

var (
	host string
)

func init() {
	host = os.Getenv("HOST")
	if host == "" {
		host = ":8888"
	}
}

func main() {
	http.HandleFunc("/github", github.Handle)
	log.Printf("Listening on %s\n", host)
	log.Fatal(http.ListenAndServe(host, nil))
}
