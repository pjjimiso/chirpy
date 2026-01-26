package main

import ( 
	"fmt"
	"net/http"
	"log"
	"sync/atomic"
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
	fsHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fsHandler))

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

