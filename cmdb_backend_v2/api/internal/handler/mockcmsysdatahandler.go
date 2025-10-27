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

// MockCMSysDataHandler 模拟CMSys数据接口，用于开发测试
func MockCMSysDataHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Info("收到Mock CMSys数据查询请求")

		// 获取请求头中的token和operator
		token := r.Header.Get("x-control-access-token")
		operator := r.Header.Get("x-control-access-operator")

		// 获取查询参数
		query := r.URL.Query().Get("query")

		logx.Infof("Mock CMSys数据查询 - Query: %s, Token: %s, Operator: %s", query, token, operator)

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

		// 生成模拟数据
		mockData := generateMockCMSysData()

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

// generateMockCMSysData 生成模拟的CMSys主机数据
func generateMockCMSysData() []map[string]interface{} {
	rand.Seed(time.Now().UnixNano())

	// 生成10台模拟主机数据
	var data []map[string]interface{}
	baseIP := []int{192, 168, 1, 0}

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

	for i := 0; i < 10; i++ {
		ipAddress := fmt.Sprintf("%d.%d.%d.%d", baseIP[0], baseIP[1], baseIP[2], baseIP[3]+i+1)

		// 生成随机但合理的监控数据
		cpuMax := 40.0 + rand.Float64()*50.0     // 40-90%
		memMax := 50.0 + rand.Float64()*40.0     // 50-90%
		diskMax := 30.0 + rand.Float64()*60.0    // 30-90%

		hostData := map[string]interface{}{
			"ipAddress":  ipAddress,
			"cpuMaxNew":  fmt.Sprintf("%.2f", cpuMax),
			"memMaxNew":  fmt.Sprintf("%.2f", memMax),
			"diskMaxNew": fmt.Sprintf("%.2f", diskMax),
			"remark":     remarks[i],
		}

		data = append(data, hostData)
	}

	logx.Infof("生成Mock CMSys响应 - 主机数量: %d", len(data))
	return data
}
