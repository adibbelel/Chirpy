package main

import (
	"fmt"
	"net/http"
	"log"
	"sync/atomic"
	"encoding/json"
	"regexp"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main () {
	fmt.Println("Starting Server")
	apiCfg := apiConfig {
		fileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", ContentTypeHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", jsonHandler)
	server := http.Server {
		Handler: mux,
		Addr: ":8080",
	}
	
	log.Fatal(server.ListenAndServe())
}

func ContentTypeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}


func (cfg *apiConfig) metricHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}

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
