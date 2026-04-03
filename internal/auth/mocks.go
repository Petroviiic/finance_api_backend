package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type MockJWTAuthenticator struct {
}

func NewMockJWTAuthenticator() *MockJWTAuthenticator {
	return &MockJWTAuthenticator{}
}

func (a *MockJWTAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(""))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
func (a *MockJWTAuthenticator) ValidateToken(plainToken string) (*jwt.Token, error) {
	return jwt.Parse(plainToken, func(t *jwt.Token) (any, error) {
		return []byte(""), nil
	})
}
