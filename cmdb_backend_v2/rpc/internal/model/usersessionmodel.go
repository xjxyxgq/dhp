package model

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserSessionModel = (*customUserSessionModel)(nil)

type (
	// UserSessionModel 用户会话模型接口
	UserSessionModel interface {
		Insert(session *UserSession) (sql.Result, error)
		FindByToken(token string) (*UserSession, error)
		FindByUserId(userId int64) ([]*UserSession, error)
		UpdateExpireTime(id int64, expireTime time.Time) error
		Deactivate(id int64) error
		DeactivateByUserId(userId int64) error
		CleanExpiredSessions() error
	}

	customUserSessionModel struct {
		*defaultUserSessionModel
	}

	UserSession struct {
		Id           int64          `db:"id"`
		CreateTime   time.Time      `db:"create_time"`
		UpdateTime   time.Time      `db:"update_time"`
		DeleteTime   sql.NullTime   `db:"delete_time"`
		IsDeleted    int64          `db:"is_deleted"`
		UserId       int64          `db:"user_id"`
		SessionToken string         `db:"session_token"`
		CasTicket    sql.NullString `db:"cas_ticket"`
		ExpiresAt    sql.NullTime   `db:"expires_at"`
		IpAddress    sql.NullString `db:"ip_address"`
		UserAgent    sql.NullString `db:"user_agent"`
		IsActive     bool           `db:"is_active"`
	}

	defaultUserSessionModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewUserSessionModel 创建用户会话模型
func NewUserSessionModel(conn sqlx.SqlConn) UserSessionModel {
	return &customUserSessionModel{
		defaultUserSessionModel: &defaultUserSessionModel{
			conn:  conn,
			table: "`user_sessions`",
		},
	}
}

// Insert 插入会话
func (m *customUserSessionModel) Insert(session *UserSession) (sql.Result, error) {
	query := fmt.Sprintf("INSERT INTO %s (`user_id`, `session_token`, `cas_ticket`, `expires_at`, `ip_address`, `user_agent`, `is_active`) VALUES (?, ?, ?, ?, ?, ?, ?)", m.table)
	return m.conn.Exec(query, session.UserId, session.SessionToken, session.CasTicket, session.ExpiresAt, session.IpAddress, session.UserAgent, session.IsActive)
}

// FindByToken 根据令牌查找会话
func (m *customUserSessionModel) FindByToken(token string) (*UserSession, error) {
	var resp UserSession
	query := fmt.Sprintf("SELECT `id`, `create_time`, `update_time`, `delete_time`, `is_deleted`, `user_id`, `session_token`, `cas_ticket`, `expires_at`, `ip_address`, `user_agent`, `is_active` FROM %s WHERE `session_token` = ? AND `is_active` = 1 AND `expires_at` > NOW() AND `is_deleted` = 0", m.table)
	err := m.conn.QueryRow(&resp, query, token)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindByUserId 根据用户ID查找会话
func (m *customUserSessionModel) FindByUserId(userId int64) ([]*UserSession, error) {
	var resp []*UserSession
	query := fmt.Sprintf("SELECT `id`, `create_time`, `update_time`, `delete_time`, `is_deleted`, `user_id`, `session_token`, `cas_ticket`, `expires_at`, `ip_address`, `user_agent`, `is_active` FROM %s WHERE `user_id` = ? AND `is_active` = 1 AND `is_deleted` = 0 ORDER BY `create_time` DESC", m.table)
	err := m.conn.QueryRows(&resp, query, userId)
	return resp, err
}

// UpdateExpireTime 更新过期时间
func (m *customUserSessionModel) UpdateExpireTime(id int64, expireTime time.Time) error {
	query := fmt.Sprintf("UPDATE %s SET `expires_at` = ?, `update_time` = CURRENT_TIMESTAMP WHERE `id` = ?", m.table)
	_, err := m.conn.Exec(query, expireTime, id)
	return err
}

// Deactivate 停用会话
func (m *customUserSessionModel) Deactivate(id int64) error {
	query := fmt.Sprintf("UPDATE %s SET `is_active` = 0, `update_time` = CURRENT_TIMESTAMP WHERE `id` = ?", m.table)
	_, err := m.conn.Exec(query, id)
	return err
}

// DeactivateByUserId 停用用户的所有会话
func (m *customUserSessionModel) DeactivateByUserId(userId int64) error {
	query := fmt.Sprintf("UPDATE %s SET `is_active` = 0, `update_time` = CURRENT_TIMESTAMP WHERE `user_id` = ?", m.table)
	_, err := m.conn.Exec(query, userId)
	return err
}

// CleanExpiredSessions 清理过期会话
func (m *customUserSessionModel) CleanExpiredSessions() error {
	query := fmt.Sprintf("UPDATE %s SET `is_active` = 0, `update_time` = CURRENT_TIMESTAMP WHERE `expires_at` <= NOW() AND `is_active` = 1", m.table)
	_, err := m.conn.Exec(query)
	return err
}