package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// MockCMSysDataByIPsRequest 基于IP列表查询的请求
type MockCMSysDataByIPsRequest struct {
	IPs []string `json:"ips"`
}

// MockCMSysDataByIPsHandler 模拟CMSys基于IP列表查询数据接口，用于开发测试
func MockCMSysDataByIPsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Info("收到Mock CMSys基于IP列表数据查询请求")

		// 获取请求头中的token和operator
		token := r.Header.Get("x-control-access-token")
		operator := r.Header.Get("x-control-access-operator")

		logx.Infof("Mock CMSys基于IP列表查询 - Token: %s, Operator: %s", token, operator)

		// 验证token（可选，开发测试时可以跳过）
		if token == "" {
			response := map[string]interface{}{
				"code": "A0001",
				"msg":  "未提供认证token",
				"data": nil,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// 解析请求体
		var req MockCMSysDataByIPsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response := map[string]interface{}{
				"code": "A0002",
				"msg":  "请求参数解析失败",
				"data": nil,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// 验证IP列表
		if len(req.IPs) == 0 {
			response := map[string]interface{}{
				"code": "A0003",
				"msg":  "IP列表不能为空",
				"data": nil,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		logx.Infof("Mock CMSys基于IP列表查询 - IP数量: %d", len(req.IPs))

		// 生成指定IP的模拟数据
		mockData := generateMockCMSysDataByIPs(req.IPs)

		// 返回响应
		response := map[string]interface{}{
			"code": "A000",
			"msg":  "success",
			"data": mockData,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// generateMockCMSysDataByIPs 根据IP列表生成模拟的CMSys主机数据
func generateMockCMSysDataByIPs(ips []string) []map[string]interface{} {
	rand.Seed(time.Now().UnixNano())

	remarks := []string{
		"生产环境数据库服务器",
		"测试环境应用服务器",
		"开发环境Web服务器",
		"备份服务器",
		"监控服务器",
		"日志服务器",
		"缓存服务器",
		"队列服务器",
		"存储服务器",
		"负载均衡服务器",
	}

	var data []map[string]interface{}
	for i, ip := range ips {
		// 生成随机但合理的监控数据
		cpuMax := 40.0 + rand.Float64()*50.0  // 40-90%
		memMax := 50.0 + rand.Float64()*40.0  // 50-90%
		diskMax := 30.0 + rand.Float64()*60.0 // 30-90%

		// 选择备注（循环使用）
		remark := remarks[i%len(remarks)]

		hostData := map[string]interface{}{
			"ipAddress":  ip,
			"cpuMaxNew":  fmt.Sprintf("%.2f", cpuMax),
			"memMaxNew":  fmt.Sprintf("%.2f", memMax),
			"diskMaxNew": fmt.Sprintf("%.2f", diskMax),
			"remark":     remark,
		}

		data = append(data, hostData)
	}

	logx.Infof("生成Mock CMSys基于IP列表响应 - 主机数量: %d", len(data))
	return data
}
