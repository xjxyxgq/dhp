package middleware

import (
	"context"
	"net/http"
	"strings"

	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type AuthMiddleware struct {
	rpcClient cmpool.CmpoolClient
}

// UserContext 用户上下文
type UserContext struct {
	UserID      int64
	Username    string
	DisplayName string
	IsAdmin     bool
	LoginSource string
}

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(rpcClient cmpool.CmpoolClient) *AuthMiddleware {
	return &AuthMiddleware{
		rpcClient: rpcClient,
	}
}

// Handle 中间件处理函数
func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 获取Authorization头
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "缺少Authorization头",
				"data":    nil,
			})
			return
		}

		// 提取Bearer token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": "无效的Authorization格式",
				"data":    nil,
			})
			return
		}

		// 通过RPC验证token
		resp, err := m.rpcClient.ValidateToken(r.Context(), &cmpool.ValidateTokenReq{
			Token: token,
		})
		if err != nil {
			logx.Errorf("RPC验证token失败: %v", err)
			httpx.WriteJson(w, http.StatusInternalServerError, map[string]interface{}{
				"code":    500,
				"message": "认证服务异常",
				"data":    nil,
			})
			return
		}

		if !resp.Valid {
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]interface{}{
				"code":    401,
				"message": resp.Message,
				"data":    nil,
			})
			return
		}

		// 将用户信息存储到上下文中
		userCtx := &UserContext{
			UserID:      resp.UserInfo.Id,
			Username:    resp.UserInfo.Username,
			DisplayName: resp.UserInfo.DisplayName,
			IsAdmin:     resp.UserInfo.IsAdmin,
			LoginSource: resp.UserInfo.LoginSource,
		}

		ctx := context.WithValue(r.Context(), "user", userCtx)
		next(w, r.WithContext(ctx))
	}
}

// GetUserFromContext 从上下文中获取用户信息
func GetUserFromContext(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value("user").(*UserContext)
	return user, ok
}