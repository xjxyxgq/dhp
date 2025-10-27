package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
	LoginSource string `json:"login_source"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret      []byte
	expireHours int
}

func NewJWTService(secret string, expireHours int) *JWTService {
	return &JWTService{
		secret:      []byte(secret),
		expireHours: expireHours,
	}
}

func (j *JWTService) GenerateToken(userID int64, username, displayName string, isAdmin bool, loginSource string) (string, error) {
	claims := Claims{
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
		IsAdmin:     isAdmin,
		LoginSource: loginSource,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * time.Duration(j.expireHours))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 生成新的token
	return j.GenerateToken(claims.UserID, claims.Username, claims.DisplayName, claims.IsAdmin, claims.LoginSource)
}