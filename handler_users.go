package main

import (
	"net/http"
	"encoding/json"
	"io"
	"time"
	"log"

	"github.com/google/uuid"
	"github.com/pjjimiso/chirpy/internal/database"
	"github.com/pjjimiso/chirpy/internal/auth"
)

type User struct {
	ID		uuid.UUID	`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
	Email		string		`json:"email"`
	Token		string		`json:"token"`
	RefreshToken	string		`json:"refresh_token"`
}

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct { 
		Password	string		`json:"password"`
		Email		string		`json:"email"`
		ExpiresIn	*float64	`json:"expires_in_seconds,string,omitempty"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("failed to read request body: %s", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't read request")
		return
	}

	params := parameters{}
	err = json.Unmarshal(dat, &params)
	if err != nil { 
		log.Printf("json unmarshal failed: %s", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't unmarshal parameters")
		return
	}

	maxDurationSeconds := time.Hour.Seconds()
	expiresIn := time.Hour

	if (params.ExpiresIn == nil || (0 > *params.ExpiresIn || *params.ExpiresIn > maxDurationSeconds)) {
		expiresIn = time.Hour
	} else {
		expiresIn = time.Duration(*params.ExpiresIn) * time.Second
	}

	user, err := cfg.db.GetUser(r.Context(), params.Email)
	if err != nil {
		log.Printf("error getting user: %s", err)
		respondWithError(w, http.StatusInternalServerError, "error getting user")
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPasswords)
	if !match || err != nil { 
		log.Printf("error getting user %s", err)
		respondWithError(w, 401, "401 Unauthorized")
		return
	}


	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		log.Printf("error creating access token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "error creating access token")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("error creating refresh token: %s", err)
		respondWithError(w, http.StatusInternalServerError, "error creating refresh token")
		return
	}

	// Expire in 60 days
	refTokenExpiration := time.Now().AddDate(0, 0, 60)

	createRefreshTokenParams := database.CreateRefreshTokenParams{
		Token:		refreshToken,
		UserID:		user.ID,
		ExpiresAt:	refTokenExpiration,
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), createRefreshTokenParams)
	if err != nil { 
		log.Printf("error running CreateRefreshToken sql query: %s", err)
		respondWithError(w, http.StatusInternalServerError, "error creating refresh token")	
		return
	}

	respondWithJSON(w, 200, User{
		ID:		user.ID,
		CreatedAt:	user.CreatedAt,
		UpdatedAt:	user.UpdatedAt,
		Email:		user.Email,
		Token:		accessToken,
		RefreshToken:	refreshToken,
	})
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct { 
		Password	string `json:"password"`
		Email		string `json:"email"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request: %s", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't read request")
		return
	}

	params := parameters{}
	err = json.Unmarshal(dat, &params)
	if err != nil { 
		log.Printf("error unmarshaling parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't unmarshal parameters")
		return
	}
	if params.Password == "" { 
		log.Printf("password field in http request is empty")
		respondWithError(w, http.StatusInternalServerError, "password cannot be empty")
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil { 
		log.Printf("error hashing password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "error hashing password")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPasswords: hash,
	})
	if err != nil {
		log.Printf("user creation failed: %s", err)
		respondWithError(w, http.StatusInternalServerError, "user creation failed")
		return

	}
	
	respondWithJSON(w, 201, User{
		ID: user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email: user.Email,
	})	
}

func (cfg *apiConfig) handlerUsersUpdateCredentials(w http.ResponseWriter, r *http.Request) { 
	type parameters struct { 
		Email		string	`json:"email"`
		Password	string	`json:"password"`
	}

	dat, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request: %s", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't read request")
		return
	}

	params := parameters{}
	err = json.Unmarshal(dat, &params)
	if err != nil { 
		log.Printf("error unmarshaling request parameters: %s", err)
		respondWithError(w, http.StatusInternalServerError, "couldn't unmarshal request parameters")
		return
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil { 
		log.Printf("error retrieving token: %s", err)
		respondWithError(w, http.StatusUnauthorized, "401 Unauthorized")
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil { 
		log.Printf("error validating jwt: %s", err)
		respondWithError(w, http.StatusUnauthorized, "401 Unauthorized")
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil { 
		log.Printf("error hashing password: %s", err)
		respondWithError(w, http.StatusInternalServerError, "error hashing password")
		return
	}

	cfg.db.UpdateUserCredentials(r.Context(), database.UpdateUserCredentialsParams{
		Email:			params.Email,
		HashedPasswords:	hash,
		ID:			userID,
	})

	response := struct {
		Email string `json:"email"`
	}{
		Email: params.Email,
	}

	respondWithJSON(w, http.StatusOK, response)
}

