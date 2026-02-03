package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"io"
	"regexp"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/pjjimiso/chirpy/internal/database"
	"github.com/pjjimiso/chirpy/internal/auth"
)

type Chirp struct {
	ID		uuid.UUID	`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	CleanedBody	string		`json:"body"`
	UserID		uuid.UUID	`json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	chirps := []Chirp{}
	chirpsJSON, err := cfg.db.GetChirps(r.Context())

	if err != nil {
		log.Printf("error getting chirps: %s", err)
		respondWithError(w, 500, "error getting chirps")
		return
	}

	for _, chirp := range chirpsJSON {
		chirps = append(chirps, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			CleanedBody: chirp.Body,
			UserID: chirp.UserID,
		})
	}

	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, r *http.Request) {
	type parameters struct { 
		ID	uuid.UUID	`json:"id"`
	}
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil { 
		log.Printf("error parsing chirp uuid: %s", err)
		respondWithError(w, 500, "error parsing chirp uuid")
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil { 
		log.Printf("error getting chirp: %s", err)
		respondWithError(w, 404, "error getting chirp")
		return
	}

	respondWithJSON(w, 200, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		CleanedBody: chirp.Body,
		UserID: chirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body	string		`json:"body"`
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil { 
		log.Printf("error retrieving token: %s", err)
		respondWithError(w, 500, "couldn't get JWT")
		return
	}

	/*
	// Debug
	fmt.Println("Authorization header:", r.Header.Get("Authorization"))
	fmt.Printf("Authorization bytes: %v\n", []byte(r.Header.Get("Authorization")))
	fmt.Println("Extracted token:", token)
	fmt.Printf("Extracted token bytes: %v\n", []byte(token))
	*/

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil { 
		log.Printf("error validating jwt: %s", err)
		respondWithError(w, 401, "401 Unauthorized")
		return
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil { 
		respondWithError(w, 500, "couldn't read request")
		return
	}

	params := parameters{}
	err = json.Unmarshal(dat, &params)
	if err != nil { 
		respondWithError(w, 500, "couldn't unmarshal parameters")
		return
	}

	origMsg := params.Body
	if len(origMsg) > 140 { 
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	cleanedMsg := cleanMessage(origMsg)
	
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleanedMsg,
		UserID: userID,
	})
	if err != nil {
		log.Printf("error creating chirp: %s", err)
		respondWithError(w, 500, "couldn't create chirp")
		return
	}

	respondWithJSON(w, 201, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		CleanedBody: chirp.Body,
		UserID: chirp.UserID,
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

