package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// ESClient ES数据源客户端
type ESClient struct {
	endpoint   string
	httpClient *http.Client
}

// NewESClient 创建ES客户端
func NewESClient(endpoint string, timeoutSeconds int) *ESClient {
	return &ESClient{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}
}

// ESQueryRequest ES查询请求
type ESQueryRequest struct {
	Index string                 `json:"index"`
	Query map[string]interface{} `json:"query"`
	Aggs  map[string]interface{} `json:"aggs,omitempty"`
	Size  int                    `json:"size,omitempty"`
}

// ESQueryResponse ES查询响应
type ESQueryResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source map[string]interface{} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
}

// Query 执行ES查询
func (c *ESClient) Query(ctx context.Context, req *ESQueryRequest) (*ESQueryResponse, error) {
	// 构建查询请求
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	logx.WithContext(ctx).Infof("ES查询请求: %s", string(reqBody))

	// 发送HTTP请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ES returned status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var esResp ESQueryResponse
	if err := json.Unmarshal(body, &esResp); err != nil {
		return nil, fmt.Errorf("decode response failed: %w", err)
	}

	return &esResp, nil
}

// QueryHostMetrics 查询主机监控指标
func (c *ESClient) QueryHostMetrics(ctx context.Context, indexPattern string, hostIP string, timeRange string) (*HostMetrics, error) {
	// 构建时间范围
	timeRangeMap := map[string]interface{}{
		"gte": "now-" + timeRange,
		"lte": "now",
	}

	// 构建查询条件
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"term": map[string]interface{}{
						"hostIp": hostIP,
					},
				},
				{
					"range": map[string]interface{}{
						"@timestamp": timeRangeMap,
					},
				},
			},
		},
	}

	// 构建聚合查询
	aggs := map[string]interface{}{
		"cpu_stats": map[string]interface{}{
			"stats": map[string]interface{}{
				"field": "cpu",
			},
		},
		"memory_stats": map[string]interface{}{
			"stats": map[string]interface{}{
				"field": "available_memory",
			},
		},
		"disk_stats": map[string]interface{}{
			"stats": map[string]interface{}{
				"field": "total_disk_space_all",
			},
		},
		"hostname": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "hostName.keyword",
				"size":  1,
			},
		},
	}

	req := &ESQueryRequest{
		Index: indexPattern,
		Query: query,
		Aggs:  aggs,
		Size:  0, // 只需要聚合结果
	}

	resp, err := c.Query(ctx, req)
	if err != nil {
		return nil, err
	}

	// 检查是否有数据
	if resp.Hits.Total.Value == 0 {
		return nil, &NoDataError{HostIP: hostIP}
	}

	// 解析聚合结果
	metrics := &HostMetrics{
		HostIP:         hostIP,
		DataPointCount: resp.Hits.Total.Value,
	}

	// 解析主机名
	if hostnameAgg, ok := resp.Aggregations["hostname"].(map[string]interface{}); ok {
		if buckets, ok := hostnameAgg["buckets"].([]interface{}); ok && len(buckets) > 0 {
			if bucket, ok := buckets[0].(map[string]interface{}); ok {
				if key, ok := bucket["key"].(string); ok {
					metrics.HostName = key
				}
			}
		}
	}
	// 如果主机名为空，使用 IP 作为主机名
	if metrics.HostName == "" {
		metrics.HostName = hostIP
	}

	// 解析CPU统计
	if cpuStats, ok := resp.Aggregations["cpu_stats"].(map[string]interface{}); ok {
		if max, ok := cpuStats["max"].(float64); ok {
			metrics.MaxCPU = max
		}
		if avg, ok := cpuStats["avg"].(float64); ok {
			metrics.AvgCPU = avg
		}
	}

	// 解析内存统计
	if memStats, ok := resp.Aggregations["memory_stats"].(map[string]interface{}); ok {
		if max, ok := memStats["max"].(float64); ok {
			metrics.MaxMemory = max
		}
		if avg, ok := memStats["avg"].(float64); ok {
			metrics.AvgMemory = avg
		}
	}

	// 解析磁盘统计
	if diskStats, ok := resp.Aggregations["disk_stats"].(map[string]interface{}); ok {
		if max, ok := diskStats["max"].(float64); ok {
			metrics.MaxDisk = max
		}
		if avg, ok := diskStats["avg"].(float64); ok {
			metrics.AvgDisk = avg
		}
	}

	logx.WithContext(ctx).Infof("主机 %s 指标查询成功: CPU(%.2f/%.2f), 内存(%.2f/%.2f), 磁盘(%.2f/%.2f), 数据点数: %d",
		hostIP, metrics.MaxCPU, metrics.AvgCPU, metrics.MaxMemory, metrics.AvgMemory,
		metrics.MaxDisk, metrics.AvgDisk, metrics.DataPointCount)

	return metrics, nil
}

// HostMetrics 主机指标数据
type HostMetrics struct {
	HostIP         string
	HostName       string
	MaxCPU         float64
	AvgCPU         float64
	MaxMemory      float64
	AvgMemory      float64
	MaxDisk        float64
	AvgDisk        float64
	DataPointCount int
}

