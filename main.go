package main

import (
	"log"
	"net/http"
)

func main() {
	port := "8080"
	root := "."

	mux := http.NewServeMux()
	server := http.Server{
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("serving on port %s", port)
	log.Fatal(server.ListenAndServe())

}
