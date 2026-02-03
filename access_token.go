package main

import (
	"net/http"
	"time"
	
	"github.com/pjjimiso/chirpy/internal/auth"
)


func (cfg *apiConfig) handlerRefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	type AccessToken struct {
		TokenString	string `json:"token"`
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't find token", err)
		return
	}

	token, err := cfg.db.GetUserFromRefreshToken(r.Context(), tokenString)
	if err != nil { 
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user refresh token", err)
		return
	}

	// Check if token has expired
	if time.Now().After(token.ExpiresAt) {
		respondWithError(w, http.StatusUnauthorized, "Refresh token expired", err)
		return
	}

	// create a new JWT
	jwt, err := auth.MakeJWT(token.UserID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, AccessToken{
		TokenString:	jwt,
	})
}

func (cfg *apiConfig) handlerRevokeAccessToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header) 
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't find token", err)
		return
	}

	err = cfg.db.UpdateTokenRevokedAt(r.Context(), token)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke session", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
