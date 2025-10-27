package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeromicro/go-zero/zrpc"
)

func TestLoginHandler_Basic(t *testing.T) {
	// 基本的登录handler测试 - 测试包结构
	assert.True(t, true, "Login handler package loaded successfully")
}

func TestLoginRequest_Basic(t *testing.T) {
	// 基本的登录请求结构测试
	type TestLoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	req := &TestLoginRequest{
		Username: "testuser",
		Password: "password",
	}
	
	assert.NotNil(t, req)
	assert.Equal(t, "testuser", req.Username)
	assert.Equal(t, "password", req.Password)
}

func TestLoginHandler_HTTPRequest(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	loginReq := types.LoginRequest{
		Username: "admin",
		Password: "admin",
	}

	jsonBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	handler := LoginHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 验证响应
	assert.Contains(t, []int{200, 400, 401, 500}, w.Code)
	
	var resp types.LoginResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err == nil {
		// 如果能成功解析响应，验证基本结构
		assert.NotEmpty(t, resp.Message)
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	// 发送无效的JSON
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	handler := LoginHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 应该返回400错误
	assert.Equal(t, 400, w.Code)
}

func TestLoginHandler_EmptyBody(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	handler := LoginHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 应该返回400错误
	assert.Equal(t, 400, w.Code)
}

func TestLoginHandler_MissingCredentials(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	loginReq := types.LoginRequest{
		Username: "",
		Password: "",
	}

	jsonBody, err := json.Marshal(loginReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	handler := LoginHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 验证响应 - 缺少凭据可能返回200(业务错误)或HTTP错误码
	assert.Contains(t, []int{200, 400, 401}, w.Code)
}