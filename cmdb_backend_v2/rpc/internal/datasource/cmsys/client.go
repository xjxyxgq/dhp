package cmsys

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// CMSysClient CMSys 数据源客户端
type CMSysClient struct {
	authEndpoint string // 认证接口地址
	dataEndpoint string // 数据接口地址
	appCode      string // 应用代码
	appSecret    string // 应用密钥
	operator     string // 操作员标识
	httpClient   *http.Client
	logger       logx.Logger
	token        string    // 当前 token
	tokenExpiry  time.Time // token 过期时间
}

// AuthRequest 认证请求
type AuthRequest struct {
	AppCode string `json:"appCode"`
	Secret  string `json:"secret"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"` // token
}

// DataResponse 数据响应
type DataResponse struct {
	Code string     `json:"code"`
	Msg  string     `json:"msg"`
	Data []HostData `json:"data"`
}

// HostData 主机数据
type HostData struct {
	IPAddress  string `json:"ipAddress"`
	CPUMaxNew  string `json:"cpuMaxNew"`
	MemMaxNew  string `json:"memMaxNew"`
	DiskMaxNew string `json:"diskMaxNew"`
	Remark     string `json:"remark"`
}

// HostMetrics 主机指标数据
type HostMetrics struct {
	IPAddress         string
	HostName          string  // 主机名（如果为空则使用 IP）
	CPUUsedPercent    float64 // CPU使用率百分比
	MemoryUsedPercent float64 // 内存使用率百分比
	DiskUsedPercent   float64 // 磁盘使用率百分比
	MaxCPU            float64 // 兼容旧字段
	MaxMemory         float64 // 兼容旧字段
	MaxDisk           float64 // 兼容旧字段
	Remark            string
}

// NewCMSysClient 创建 CMSys 客户端
func NewCMSysClient(authEndpoint, dataEndpoint, appCode, appSecret, operator string, timeoutSeconds int) *CMSysClient {
	// 创建不验证证书的 HTTP 客户端（生产环境应该验证证书）
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &CMSysClient{
		authEndpoint: authEndpoint,
		dataEndpoint: dataEndpoint,
		appCode:      appCode,
		appSecret:    appSecret,
		operator:     operator,
		httpClient: &http.Client{
			Timeout:   time.Duration(timeoutSeconds) * time.Second,
			Transport: tr,
		},
		logger: logx.WithContext(context.Background()),
	}
}

// getToken 获取认证 token
func (c *CMSysClient) getToken(ctx context.Context) (string, error) {
	// 如果 token 仍然有效，直接返回
	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return c.token, nil
	}

	// 构建认证请求
	authReq := AuthRequest{
		AppCode: c.appCode,
		Secret:  c.appSecret,
	}

	reqBody, err := json.Marshal(authReq)
	if err != nil {
		return "", fmt.Errorf("序列化认证请求失败: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", c.authEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("创建认证请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("发送认证请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取认证响应失败: %v", err)
	}

	// 解析响应
	var authResp AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return "", fmt.Errorf("解析认证响应失败: %v", err)
	}

	// 检查响应状态
	if authResp.Code != "A0000" {
		return "", fmt.Errorf("认证失败: code=%s, msg=%s", authResp.Code, authResp.Msg)
	}

	// 保存 token，设置 1 小时后过期
	c.token = authResp.Data
	c.tokenExpiry = time.Now().Add(1 * time.Hour)

	c.logger.Infof("成功获取 CMSys token")
	return c.token, nil
}

// QueryHostMetrics 查询主机指标数据
func (c *CMSysClient) QueryHostMetrics(ctx context.Context, query string) ([]*HostMetrics, error) {
	// 获取 token
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 token 失败: %v", err)
	}

	// 构建请求 URL
	url := c.dataEndpoint
	if query != "" {
		url = fmt.Sprintf("%s?query=%s", url, query)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建数据请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("x-control-access-token", token)
	req.Header.Set("x-control-access-operator", c.operator)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送数据请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取数据响应失败: %v", err)
	}

	// 解析响应
	var dataResp DataResponse
	if err := json.Unmarshal(body, &dataResp); err != nil {
		return nil, fmt.Errorf("解析数据响应失败: %v", err)
	}

	// 检查响应状态
	if dataResp.Code != "A000" {
		return nil, fmt.Errorf("查询数据失败: code=%s, msg=%s", dataResp.Code, dataResp.Msg)
	}

	// 转换数据格式
	metrics := make([]*HostMetrics, 0, len(dataResp.Data))
	for _, host := range dataResp.Data {
		// 解析数值
		maxCPU, _ := strconv.ParseFloat(host.CPUMaxNew, 64)
		maxMemory, _ := strconv.ParseFloat(host.MemMaxNew, 64)
		maxDisk, _ := strconv.ParseFloat(host.DiskMaxNew, 64)

		metrics = append(metrics, &HostMetrics{
			IPAddress:         host.IPAddress,
			HostName:          host.IPAddress, // CMSys 接口不返回 hostName，使用 IP 作为主机名
			CPUUsedPercent:    maxCPU,
			MemoryUsedPercent: maxMemory,
			DiskUsedPercent:   maxDisk,
			MaxCPU:            maxCPU,    // 兼容旧字段
			MaxMemory:         maxMemory, // 兼容旧字段
			MaxDisk:           maxDisk,   // 兼容旧字段
			Remark:            host.Remark,
		})
	}

	c.logger.Infof("成功获取 %d 条主机数据", len(metrics))
	return metrics, nil
}

