package main

import (
	"log"
	"net/http"
)

func main() {
	const address = "127.0.0.1:8080"
	log.Printf("Lead portfolio listening at http://%s", address)
	if err := http.ListenAndServe(address, NewHandler()); err != nil {
		log.Fatal(err)
	}
}
