package main

import (
	"net/http"
	"time"
	"log"
	
	"github.com/pjjimiso/chirpy/internal/auth"
)


func (cfg *apiConfig) handlerRefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	type AccessToken struct {
		TokenString	string `json:"token"`
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil { 
		log.Printf("error retrieving token: %s", err)
		respondWithError(w, 500, "error retrieving refresh token")
		return
	}

	token, err := cfg.db.GetUserFromRefreshToken(r.Context(), tokenString)
	if err != nil { 
		log.Printf("error querying refresh token: %s", err)
		respondWithError(w, 401, "401 Unauthorized")
		return
	}

	// Check if token has expired
	if time.Now().After(token.ExpiresAt) {
		log.Printf("refresh token has expired")
		respondWithError(w, 401, "401 Unauthorized")
		return
	}

	// create a new JWT
	jwt, err := auth.MakeJWT(token.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		log.Printf("error creating access token: %s", err)
		respondWithError(w, 500, "error creating access token")
		return
	}

	respondWithJSON(w, 200, AccessToken{
		TokenString:	jwt,
	})
}

func (cfg *apiConfig) handlerRevokeAccessToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header) 
	if err != nil { 
		log.Printf("error retrieving token: %s", err)
		respondWithError(w, 500, "error retrieving refresh token")
		return
	}

	cfg.db.UpdateTokenRevokedAt(r.Context(), token)
	w.WriteHeader(204)
}
