package main

import (
	"encoding/json"
	"log"
	"net/http"
)

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
