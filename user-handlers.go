package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/itsMe-ThatOneGuy/go-chirpy/internal/auth"
	"github.com/itsMe-ThatOneGuy/go-chirpy/internal/database"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	IsChirpyRead bool      `json:"is_chirpy_red"`
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
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			IsChirpyRead: user.IsChirpyRed,
		},
	})

}

func (cfg *apiConfig) handleUserUpdate(w http.ResponseWriter, r *http.Request) {
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

	hashedPassword, err := auth.HashPassword(params.Password)

	updatedUser, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Couldn't update user info", err)
		return
	}

	jsonResponse(w, http.StatusOK, jsonResParams{
		User: User{
			ID:           updatedUser.ID,
			UpdatedAt:    updatedUser.UpdatedAt,
			CreatedAt:    updatedUser.CreatedAt,
			Email:        updatedUser.Email,
			IsChirpyRead: updatedUser.IsChirpyRed,
		},
	})

}
