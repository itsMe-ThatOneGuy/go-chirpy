package main

import (
	"fmt"
	"net/http"
)

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
	if cfg.platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset only allowed in dev environment"))
		return
	}

	err := cfg.db.DeleteAllUsers(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error deleting users", err)
		return
	}

	cfg.fileServerHits.Store(0)
	cfg.fileServerHits.Load()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hits set to 0 and user table reset"))
}
