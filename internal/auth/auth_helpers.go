package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedByt, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedByt), err
}

func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})

	return token.SignedString([]byte(tokenSecret))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	if !token.Valid {
		return uuid.UUID{}, fmt.Errorf("invalid token")
	}

	strUserID, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}

	uuidUserID, err := uuid.Parse(strUserID)
	if err != nil {
		return uuid.UUID{}, err
	}

	return uuidUserID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authReqHeader := headers.Get("Authorization")
	if authReqHeader == "" {
		return "", fmt.Errorf("Empty Authorization header")
	}

	splitHeader := strings.Split(authReqHeader, " ")
	if splitHeader[0] != "Bearer" {
		return "", fmt.Errorf("Authorization header is not a Bearer token")
	}

	token := splitHeader[1]
	if token == "" {
		return "", fmt.Errorf("Empty bearer token")
	}

	return strings.TrimSpace(token), nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authReqHeader := headers.Get("Authorization")
	if authReqHeader == "" {
		return "", errors.New("Empty Authorization header")
	}

	splitHeader := strings.Split(authReqHeader, " ")
	if splitHeader[0] != "ApiKey" {
		return "", fmt.Errorf("Authorization header is not an ApiKey")
	}

	key := splitHeader[1]
	if key == "" {
		return "", fmt.Errorf("Empty bearer token")
	}

	return strings.TrimSpace(key), nil
}

func MakeRefreshToken() (string, error) {
	ranData := make([]byte, 32)
	_, err := rand.Read(ranData)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(ranData), nil
}
