package main

import (
	"log"
	"net/http"
)

func main() {
	port := "8080"
	root := "."

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(root)))

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("serving on port %s", port)
	log.Fatal(server.ListenAndServe())

}
