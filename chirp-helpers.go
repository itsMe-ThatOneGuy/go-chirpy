package main

import (
	"net/http"
	"strings"
)

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

func handleValidateChirp(w http.ResponseWriter, body string) string {
	if len(body) > maxChirpLen {
		responseError(w, http.StatusBadRequest, "Chirp too long", nil)
		return ""
	}

	clean := cleanBody(body, getBlackListWords())

	return clean
}
