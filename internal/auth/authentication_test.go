package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "p@ssword123"
	hash, err := HashPassword(password)

	if hash == "" || err != nil {
		t.Errorf(`HashPassword("%s") = %s, %v, want non-empty string`, password, hash, err)
	} 

	match, err := CheckPasswordHash(password, hash)

	if !match || err != nil { 
		t.Errorf(`CheckPasswordHash("%s","%s") = %t, %v, want true`, password, hash, match, err)
	}

	newPass := "123password"
	match, err = CheckPasswordHash(newPass, hash)

	if match || err != nil {
		t.Errorf(`CheckPasswordHash("%s","%s") = %t, %v, want false`, newPass, hash, match, err)
	}
}

