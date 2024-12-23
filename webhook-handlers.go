package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/itsMe-ThatOneGuy/go-chirpy/internal/auth"
)

func (cfg *apiConfig) handleUserUpgrade(w http.ResponseWriter, r *http.Request) {
	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		responseError(w, http.StatusUnauthorized, "Authorization header error", err)
		return
	}
	if key != cfg.polkaKey {
		responseError(w, http.StatusUnauthorized, "ApiKey invaild", err)
		return
	}

	type jsonReqParams struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		}
	}

	decoder := json.NewDecoder(r.Body)
	params := jsonReqParams{}
	err = decoder.Decode(&params)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error decoding parameter", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	reqUserID, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		responseError(w, http.StatusInternalServerError, "Error parsing string to uuid", err)
		return
	}

	err = cfg.db.UpgradeUser(r.Context(), reqUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			responseError(w, http.StatusNotFound, "Couldn't find user", err)
			return
		}
		responseError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
