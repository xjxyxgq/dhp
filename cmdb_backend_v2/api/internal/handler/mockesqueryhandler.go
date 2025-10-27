package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// MockEsQueryHandler 模拟ES查询接口，用于开发测试
func MockEsQueryHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Info("收到Mock ES查询请求")

		// 解析请求体
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			httpx.Error(w, err)
			return
		}

		// 提取查询参数
		index := ""
		hostIP := ""
		groupName := ""
		if indexVal, ok := reqBody["index"].(string); ok {
			index = indexVal
		}
		if query, ok := reqBody["query"].(map[string]interface{}); ok {
			if boolQuery, ok := query["bool"].(map[string]interface{}); ok {
				if must, ok := boolQuery["must"].([]interface{}); ok {
					for _, condition := range must {
						if condMap, ok := condition.(map[string]interface{}); ok {
							if term, ok := condMap["term"].(map[string]interface{}); ok {
								// 提取 hostIP
								if ip, ok := term["hostIp"].(string); ok {
									hostIP = ip
								}
								// 提取 group.keyword
								if group, ok := term["group.keyword"].(string); ok {
									groupName = group
								}
								// 也尝试提取 group（兼容不带.keyword的情况）
								if groupName == "" {
									if group, ok := term["group"].(string); ok {
										groupName = group
									}
								}
							}
						}
					}
				}
			}
		}

		// 判断是单主机查询还是group查询
		var mockResponse map[string]interface{}
		if groupName != "" {
			// group查询
			logx.Infof("Mock ES group查询 - Index: %s, Group: %s", index, groupName)
			mockResponse = generateGroupMockESResponse(groupName, 100)
		} else {
			// 单主机查询
			logx.Infof("Mock ES单主机查询 - Index: %s, HostIP: %s", index, hostIP)
			mockResponse = generateSingleHostMockESResponse(hostIP)
		}

		// 返回响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}
}

// generateSingleHostMockESResponse 生成单主机查询的模拟ES响应
func generateSingleHostMockESResponse(hostIP string) map[string]interface{} {
	rand.Seed(time.Now().UnixNano())

	// 生成随机但合理的监控数据
	maxCPU := 60.0 + rand.Float64()*30.0          // 60-90%
	avgCPU := maxCPU * (0.6 + rand.Float64()*0.2) // 60-80% of max
	maxMemory := 50.0 + rand.Float64()*40.0       // 50-90 GB
	avgMemory := maxMemory * (0.7 + rand.Float64()*0.2)
	maxDisk := 800.0 + rand.Float64()*200.0 // 800-1000 GB
	avgDisk := maxDisk * (0.8 + rand.Float64()*0.15)

	// 数据点数量（假设每5分钟一个点，30天 = 8640个点）
	dataPoints := 7000 + rand.Intn(1640)

	response := map[string]interface{}{
		"took":      15,
		"timed_out": false,
		"_shards": map[string]interface{}{
			"total":      5,
			"successful": 5,
			"skipped":    0,
			"failed":     0,
		},
		"hits": map[string]interface{}{
			"total": map[string]interface{}{
				"value":    dataPoints,
				"relation": "eq",
			},
			"max_score": nil,
			"hits":      []interface{}{},
		},
		"aggregations": map[string]interface{}{
			"cpu_stats": map[string]interface{}{
				"count": dataPoints,
				"min":   10.5,
				"max":   maxCPU,
				"avg":   avgCPU,
				"sum":   avgCPU * float64(dataPoints),
			},
			"memory_stats": map[string]interface{}{
				"count": dataPoints,
				"min":   20.0,
				"max":   maxMemory,
				"avg":   avgMemory,
				"sum":   avgMemory * float64(dataPoints),
			},
			"disk_stats": map[string]interface{}{
				"count": dataPoints,
				"min":   500.0,
				"max":   maxDisk,
				"avg":   avgDisk,
				"sum":   avgDisk * float64(dataPoints),
			},
		},
	}

	logx.Infof("生成Mock ES单主机响应 - HostIP: %s, DataPoints: %d, MaxCPU: %.2f, MaxMemory: %.2f, MaxDisk: %.2f",
		hostIP, dataPoints, maxCPU, maxMemory, maxDisk)

	return response
}

