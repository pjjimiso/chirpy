package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"io"
	"regexp"
	"time"
	"slices"

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

func (cfg *apiConfig) handlerChirpsGetAll(w http.ResponseWriter, r *http.Request) {
	var chirpsJSON []database.Chirp
	var err error


	author := r.URL.Query().Get("author_id")
	if author != "" { 
		userID, err := uuid.Parse(author)
		if err != nil { 
			respondWithError(w, http.StatusNotFound, "User not found", err)
			return
		}

		chirpsJSON, err = cfg.db.GetChirpsByAuthor(r.Context(), userID)
		if err != nil {
			respondWithError(w, http.StatusNotFound, "No chirps by that author were found", err)
			return
		}
	} else {
		chirpsJSON, err = cfg.db.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusNotFound, "No chirps were found", err)
			return
		}
	}

	chirps := []Chirp{}
	for _, chirp := range chirpsJSON {
		chirps = append(chirps, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			CleanedBody: chirp.Body,
			UserID: chirp.UserID,
		})
	}

	sortOrder := r.URL.Query().Get("sort")
	if sortOrder != "" {
		sortByCreatedAt(chirps, sortOrder)
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func sortByCreatedAt(items []Chirp, sortOrder string) {
	slices.SortFunc(items, func(a, b Chirp) int {
		if sortOrder == "desc" {
			return b.CreatedAt.Compare(a.CreatedAt)
		}
		return a.CreatedAt.Compare(b.CreatedAt)
	})
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	type parameters struct { 
		ID	uuid.UUID	`json:"id"`
	}
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirp", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil { 
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
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
		respondWithError(w, http.StatusInternalServerError, "Couldn't find token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil { 
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't read request parameters", err)
		return
	}

	params := parameters{}
	err = json.Unmarshal(dat, &params)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't unmarshal parameters", err)
		return
	}

	origMsg := params.Body
	if len(origMsg) > 140 { 
		respondWithError(w, http.StatusBadRequest, "Chirp exceeds 140 character limit", nil)
		return
	}

	cleanedMsg := cleanMessage(origMsg)
	
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: cleanedMsg,
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID: chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		CleanedBody: chirp.Body,
		UserID: chirp.UserID,
	})
}

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) { 
	token, err := auth.GetBearerToken(r.Header)
	if err != nil { 
		respondWithError(w, http.StatusUnauthorized, "Couldn't find token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil { 
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Invalid chirp ID", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil { 
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	if chirp.UserID != userID { 
		respondWithError(w, http.StatusForbidden, "You can't delete this chirp", nil)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:	chirpID,
		UserID:	userID,
	})
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
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

