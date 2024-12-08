package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func main() {
	port := "8080"
	root := "."

	apiCfg := apiConfig{}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(root)))))

	mux.HandleFunc("GET /api/healthz", handleReadCheck)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("serving on port %s", port)
	log.Fatal(server.ListenAndServe())

}

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func handleReadCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileServerHits.Load()

	htmlTemplate :=
		`<html>
            <body>
                <h1>Welcome, Chirpy Admin</h1>
                <p>Chirpy has been visited %d times!</p>
            </body>
        </html>`

	html := fmt.Sprintf(htmlTemplate, hits)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write([]byte(html))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits.Store(0)
	hits := cfg.fileServerHits.Load()
	msg := fmt.Sprintf("Hits: %d", hits)

	w.Write([]byte(msg))
}
func responseError(w http.ResponseWriter, status int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}

	if status > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}

	type jsonResError struct {
		Error string `json:"error"`
	}

	jsonResponse(w, status, jsonResError{
		Error: msg,
	})
}

func jsonResponse(w http.ResponseWriter, status int, payload interface{}) {
	res, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(res)

}