// generateGroupMockESResponse 生成group查询的模拟ES响应（聚合多个主机）
func generateGroupMockESResponse(groupName string, hostCount int) map[string]interface{} {
	rand.Seed(time.Now().UnixNano())

	logx.Infof("Mock ES group查询 - Group: %s, 生成 %d 台主机数据", groupName, hostCount)

	// 创建主机buckets数组
	buckets := make([]map[string]interface{}, hostCount)
	totalDocCount := 0

	for i := 0; i < hostCount; i++ {
		// 生成主机IP：10.0.1.1 到 10.0.1.100
		hostIP := fmt.Sprintf("10.0.1.%d", i+1)
		// 生成主机名：db-server-001 到 db-server-100
		hostName := fmt.Sprintf("db-server-%03d", i+1)

		// 数据点数量（7000-8640之间）
		docCount := 7000 + rand.Intn(1640)
		totalDocCount += docCount

		// 生成随机但合理的监控数据
		// CPU: 30-95%
		minCPU := 15.0 + rand.Float64()*20.0   // 15-35%
		maxCPU := 60.0 + rand.Float64()*35.0   // 60-95%
		avgCPU := minCPU + (maxCPU-minCPU)*0.6 // 介于min和max之间
		sumCPU := avgCPU * float64(docCount)

		// 内存: 40-95 GB
		minMemory := 30.0 + rand.Float64()*20.0   // 30-50 GB
		maxMemory := 60.0 + rand.Float64()*35.0   // 60-95 GB
		avgMemory := minMemory + (maxMemory-minMemory)*0.65
		sumMemory := avgMemory * float64(docCount)

		// 磁盘: 700-2000 GB
		minDisk := 500.0 + rand.Float64()*300.0   // 500-800 GB
		maxDisk := 900.0 + rand.Float64()*1100.0  // 900-2000 GB
		avgDisk := minDisk + (maxDisk-minDisk)*0.7
		sumDisk := avgDisk * float64(docCount)

		bucket := map[string]interface{}{
			"key":       hostIP,
			"doc_count": docCount,
			"hostname": map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets": []map[string]interface{}{
					{
						"key":       hostName,
						"doc_count": docCount,
					},
				},
			},
			"cpu_stats": map[string]interface{}{
				"count": docCount,
				"min":   minCPU,
				"max":   maxCPU,
				"avg":   avgCPU,
				"sum":   sumCPU,
			},
			"memory_stats": map[string]interface{}{
				"count": docCount,
				"min":   minMemory,
				"max":   maxMemory,
				"avg":   avgMemory,
				"sum":   sumMemory,
			},
			"disk_stats": map[string]interface{}{
				"count": docCount,
				"min":   minDisk,
				"max":   maxDisk,
				"avg":   avgDisk,
				"sum":   sumDisk,
			},
		}

		buckets[i] = bucket
	}

	response := map[string]interface{}{
		"took":      25,
		"timed_out": false,
		"_shards": map[string]interface{}{
			"total":      5,
			"successful": 5,
			"skipped":    0,
			"failed":     0,
		},
		"hits": map[string]interface{}{
			"total": map[string]interface{}{
				"value":    totalDocCount,
				"relation": "eq",
			},
			"max_score": nil,
			"hits":      []interface{}{},
		},
		"aggregations": map[string]interface{}{
			"hosts": map[string]interface{}{
				"doc_count_error_upper_bound": 0,
				"sum_other_doc_count":         0,
				"buckets":                     buckets,
			},
		},
	}

	logx.Infof("生成Mock ES group响应 - Group: %s, 主机数: %d, 总数据点数: %d",
		groupName, hostCount, totalDocCount)

	return response
}
