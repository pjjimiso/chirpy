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
		respondWithError(w, http.StatusInternalServerError, "error retrieving refresh token")
		return
	}

	token, err := cfg.db.GetUserFromRefreshToken(r.Context(), tokenString)
	if err != nil { 
		log.Printf("error querying refresh token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "http.StatusUnauthorized Unauthorized")
		return
	}

	// Check if token has expired
	if time.Now().After(token.ExpiresAt) {
		log.Printf("refresh token has expired")
		respondWithError(w, http.StatusUnauthorized, "http.StatusUnauthorized Unauthorized")
		return
	}

	// create a new JWT
	jwt, err := auth.MakeJWT(token.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		log.Printf("error creating access token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "error creating access token")
		return
	}

	respondWithJSON(w, http.StatusOK, AccessToken{
		TokenString:	jwt,
	})
}

func (cfg *apiConfig) handlerRevokeAccessToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header) 
	if err != nil { 
		log.Printf("error retrieving token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "error retrieving refresh token")
		return
	}

	cfg.db.UpdateTokenRevokedAt(r.Context(), token)
	w.WriteHeader(http.StatusNoContent)
}
