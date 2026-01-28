package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"io"
	"regexp"
	"log"

	"github.com/google/uuid"
	"github.com/pjjimiso/chirpy/internal/database"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
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

	//fmt.Println("chirp 0:", chirps[0].CleanedBody)
	//fmt.Println("chirp 1:", chirps[1].CleanedBody)

	respondWithJSON(w, 200, chirps)
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body	string		`json:"body"`
		UserID	uuid.UUID	`json:"user_id"`
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
		UserID: params.UserID,
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

