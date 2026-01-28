package main

import (
	"net/http"
	"encoding/json"
	"io"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID		uuid.UUID	`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	Email		string		`json:"email"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct { 
		Email string `json:"email"`
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

	user, err := cfg.db.CreateUser(r.Context(), params.Email)
	
	respondWithJSON(w, 201, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})	
}


