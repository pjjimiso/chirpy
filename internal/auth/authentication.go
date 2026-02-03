package auth

import (
	"fmt"
	"time"
	"net/http"
	"regexp"
	"crypto/rand"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil { 
		return "", err
	}
	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil { 
		return match, fmt.Errorf("error comparing password hash: %s", err)
	}
	return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	claims := &jwt.RegisteredClaims{
		IssuedAt:	jwt.NewNumericDate(now),
		ExpiresAt:	jwt.NewNumericDate(now.Add(expiresIn)),
		Issuer:		"chirpy",
		Subject:	userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tokenSecret))
	if err != nil { 
		return "", err
	}
	
	return tokenString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil { 
		return uuid.Nil, fmt.Errorf("validating jwt: %s", err)
	}
	if !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil { 
		return uuid.Nil, fmt.Errorf("error parsing id into uuid")
	}
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	param := headers.Get("Authorization")
	if param == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	re := regexp.MustCompile(`Bearer\s*`)
	token := re.ReplaceAllString(param, "")
	if token == "" {
		return "", fmt.Errorf("token string is empty")
	}

	return token, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	rand.Read(key)
	encodedKey := hex.EncodeToString(key)
	return encodedKey, nil
}

