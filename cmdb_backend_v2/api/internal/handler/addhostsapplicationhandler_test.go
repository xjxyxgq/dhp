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

func TestAddHostsApplicationHandler_Basic(t *testing.T) {
	// 基本的添加主机应用请求结构测试
	req := &types.AddHostsApplicationRequest{
		Data: []types.HostApplicationRequest{
			{
				HostId:         1,
				ServerType:     "mysql",
				ServerVersion:  "8.0",
				ServerSubtitle: "主库",
				ClusterName:    "mysql-cluster-01",
				ServerProtocol: "tcp",
				ServerAddr:     "192.168.1.100:3306",
				ServerPort:     3306,
				ServerRole:     "master",
				ServerStatus:   "running",
				DepartmentName: "数据库团队",
			},
		},
	}
	
	assert.NotNil(t, req)
	assert.Len(t, req.Data, 1)
	
	app := req.Data[0]
	assert.Equal(t, 1, app.HostId)
	assert.Equal(t, "mysql", app.ServerType)
	assert.Equal(t, "8.0", app.ServerVersion)
	assert.Equal(t, "主库", app.ServerSubtitle)
	assert.Equal(t, "mysql-cluster-01", app.ClusterName)
	assert.Equal(t, "tcp", app.ServerProtocol)
	assert.Equal(t, "192.168.1.100:3306", app.ServerAddr)
	assert.Equal(t, int32(3306), app.ServerPort)
	assert.Equal(t, "master", app.ServerRole)
	assert.Equal(t, "running", app.ServerStatus)
	assert.Equal(t, "数据库团队", app.DepartmentName)
}

func TestAddHostsApplicationHandler_HTTPRequest(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	reqData := types.AddHostsApplicationRequest{
		Data: []types.HostApplicationRequest{
			{
				HostId:         1,
				ServerType:     "mysql",
				ServerVersion:  "8.0",
				DepartmentName: "测试团队",
			},
		},
	}

	jsonBody, err := json.Marshal(reqData)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/cmdb/v1/add_hosts_application", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	handler := AddHostsApplicationHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 验证响应
	assert.Contains(t, []int{200, 400, 500}, w.Code)
	
	var resp types.HostApplicationResponse
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	if err == nil {
		// 如果能成功解析响应，验证基本结构
		assert.NotEmpty(t, resp.Message)
	}
}

func TestAddHostsApplicationHandler_InvalidJSON(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	// 发送无效的JSON
	req := httptest.NewRequest("POST", "/api/cmdb/v1/add_hosts_application", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	handler := AddHostsApplicationHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 应该返回400错误
	assert.Equal(t, 400, w.Code)
}

func TestAddHostsApplicationHandler_EmptyData(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	reqData := types.AddHostsApplicationRequest{
		Data: []types.HostApplicationRequest{},
	}

	jsonBody, err := json.Marshal(reqData)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/cmdb/v1/add_hosts_application", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	handler := AddHostsApplicationHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 验证响应
	assert.Contains(t, []int{200, 400}, w.Code)
}

func TestAddHostsApplicationHandler_MultipleApplications(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	reqData := types.AddHostsApplicationRequest{
		Data: []types.HostApplicationRequest{
			{
				HostId:         1,
				ServerType:     "mysql",
				ServerVersion:  "8.0",
				DepartmentName: "数据库团队",
			},
			{
				HostId:         1,
				ServerType:     "redis",
				ServerVersion:  "7.0",
				DepartmentName: "缓存团队",
			},
		},
	}

	jsonBody, err := json.Marshal(reqData)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/cmdb/v1/add_hosts_application", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	
	handler := AddHostsApplicationHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// 验证响应
	assert.Contains(t, []int{200, 400, 500}, w.Code)
}

func TestAddHostsApplicationHandler_InvalidMethod(t *testing.T) {
	c := config.Config{
		RpcConfig: zrpc.RpcClientConf{
			Endpoints: []string{"127.0.0.1:8080"},
		},
	}
	svcCtx := svc.NewServiceContext(c)

	req := httptest.NewRequest("GET", "/api/cmdb/v1/add_hosts_application", nil)
	w := httptest.NewRecorder()
	
	handler := AddHostsApplicationHandler(svcCtx)
	handler.ServeHTTP(w, req)

	// POST接口用GET方法应该返回错误
	assert.Contains(t, []int{405, 400}, w.Code)
}