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
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("serving on port %s", port)
	log.Fatal(server.ListenAndServe())

}
