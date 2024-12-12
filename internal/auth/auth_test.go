package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashed(t *testing.T) {
	password := "testingpassword"
	hashed, err := HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword(%s) returned an err: %v", password, err)
	}

	if len(hashed) == 0 {
		t.Errorf("HashPassword(%s) func returned empty hash", password)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testingpassword"
	hashed, _ := HashPassword(password)

	err := CheckPasswordHash(password, hashed)
	if err != nil {
		t.Error("Password expect to match, but did not")
	}

	wrongPassword := "wrong"

	err = CheckPasswordHash(wrongPassword, hashed)
	if err == nil {
		t.Error("Password expect to not match, but did match")
	}
}

func TestJWT(t *testing.T) {
	userID, _ := uuid.Parse("a0de6ac2-0b73-432c-a5db-b96d6451251f")
	tokenSecret := "test"
	wrongSecret := "wrong"

	JWT, err := MakeJWT(userID, tokenSecret, time.Second*30)
	if err != nil {
		t.Errorf("MakeJWT returned a non nil error: %v", err)
	}

	validatedUser, err := ValidateJWT(JWT, tokenSecret)
	if err != nil {
		t.Errorf("ValidateJWT(JWT, %s) returning error when trying to validate JWT: %v", tokenSecret, err)
	}

	if validatedUser != userID {
		t.Errorf("validatedUser: %s, is not matching the provided userID: %s", validatedUser, userID)
	}

	_, err = ValidateJWT(JWT, wrongSecret)
	if err == nil {
		t.Errorf("ValidateJWT(JWT, %s) validated token when it should of failed wrong secret: %v", wrongSecret, err)
	}

	expiredJWT, err := MakeJWT(userID, tokenSecret, time.Second*0)
	if err != nil {
		t.Errorf("MakeJWT returned a non nil error: %v", err)
	}

	_, err = ValidateJWT(expiredJWT, tokenSecret)
	if err == nil {
		t.Errorf("ValidateJWT(expiredJWT, %s) validated token when it should of failed: expired", tokenSecret)
	}

}
