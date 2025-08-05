package main

import (
	_ "github.com/lib/pq"
	"fmt"
	"net/http"
	"log"
	"sync/atomic"
	"encoding/json"
	"github.com/joho/godotenv"
	"github.com/google/uuid"
	"os"
	"time"
	"database/sql"
	"github.com/adibbelel/Chirpy/internal/database"
	
)

type apiConfig struct {
	fileserverHits atomic.Int32
	queries *database.Queries
	auth string
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func main () {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Could not open database URL: %v", err)
	}
	dbQueries := database.New(db)

	fmt.Println("Starting Server")
	apiCfg := apiConfig {
		fileserverHits: atomic.Int32{},
		queries: dbQueries,
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", ContentTypeHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)
	mux.HandleFunc("POST /api/validate_chirp", jsonHandler)
	mux.HandleFunc("POST /api/users", apiCfg.emailHandler)
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

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.auth != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Reset is only allowed in dev environment"))
		return
	}
	err := cfg.queries.Reset(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to reset the database: " + err.Error()))
		return
	}
	cfg.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) emailHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	user, err := cfg.queries.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Fatalf("Could not create user: %v", err)
	}

	dat, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(dat)

}
