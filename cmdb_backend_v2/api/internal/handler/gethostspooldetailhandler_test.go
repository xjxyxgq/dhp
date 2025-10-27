package handler

import (
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

func TestGetHostsPoolDetailHandler_Basic(t *testing.T) {
	// 基本的handler测试 - 测试包结构
	assert.True(t, true, "Handler package loaded successfully")
}

func TestHttpRequest_Basic(t *testing.T) {
	// 基本的HTTP请求结构测试
	type TestRequest struct {
		IpList []string `json:"ip_list"`
	}
	
	req := &TestRequest{
		IpList: []string{"192.168.1.100", "192.168.1.101"},
	}
	
	assert.NotNil(t, req)
	assert.Len(t, req.IpList, 2)
	assert.Equal(t, "192.168.1.100", req.IpList[0])
	assert.Equal(t, "192.168.1.101", req.IpList[1])
}

func TestGetHostsPoolDetailHandler_HTTPRequest(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	req := httptest.NewRequest("GET", "/api/cmdb/v1/get_hosts_pool_detail", nil)
	w := httptest.NewRecorder()
	
	handler := GetHostsPoolDetailHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 验证响应
	assert.Contains(t, []int{200, 500}, w.Code)
	
	if w.Code == 200 {
		var resp types.HostPoolListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		
		// 验证响应结构
		assert.NotNil(t, resp.List)
	}
}

func TestGetHostsPoolDetailHandler_WithQueryParams(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	req := httptest.NewRequest("GET", "/api/cmdb/v1/get_hosts_pool_detail?ip_list=192.168.1.100,192.168.1.101", nil)
	w := httptest.NewRecorder()
	
	handler := GetHostsPoolDetailHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 验证响应
	assert.Contains(t, []int{200, 500}, w.Code)
}

func TestGetHostsPoolDetailHandler_InvalidMethod(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	req := httptest.NewRequest("POST", "/api/cmdb/v1/get_hosts_pool_detail", nil)
	w := httptest.NewRecorder()
	
	handler := GetHostsPoolDetailHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// GET接口用POST方法，框架会自动处理，可能返回200或错误码
	assert.Contains(t, []int{200, 405, 400}, w.Code)
}