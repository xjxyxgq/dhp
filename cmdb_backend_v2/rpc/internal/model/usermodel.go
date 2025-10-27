package model

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ UserModel = (*customUserModel)(nil)

type (
	// UserModel 用户模型接口
	UserModel interface {
		Insert(user *User) (sql.Result, error)
		FindOne(id int64) (*User, error)
		FindByUsername(username string) (*User, error) 
		FindByEmail(email string) (*User, error)
		Update(user *User) error
		Delete(id int64) error
		UpdateLoginTime(id int64, loginTime time.Time) error
	}

	customUserModel struct {
		*defaultUserModel
	}

	User struct {
		Id           int64          `db:"id"`
		CreateTime   time.Time      `db:"create_time"`
		UpdateTime   time.Time      `db:"update_time"`
		DeleteTime   sql.NullTime   `db:"delete_time"`
		IsDeleted    int64          `db:"is_deleted"`
		Username     string         `db:"username"`
		PasswordHash sql.NullString `db:"password_hash"`
		Email        sql.NullString `db:"email"`
		DisplayName  sql.NullString `db:"display_name"`
		IsActive     bool           `db:"is_active"`
		IsAdmin      bool           `db:"is_admin"`
		LastLoginAt  sql.NullTime   `db:"last_login_at"`
		LoginSource  string         `db:"login_source"`
	}

	defaultUserModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewUserModel 创建用户模型
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &customUserModel{
		defaultUserModel: &defaultUserModel{
			conn:  conn,
			table: "`users`",
		},
	}
}

// Insert 插入用户
func (m *customUserModel) Insert(user *User) (sql.Result, error) {
	query := fmt.Sprintf("INSERT INTO %s (`username`, `password_hash`, `email`, `display_name`, `is_active`, `is_admin`, `login_source`) VALUES (?, ?, ?, ?, ?, ?, ?)", m.table)
	return m.conn.Exec(query, user.Username, user.PasswordHash, user.Email, user.DisplayName, user.IsActive, user.IsAdmin, user.LoginSource)
}

// FindOne 根据ID查找用户
func (m *customUserModel) FindOne(id int64) (*User, error) {
	var resp User
	query := fmt.Sprintf("SELECT `id`, `create_time`, `update_time`, `delete_time`, `is_deleted`, `username`, `password_hash`, `email`, `display_name`, `is_active`, `is_admin`, `last_login_at`, `login_source` FROM %s WHERE `id` = ? AND `is_deleted` = 0", m.table)
	err := m.conn.QueryRow(&resp, query, id)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindByUsername 根据用户名查找用户
func (m *customUserModel) FindByUsername(username string) (*User, error) {
	var resp User
	query := fmt.Sprintf("SELECT `id`, `create_time`, `update_time`, `delete_time`, `is_deleted`, `username`, `password_hash`, `email`, `display_name`, `is_active`, `is_admin`, `last_login_at`, `login_source` FROM %s WHERE `username` = ? AND `is_deleted` = 0", m.table)
	err := m.conn.QueryRow(&resp, query, username)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// FindByEmail 根据邮箱查找用户
func (m *customUserModel) FindByEmail(email string) (*User, error) {
	var resp User
	query := fmt.Sprintf("SELECT `id`, `create_time`, `update_time`, `delete_time`, `is_deleted`, `username`, `password_hash`, `email`, `display_name`, `is_active`, `is_admin`, `last_login_at`, `login_source` FROM %s WHERE `email` = ? AND `is_deleted` = 0", m.table)
	err := m.conn.QueryRow(&resp, query, email)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Update 更新用户
func (m *customUserModel) Update(user *User) error {
	query := fmt.Sprintf("UPDATE %s SET `password_hash` = ?, `email` = ?, `display_name` = ?, `is_active` = ?, `is_admin` = ?, `update_time` = CURRENT_TIMESTAMP WHERE `id` = ? AND `is_deleted` = 0", m.table)
	_, err := m.conn.Exec(query, user.PasswordHash, user.Email, user.DisplayName, user.IsActive, user.IsAdmin, user.Id)
	return err
}

// Delete 删除用户（软删除）
func (m *customUserModel) Delete(id int64) error {
	query := fmt.Sprintf("UPDATE %s SET `is_deleted` = 1, `delete_time` = CURRENT_TIMESTAMP WHERE `id` = ?", m.table)
	_, err := m.conn.Exec(query, id)
	return err
}

// UpdateLoginTime 更新最后登录时间
func (m *customUserModel) UpdateLoginTime(id int64, loginTime time.Time) error {
	query := fmt.Sprintf("UPDATE %s SET `last_login_at` = ?, `update_time` = CURRENT_TIMESTAMP WHERE `id` = ? AND `is_deleted` = 0", m.table)
	_, err := m.conn.Exec(query, loginTime, id)
	return err
}