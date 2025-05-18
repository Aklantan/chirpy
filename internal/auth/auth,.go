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
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("%v : cannot hash password\n", err)
		return "", err
	}
	return string(hash), nil
}

func CheckPasswordHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		fmt.Printf("%v : incorrect password\n", err)
		return err

	}
	return nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		fmt.Printf("%v : cannot sign token\n", err)
		return "", err
	}
	fmt.Println(signedToken, err)
	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		fmt.Printf("%v : cannot parse token\n", err)
		return uuid.Nil, err
	}
	if !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}
	//expired := claims.ExpiresAt
	//if expired < time.Now() {
	//	return uuid.Nil, fmt.Errorf("expired token")
	//}
	userID := claims.Subject
	if err != nil {
		fmt.Printf("%v : incorrect token\n", err)
		return uuid.Nil, err
	}
	return uuid.Parse(userID)
}

func GetBearerToken(headers http.Header) (string, error) {
	const bearerPrefix = "Bearer "

	fullStr := headers.Get("Authorization")
	if fullStr == "" {
		return "", errors.New("authorization header missing")
	}

	if !strings.HasPrefix(fullStr, bearerPrefix) {
		return "", errors.New("authorization header is not a bearer token")
	}

	token := strings.TrimSpace(strings.TrimPrefix(fullStr, bearerPrefix))
	if token == "" {
		return "", errors.New("bearer token is empty")
	}

	return token, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", errors.New("unable to generate data")

	}
	refreshToken := hex.EncodeToString(key)
	return refreshToken, nil

}
