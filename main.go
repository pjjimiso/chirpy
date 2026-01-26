package main

import ( 
	"fmt"
	"net/http"
	"log"
	"sync/atomic"
	"encoding/json"
	"io"
	"os"
	"database/sql"

	"github.com/pjjimiso/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db *database.Queries
}


func main() {
	const filepathRoot = "."
	const port = "8080"


	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error opening database: %s", err)
	}
	dbQueries := database.New(db)
	apiCfg := apiConfig{
		fileserverHits:	atomic.Int32{},
		db:		dbQueries,
	}

	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /api/healthz", handlerReadyCheck)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerResetRequestCount)

	srv := http.Server {
		Addr:		":" + port,
		Handler:	mux,
	}

	if err := srv.ListenAndServe(); err != nil {
		// Error starting or closing listener
		fmt.Errorf("HTTP server ListenAndServe: %v", err)
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
}

func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		CleanedBody string `json:"cleaned_body"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil { 
		respondWithError(w, 500, "couldn't read request")
		return
	}

	req := requestBody{}
	err = json.Unmarshal(dat, &req)
	if err != nil { 
		respondWithError(w, 500, "couldn't unmarshal parameters")
		return
	}

	origMsg := req.Body
	if len(origMsg) > 140 { 
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	respondWithJSON(w, 200, responseBody{
		CleanedBody: cleanMessage(origMsg),
	})
}

