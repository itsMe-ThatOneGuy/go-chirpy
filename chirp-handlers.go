package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/itsMe-ThatOneGuy/go-chirpy/internal/auth"
	"github.com/itsMe-ThatOneGuy/go-chirpy/internal/database"
)

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
		responseError(w, http.StatusInternalServerError, "Error decoding parameters", err)
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

	queryUserID := r.URL.Query().Get("author_id")
	authorID := uuid.Nil
	if queryUserID != "" {
		authorID, err = uuid.Parse(queryUserID)
		if err != nil {
			responseError(w, http.StatusBadRequest, "Invallid user ID", err)
			return
		}
	}

	sortDirection := "asc"
	sortDirectionParam := r.URL.Query().Get("sort")
	if sortDirectionParam == "desc" {
		sortDirection = "desc"
	}

	chirps := []Chirp{}
	for _, chirp := range dbChirps {
		if authorID != uuid.Nil && chirp.UserID != authorID {
			continue
		}

		chirps = append(chirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	sort.Slice(chirps, func(i, j int) bool {
		if sortDirection == "desc" {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})

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

func (cfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	strChirpID := r.PathValue("chirpID")
	uuidChirpID, err := uuid.Parse(strChirpID)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Invaild chirpID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		responseError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.seceret)
	if err != nil {
		responseError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), uuidChirpID)
	if err != nil {
		responseError(w, http.StatusNotFound, "Could not get chirp", err)
	}

	if chirp.UserID != userID {
		responseError(w, http.StatusForbidden, "Can't delete someone else's chirp", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error deleting chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
