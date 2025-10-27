package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims JWT载荷
type Claims struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
	LoginSource string `json:"login_source"`
	jwt.RegisteredClaims
}

// JWTService JWT服务
type JWTService struct {
	secret    []byte
	expireHours int
}

// NewJWTService 创建JWT服务
func NewJWTService(secret string, expireHours int) *JWTService {
	return &JWTService{
		secret:      []byte(secret),
		expireHours: expireHours,
	}
}

// GenerateToken 生成JWT令牌
func (j *JWTService) GenerateToken(userID int64, username, displayName string, isAdmin bool, loginSource string) (string, error) {
	if j.expireHours <= 0 {
		j.expireHours = 24 // 默认24小时
	}
	
	claims := Claims{
		UserID:      userID,
		Username:    username,
		DisplayName: displayName,
		IsAdmin:     isAdmin,
		LoginSource: loginSource,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(j.expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "cmdb-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

// ValidateToken 验证JWT令牌
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func (j *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 如果令牌还有超过1小时才过期，不需要刷新
	if claims.ExpiresAt.Time.Sub(time.Now()) > time.Hour {
		return tokenString, nil
	}

	// 生成新令牌
	return j.GenerateToken(claims.UserID, claims.Username, claims.DisplayName, claims.IsAdmin, claims.LoginSource)
}