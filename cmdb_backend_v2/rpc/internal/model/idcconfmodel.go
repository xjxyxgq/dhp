package model

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"regexp"
)

var _ IdcConfModel = (*customIdcConfModel)(nil)

type (
	// IdcConfModel is an interface to be customized, add more methods here,
	// and implement the added methods in customIdcConfModel.
	IdcConfModel interface {
		idcConfModel
		withSession(session sqlx.Session) IdcConfModel
		FindAllActive(ctx context.Context) ([]*IdcConf, error)
		FindAllByPriority(ctx context.Context, activeOnly bool) ([]*IdcConf, error)
		MatchIdcByIp(ctx context.Context, hostIp string) (*IdcConf, error)
		SoftDelete(ctx context.Context, id uint64) error
	}

	customIdcConfModel struct {
		*defaultIdcConfModel
	}
)

// NewIdcConfModel returns a model for the database table.
func NewIdcConfModel(conn sqlx.SqlConn) IdcConfModel {
	return &customIdcConfModel{
		defaultIdcConfModel: newIdcConfModel(conn),
	}
}

func (m *customIdcConfModel) withSession(session sqlx.Session) IdcConfModel {
	return NewIdcConfModel(sqlx.NewSqlConnFromSession(session))
}

// FindAllActive 查找所有激活的IDC配置
func (m *customIdcConfModel) FindAllActive(ctx context.Context) ([]*IdcConf, error) {
	query := fmt.Sprintf("select %s from %s where `is_active` = 1 and `deleted_at` is null order by `priority` asc", idcConfRows, m.table)
	var resp []*IdcConf
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

// FindAllByPriority 按优先级查找所有IDC配置
func (m *customIdcConfModel) FindAllByPriority(ctx context.Context, activeOnly bool) ([]*IdcConf, error) {
	var query string
	if activeOnly {
		query = fmt.Sprintf("select %s from %s where `is_active` = 1 and `deleted_at` is null order by `priority` asc", idcConfRows, m.table)
	} else {
		query = fmt.Sprintf("select %s from %s where `deleted_at` is null order by `priority` asc", idcConfRows, m.table)
	}
	var resp []*IdcConf
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

// MatchIdcByIp 根据主机IP匹配对应的IDC配置
func (m *customIdcConfModel) MatchIdcByIp(ctx context.Context, hostIp string) (*IdcConf, error) {
	// 先获取所有激活的IDC配置，按优先级排序
	configs, err := m.FindAllActive(ctx)
	if err != nil {
		return nil, err
	}

	// 按优先级遍历配置，尝试匹配IP
	for _, config := range configs {
		matched, err := regexp.MatchString(config.IdcIpRegexp, hostIp)
		if err != nil {
			// 正则表达式有错误，跳过这个配置
			continue
		}
		if matched {
			return config, nil
		}
	}

	// 没有匹配到任何IDC配置
	return nil, sql.ErrNoRows
}

// SoftDelete 软删除IDC配置
func (m *customIdcConfModel) SoftDelete(ctx context.Context, id uint64) error {
	query := fmt.Sprintf("update %s set `deleted_at` = now() where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
