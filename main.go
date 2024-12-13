package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/itsMe-ThatOneGuy/go-chirpy/internal/auth"
	"github.com/itsMe-ThatOneGuy/go-chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const maxChirpLen = 140

func main() {
	port := "8080"
	root := "."

	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("env variable PLATFORM not set")
	}
	seceret := os.Getenv("SECERET")
	if seceret == "" {
		log.Fatal("env variable SECERET not set")
	}

	dbCon, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("error connecting to db: %s", err)
	}
	defer dbCon.Close()

	dbQueries := database.New(dbCon)

	apiCfg := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		seceret:        seceret,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(root)))))

	mux.HandleFunc("GET /api/healthz", handleReadCheck)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/chirps", apiCfg.handleChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handleGetChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("serving on port %s", port)
	log.Fatal(server.ListenAndServe())

}

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	platform       string
	seceret        string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type Chirps struct {
	ChirpList []Chirp
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

func handleValidateChirp(w http.ResponseWriter, body string) string {
	if len(body) > maxChirpLen {
		responseError(w, http.StatusBadRequest, "Chirp too long", nil)
		return ""
	}

	clean := cleanBody(body, getBlackListWords())

	return clean
}

func (cfg *apiConfig) handleChirp(w http.ResponseWriter, r *http.Request) {
	type jsonReqParams struct {
		Body string `json:"body"`
	}

	type jsonResParams struct {
		Chirp
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error getting bearerToken", err)
		return
	}

	jwtUserID, err := auth.ValidateJWT(bearerToken, cfg.seceret)
	if err != nil {
		responseError(w, http.StatusUnauthorized, "Error validating JWT", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := jsonReqParams{}
	err = decoder.Decode(&params)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error decoding parameter 1", err)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   handleValidateChirp(w, params.Body),
		UserID: jwtUserID,
	})
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error creating chirp", err)
		return
	}

	jsonResponse(w, http.StatusCreated, jsonResParams{
		Chirp: Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		},
	})
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error getting chirps", err)
		return
	}

	chirps := make([]Chirp, len(dbChirps))
	for i, chirp := range dbChirps {
		chirps[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	jsonResponse(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	strChirpID := r.PathValue("chirpID")
	uuidChirpID, err := uuid.Parse(strChirpID)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Invaild chirpID", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), uuidChirpID)
	if err != nil {
		responseError(w, http.StatusNotFound, "Could not get chirp", err)
	}

	jsonResponse(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type jsonReqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type jsonResParams struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := jsonReqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error decoding parameter", err)
		return
	}
	if params.Email == "" {
		responseError(w, http.StatusBadRequest, "Empty email", nil)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error hashing password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}

	jsonResponse(w, http.StatusCreated, jsonResParams{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})

}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type jsonReqParams struct {
		Email        string `json:"email"`
		Password     string `json:"password"`
		ExpiresInSec *int   `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := jsonReqParams{}
	err := decoder.Decode(&params)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error decoding parameter", err)
		return
	}
	if params.Email == "" {
		responseError(w, http.StatusBadRequest, "Empty email", nil)
		return
	}

	tokenExpiresIn := 3600
	if params.ExpiresInSec != nil {
		if *params.ExpiresInSec <= tokenExpiresIn {
			tokenExpiresIn = *params.ExpiresInSec
		}
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		responseError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		responseError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.seceret, time.Duration(tokenExpiresIn)*time.Second)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error decoding parameter", err)
		return
	}

	jsonResponse(w, http.StatusOK, User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})
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

func cleanBody(reqBody string, blWords map[string]struct{}) string {
	reqWords := strings.Split(reqBody, " ")

	for i, word := range reqWords {
		lowerCaseWord := strings.ToLower(word)
		if _, ok := blWords[lowerCaseWord]; ok {
			reqWords[i] = "****"
		}
	}
	cleaned := strings.Join(reqWords, " ")

	return cleaned
}

func getBlackListWords() map[string]struct{} {
	return map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
}