// GroupHostMetrics 分组主机指标数据（全量同步用）
type GroupHostMetrics struct {
	HostIP         string
	HostName       string
	Group          string
	MaxCPU         float64
	AvgCPU         float64
	MaxMemory      float64
	AvgMemory      float64
	MaxDisk        float64
	AvgDisk        float64
	DataPointCount int
}

// NoDataError ES查询成功但无数据的错误
type NoDataError struct {
	HostIP string
}

func (e *NoDataError) Error() string {
	return fmt.Sprintf("主机 %s 在ES中没有数据", e.HostIP)
}

// IsNoDataError 判断是否为无数据错误
func IsNoDataError(err error) bool {
	_, ok := err.(*NoDataError)
	return ok
}

// QueryGroupHosts 查询指定group的所有主机数据（用于全量同步）
func (c *ESClient) QueryGroupHosts(ctx context.Context, indexPattern string, groupName string, timeRange string) ([]*GroupHostMetrics, error) {
	// 构建时间范围
	timeRangeMap := map[string]interface{}{
		"gte": "now-" + timeRange,
		"lte": "now",
	}

	// 构建查询条件
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []map[string]interface{}{
				{
					"term": map[string]interface{}{
						"group.keyword": groupName,
					},
				},
				{
					"range": map[string]interface{}{
						"@timestamp": timeRangeMap,
					},
				},
			},
		},
	}

	// 构建聚合查询，按 hostIp 分组
	aggs := map[string]interface{}{
		"hosts": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "hostIp.keyword",
				"size":  10000, // 最多返回10000个主机
			},
			"aggs": map[string]interface{}{
				"hostname": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "hostName.keyword",
						"size":  1,
					},
				},
				"cpu_stats": map[string]interface{}{
					"stats": map[string]interface{}{
						"field": "cpu",
					},
				},
				"memory_stats": map[string]interface{}{
					"stats": map[string]interface{}{
						"field": "available_memory",
					},
				},
				"disk_stats": map[string]interface{}{
					"stats": map[string]interface{}{
						"field": "total_disk_space_all",
					},
				},
			},
		},
	}

	req := &ESQueryRequest{
		Index: indexPattern,
		Query: query,
		Aggs:  aggs,
		Size:  0, // 只需要聚合结果
	}

	resp, err := c.Query(ctx, req)
	if err != nil {
		return nil, err
	}

	// 解析聚合结果
	var result []*GroupHostMetrics

	if hostsAgg, ok := resp.Aggregations["hosts"].(map[string]interface{}); ok {
		if buckets, ok := hostsAgg["buckets"].([]interface{}); ok {
			for _, bucket := range buckets {
				if bucketMap, ok := bucket.(map[string]interface{}); ok {
					hostIP, _ := bucketMap["key"].(string)
					docCount, _ := bucketMap["doc_count"].(float64)

					metrics := &GroupHostMetrics{
						HostIP:         hostIP,
						Group:          groupName,
						DataPointCount: int(docCount),
					}

					// 解析主机名
					if hostnameAgg, ok := bucketMap["hostname"].(map[string]interface{}); ok {
						if hostnameBuckets, ok := hostnameAgg["buckets"].([]interface{}); ok && len(hostnameBuckets) > 0 {
							if hostnameBucket, ok := hostnameBuckets[0].(map[string]interface{}); ok {
								if key, ok := hostnameBucket["key"].(string); ok {
									metrics.HostName = key
								}
							}
						}
					}
					// 如果主机名为空，使用 IP 作为主机名
					if metrics.HostName == "" {
						metrics.HostName = hostIP
					}

					// 解析CPU统计
					if cpuStats, ok := bucketMap["cpu_stats"].(map[string]interface{}); ok {
						if max, ok := cpuStats["max"].(float64); ok {
							metrics.MaxCPU = max
						}
						if avg, ok := cpuStats["avg"].(float64); ok {
							metrics.AvgCPU = avg
						}
					}

					// 解析内存统计
					if memStats, ok := bucketMap["memory_stats"].(map[string]interface{}); ok {
						if max, ok := memStats["max"].(float64); ok {
							metrics.MaxMemory = max
						}
						if avg, ok := memStats["avg"].(float64); ok {
							metrics.AvgMemory = avg
						}
					}

					// 解析磁盘统计
					if diskStats, ok := bucketMap["disk_stats"].(map[string]interface{}); ok {
						if max, ok := diskStats["max"].(float64); ok {
							metrics.MaxDisk = max
						}
						if avg, ok := diskStats["avg"].(float64); ok {
							metrics.AvgDisk = avg
						}
					}

					result = append(result, metrics)
				}
			}
		}
	}

	logx.WithContext(ctx).Infof("查询group=%s的主机数据成功，共 %d 台主机", groupName, len(result))

	// 调试日志：打印前3台主机的详细数据
	for i, host := range result {
		if i >= 3 {
			break
		}
		logx.WithContext(ctx).Infof("主机[%d] IP=%s, Name=%s, MaxCPU=%.2f, MaxMem=%.2f, MaxDisk=%.2f, DataPoints=%d",
			i, host.HostIP, host.HostName, host.MaxCPU, host.MaxMemory, host.MaxDisk, host.DataPointCount)
	}

	return result, nil
}
