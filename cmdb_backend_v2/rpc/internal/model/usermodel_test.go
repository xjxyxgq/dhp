package model

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUsers_Basic(t *testing.T) {
	// 测试 Users 结构体的基本功能
	user := &Users{
		Id:           1,
		Username:     "testuser",
		Email:        sql.NullString{String: "test@example.com", Valid: true},
		DisplayName:  sql.NullString{String: "Test User", Valid: true},
		PasswordHash: sql.NullString{String: "hashed_password", Valid: true},
		IsActive:     1,
		IsAdmin:      0,
		LoginSource:  "local",
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	
	assert.NotNil(t, user)
	assert.Equal(t, uint64(1), user.Id)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email.String)
	assert.True(t, user.Email.Valid)
	assert.Equal(t, "Test User", user.DisplayName.String)
	assert.True(t, user.DisplayName.Valid)
	assert.Equal(t, "hashed_password", user.PasswordHash.String)
	assert.True(t, user.PasswordHash.Valid)
	assert.Equal(t, int64(1), user.IsActive)
	assert.Equal(t, int64(0), user.IsAdmin)
	assert.Equal(t, "local", user.LoginSource)
}

func TestUserSessions_Basic(t *testing.T) {
	// 测试 UserSessions 结构体的基本功能
	session := &UserSessions{
		Id:           1,
		UserId:       1,
		SessionToken: "test-token-123",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	
	assert.NotNil(t, session)
	assert.Equal(t, uint64(1), session.Id)
	assert.Equal(t, uint64(1), session.UserId)
	assert.Equal(t, "test-token-123", session.SessionToken)
	assert.True(t, session.ExpiresAt.After(time.Now()))
}