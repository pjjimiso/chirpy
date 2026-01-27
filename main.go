package main

import ( 
	"fmt"
	"net/http"
	"log"
	"sync/atomic"
	"os"
	"time"
	"database/sql"

	"github.com/pjjimiso/chirpy/internal/database"
	"github.com/joho/godotenv"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits	atomic.Int32
	db		*database.Queries
	platform	string
}

type User struct {
	ID		uuid.UUID	`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	Email		string		`json:"email"`
}

type Chirp struct {
	ID		uuid.UUID	`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	CleanedBody	string		`json:"body"`
	UserID		uuid.UUID	`json:"user_id"`
}

func main() {
	const filepathRoot = "."
	const port = "8080"


	godotenv.Load()
	plat := os.Getenv("PLATFORM")
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
		platform:	plat,
	}

	mux := http.NewServeMux()
	fsHandler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fsHandler))

	mux.HandleFunc("GET /api/healthz", handlerReadyCheck)
	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirp)
	mux.HandleFunc("POST /api/users", apiCfg.handlerGetUser)

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerAdminReset)


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

