package main 

import (
	"net/http"
	"encoding/json"
	"log"
	"regexp"
	"github.com/google/uuid"
	"time"
	"github.com/adibbelel/Chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body     string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type Chirps struct {
	Chirps []database.Chirp
}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
		UserID uuid.UUID`json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	chirp, err := cfg.db.CreateChirp(r.Context(),database.CreateChirpParams{
		Body: params.Body,
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not create Chirp", err)
	}
	type response struct {
		Error string `json:"error"`
		Validity bool `json:"valid"`
		CleanBody string `json:"cleaned_body"`
		Chirp
	}
	respBody := response{
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

	respondWithJSON(w, http.StatusCreated, response{
		Chirp: Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		},
	})

}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not get chirps", err)
	}

	responseChirps := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		responseChirps[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
            UpdatedAt: chirp.UpdatedAt,
            Body:      chirp.Body,
            UserID:    chirp.UserID,
        }
    }


	respondWithJSON(w, http.StatusOK, responseChirps)
}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	chirp, err := cfg.db.GetChirp(r.Context(), uuid.MustParse(chirpID))
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not get chirp", err)
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
