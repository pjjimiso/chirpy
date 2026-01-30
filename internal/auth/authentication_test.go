package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashValidatePassword(t *testing.T) {
	password1 := "p@ssword123"
	password2 := "someotherpassword"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name		string
		password	string
		hash		string
		wantMatch	bool
		wantErr		bool
	}{
		{
			name:		"Correct password",
			password:	password1,
			hash:		hash1,
			wantMatch:	true,
			wantErr:	false,
		},
		{
			name:		"Incorrect password",
			password:	"wrongpassword",
			hash:		hash1,
			wantMatch:	false,
			wantErr:	false,
		},
		{
			name:		"Password doesn't match different hash",
			password:	password1,
			hash:		hash2,
			wantMatch:	false,
			wantErr:	false,
		},
		{
			name:		"Empty password",
			password:	"",
			hash:		hash1,
			wantMatch:	false,
			wantErr:	false,
		},
		{
			name:		"Invalid hash",
			password:	password1,
			hash:		"invalidhash",
			wantErr:	true,
			wantMatch:	false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if match != tt.wantMatch {
				t.Errorf("HashPassword() match = %v, wantMatch %v", match, tt.wantMatch)
				return
			}
		})

	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	validToken, _ := MakeJWT(userID, "itsasecret", 3 * time.Second)

	tests := []struct {
		name		string
		tokenString	string
		tokenSecret	string
		sleepDuration	time.Duration
		wantUserID	uuid.UUID
		wantErr		bool
	}{
		{
			name:		"Valid token",
			tokenString:	validToken,
			tokenSecret:	"itsasecret",
			sleepDuration:	0 * time.Second,
			wantUserID:	userID,
			wantErr:	false,

		},
		{
			name:		"Invalid token",
			tokenString:	"invalid.token.string",
			tokenSecret:	"itsasecret",
			sleepDuration:	0 * time.Second,
			wantUserID:	uuid.Nil,
			wantErr:	true,
		},
		{
			name:		"Wrong secret",
			tokenString:	validToken,
			tokenSecret:	"wrongsecret",
			sleepDuration:	0 * time.Second,
			wantUserID:	uuid.Nil,
			wantErr:	true,
		},
		{
			name:		"Expired token",
			tokenString:	validToken,
			tokenSecret:	"itsasecret",
			sleepDuration:	3 * time.Second,
			wantUserID:	uuid.Nil,
			wantErr:	true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			time.Sleep(tt.sleepDuration)
			gotUserID, err := ValidateJWT(tt.tokenString, tt.tokenSecret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotUserID != tt.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, wantUserID %v", gotUserID, tt.wantUserID)
			}
		})
	}
}

