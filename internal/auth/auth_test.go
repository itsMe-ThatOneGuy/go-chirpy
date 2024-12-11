package auth

import "testing"

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
