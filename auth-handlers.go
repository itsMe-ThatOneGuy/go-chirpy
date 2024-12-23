package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/itsMe-ThatOneGuy/go-chirpy/internal/auth"
	"github.com/itsMe-ThatOneGuy/go-chirpy/internal/database"
)

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type jsonReqParams struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type jsonResParams struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	token, err := auth.MakeJWT(user.ID, cfg.seceret, time.Hour)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error creating access JWT", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error creating refresh token", err)
		return
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
		UserID:    user.ID,
	})
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	jsonResponse(w, http.StatusOK, jsonResParams{
		User: User{
			ID:           user.ID,
			CreatedAt:    user.CreatedAt,
			UpdatedAt:    user.UpdatedAt,
			Email:        user.Email,
			IsChirpyRead: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {
	type jsonResParams struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Error with authorization header", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		responseError(w, http.StatusUnauthorized, "Could not get user for refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(
		user.ID,
		cfg.seceret,
		time.Hour,
	)
	if err != nil {
		responseError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	jsonResponse(w, http.StatusOK, jsonResParams{
		Token: accessToken,
	})

}

func (cfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		responseError(w, http.StatusBadRequest, "Error with authorization header", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Couldn't revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
