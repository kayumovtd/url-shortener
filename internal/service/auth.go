package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserProvider interface {
	GetUserID(ctx context.Context) (string, bool)
}

type AuthService struct {
	secret []byte
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{secret: []byte(secret)}
}

type Claims struct {
	UserID string `json:"userId"`
	jwt.RegisteredClaims
}

func (s *AuthService) GenerateNewToken() (userID string, token string, err error) {
	id := uuid.NewString()
	tok, err := s.generateToken(id)
	return id, tok, err
}

func (s *AuthService) ParseAndVerify(tokenStr string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.secret, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}

func (s *AuthService) generateToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(365 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

type userIDCtxKey struct{}

func (s *AuthService) WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDCtxKey{}, userID)
}

func (s *AuthService) GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDCtxKey{}).(string)
	return id, ok
}