// QueryHostMetricsByIP 按 IP 查询单个主机的指标数据
func (c *CMSysClient) QueryHostMetricsByIP(ctx context.Context, ip string) (*HostMetrics, error) {
	// 构建查询参数 - 按 IP 查询
	query := fmt.Sprintf("ipAddress=%s", ip)

	// 调用通用查询方法
	metricsList, err := c.QueryHostMetrics(ctx, query)
	if err != nil {
		return nil, err
	}

	// 如果没有数据
	if len(metricsList) == 0 {
		return nil, &NoDataError{Query: query}
	}

	// 返回第一条数据
	return metricsList[0], nil
}

// DataByIPsRequest 按IP列表查询的请求
type DataByIPsRequest struct {
	IPs []string `json:"ips"`
}

// QueryHostMetricsByIPs 按 IP 列表批量查询主机指标数据
// 使用 POST /platform/cmsys/data-by-ips 接口
func (c *CMSysClient) QueryHostMetricsByIPs(ctx context.Context, ips []string) ([]*HostMetrics, error) {
	if len(ips) == 0 {
		return []*HostMetrics{}, nil
	}

	// 获取 token
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取 token 失败: %v", err)
	}

	// 构建请求 URL (假设 dataEndpoint 是 /platform/cmsys/data，我们需要改为 /platform/cmsys/data-by-ips)
	// 提取 base URL
	baseURL := strings.TrimSuffix(c.dataEndpoint, "/data")
	url := baseURL + "/data-by-ips"

	// 构建请求体
	reqBody := DataByIPsRequest{IPs: ips}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("创建数据请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-control-access-token", token)
	req.Header.Set("x-control-access-operator", c.operator)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送数据请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取数据响应失败: %v", err)
	}

	// 解析响应
	var dataResp DataResponse
	if err := json.Unmarshal(body, &dataResp); err != nil {
		return nil, fmt.Errorf("解析数据响应失败: %v", err)
	}

	// 检查响应状态
	if dataResp.Code != "A000" {
		return nil, fmt.Errorf("查询数据失败: code=%s, msg=%s", dataResp.Code, dataResp.Msg)
	}

	// 转换数据格式
	metrics := make([]*HostMetrics, 0, len(dataResp.Data))
	for _, host := range dataResp.Data {
		// 解析数值
		maxCPU, _ := strconv.ParseFloat(host.CPUMaxNew, 64)
		maxMemory, _ := strconv.ParseFloat(host.MemMaxNew, 64)
		maxDisk, _ := strconv.ParseFloat(host.DiskMaxNew, 64)

		metrics = append(metrics, &HostMetrics{
			IPAddress:         host.IPAddress,
			HostName:          host.IPAddress, // CMSys 接口不返回 hostName，使用 IP 作为主机名
			CPUUsedPercent:    maxCPU,
			MemoryUsedPercent: maxMemory,
			DiskUsedPercent:   maxDisk,
			MaxCPU:            maxCPU,    // 兼容旧字段
			MaxMemory:         maxMemory, // 兼容旧字段
			MaxDisk:           maxDisk,   // 兼容旧字段
			Remark:            host.Remark,
		})
	}

	c.logger.Infof("按IP列表成功获取 %d 条主机数据（请求了 %d 个IP）", len(metrics), len(ips))
	return metrics, nil
}

// NoDataError CMSys 查询成功但无数据的错误
type NoDataError struct {
	Query string
}

func (e *NoDataError) Error() string {
	return fmt.Sprintf("CMSys 数据源中没有符合条件的数据: query=%s", e.Query)
}

// IsNoDataError 判断是否为无数据错误
func IsNoDataError(err error) bool {
	_, ok := err.(*NoDataError)
	return ok
}
