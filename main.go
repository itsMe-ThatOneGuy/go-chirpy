package main

import (
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	server := http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	log.Printf("serving on port %s", port)
	log.Fatal(server.ListenAndServe())

}
