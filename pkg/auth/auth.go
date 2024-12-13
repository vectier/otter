package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var signingKey = "key"

var (
	GrantedAllDirectories = []string{"**"}
)

type Claims struct {
	jwt.RegisteredClaims

	// Use glob pattern
	GrantedDirectories []string `json:"grantedDirectories"`
	// Client who connect Otter
	Client string `json:"client"`
}

func CreateToken(client string, directories []string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:  uuid.NewString(),
			Audience: jwt.ClaimStrings{"otter"},
			Issuer:   "otter",
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		GrantedDirectories: directories,
		Client:             client,
	})
	return token.SignedString([]byte(signingKey))
}

func VerifyToken(input string) (*Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(input, &claims, getVerifyKey)
	if err != nil {
		return nil, fmt.Errorf("cannot parse token: %w", err)
	}
	return &claims, nil
}

func getVerifyKey(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}
	return []byte(signingKey), nil
}

func GetClaims(r *http.Request) (*Claims, error) {
	// Try to get token from header first, then getting from cookie
	bearer := r.Header.Get("Authorization")
	if strings.HasPrefix(bearer, "bearer ") {
		token := bearer[7:]
		return VerifyToken(token)
	}
	cookie, err := r.Cookie("accessToken")
	if err != nil {
		return nil, errors.New("cannot get access token from cookie")
	}
	return VerifyToken(cookie.Value)
}
