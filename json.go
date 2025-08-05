package main

import (
	"net/http"
	"regexp"
	"encoding/json"
	"log"
)

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	type returnVals struct {
		Error string `json:"error"`
		Validity bool `json:"valid"`
		CleanBody string `json:"cleaned_body"`
	}
	respBody := returnVals{
		Validity: true,
	}

	cleaned := params.Body
	replacements := map[string]string{
		"kerfuffle": "****",
		"sharbert":  "****",
		"fornax":    "****",
	}

	for word, replacement := range replacements {
		pattern := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(word))
		cleaned = pattern.ReplaceAllString(cleaned, replacement)
	}
    respBody.CleanBody = cleaned

	if len(params.Body) > 140 {
		respBody.Error = "oopsie, too long"
		respBody.Validity = false
		w.WriteHeader(400)
		return
	}

	dat, err := json.Marshal(respBody)
	if err != nil {
		respBody.Error = "Error marshalling JSON"
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(dat)

}

