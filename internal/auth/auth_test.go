package auth

import (
	"fmt"
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

func TestMakeJWT(t *testing.T) {
	userID, _ := uuid.Parse("a0de6ac2-0b73-432c-a5db-b96d6451251f")
	tokenSecret := "test"
	expiresIn := time.Duration.Seconds(30)

	jwt, err := MakeJWT(userID, tokenSecret, time.Duration(expiresIn))
	if err != nil {
		t.Errorf("MakeJWT returned a non nil error: %v", err)
	}

	fmt.Println(jwt)
}
