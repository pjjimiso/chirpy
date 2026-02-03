package main

import (
	"net/http"
	"encoding/json"
	"io"
	"time"

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
		respondWithError(w, http.StatusInternalServerError, "Couldn't read request parameters", err)
		return
	}

	params := parameters{}
	err = json.Unmarshal(dat, &params)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't unmarshal json parameters", err)
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
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPasswords)
	if !match || err != nil { 
		respondWithError(w, http.StatusUnauthorized, "Invalid password", err)
		return
	}


	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create session token", err)
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
		respondWithError(w, http.StatusInternalServerError, "Couldn't create session", err)	
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
		respondWithError(w, http.StatusInternalServerError, "Couldn't read request parameters", err)
		return
	}

	params := parameters{}
	err = json.Unmarshal(dat, &params)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't unmarshal json parameters", err)
		return
	}
	if params.Password == "" { 
		respondWithError(w, http.StatusInternalServerError, "Password field can't be empty", nil)
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email: params.Email,
		HashedPasswords: hash,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
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
		respondWithError(w, http.StatusInternalServerError, "Couldn't read request parameters", err)
		return
	}

	params := parameters{}
	err = json.Unmarshal(dat, &params)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't unmarshal json parameters", err)
		return
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil { 
		respondWithError(w, http.StatusUnauthorized, "Couldn't get token", err)
		return
	}

	userID, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil { 
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate token", err)
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil { 
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	// TODO modify DB query to update the updated_at field

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

