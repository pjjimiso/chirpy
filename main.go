package main

import _ "github.com/lib/pq"

import ( 
	"fmt"
	"net/http"
	"log"
	"sync/atomic"
	"encoding/json"
	"io"
	"regexp"
	"os"
	"database/sql"

	"github.com/pjjimiso/chirpy/internal/database"

	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
}


func main() {
	const filepathRoot = "."
	const port = "8080"


	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Errorf("An error occured: %v", err)
	}
	apiCfg := apiConfig{
		dbQueries: database.New(db),
	}

	mux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))

	mux.HandleFunc("GET /api/healthz", handlerReadyCheck)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerWriteRequestCount)
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

func handlerReadyCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
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

func cleanMessage(msg string) string {
	cleanedMsg := msg
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	for _, badWord := range badWords {
		pattern := fmt.Sprintf("(?i)%s", badWord)
		re := regexp.MustCompile(pattern)
		cleanedMsg = re.ReplaceAllString(cleanedMsg, "****")
	}
	return cleanedMsg
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

func respondWithError(w http.ResponseWriter, code int, msg string) error { 
	return respondWithJSON(w, code, map[string]string{"error": msg})
}

func (cfg *apiConfig) handlerWriteRequestCount(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerResetRequestCount(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	cfg.fileserverHits.Store(0)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

