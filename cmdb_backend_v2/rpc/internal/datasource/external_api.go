package datasource

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/config"
	"cmdb-rpc/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ExternalAPI struct {
	conn          sqlx.SqlConn
	hostPoolModel model.HostsPoolModel
	config        *config.ExternalAPIConfig
}

func NewExternalAPI(conn sqlx.SqlConn, apiConfig *config.ExternalAPIConfig) *ExternalAPI {
	return &ExternalAPI{
		conn:          conn,
		hostPoolModel: model.NewHostsPoolModel(conn),
		config:        apiConfig,
	}
}

// SyncHostsPoolData 同步hosts_pool数据
func (e *ExternalAPI) SyncHostsPoolData() error {
	logx.Info("开始同步hosts_pool数据")

	// 模拟调用外部接口获取数据
	hosts, err := e.fetchHostsFromExternalAPI()
	if err != nil {
		return fmt.Errorf("获取外部接口数据失败: %v", err)
	}

	// 将数据插入数据库
	for _, host := range hosts {
		_, err = e.hostPoolModel.Insert(context.Background(), host)
		if err != nil {
			logx.Errorf("插入主机数据失败: %v", err)
			continue
		}
		logx.Infof("成功插入主机数据: %s (%s)", host.HostName, host.HostIp)
	}

	logx.Infof("hosts_pool数据同步完成，共处理%d条记录", len(hosts))
	return nil
}

// fetchHostsFromExternalAPI 模拟从外部接口获取主机数据
func (e *ExternalAPI) fetchHostsFromExternalAPI() ([]*model.HostsPool, error) {
	// 这里模拟一个外部接口调用
	// 实际实现时，这里应该是真实的HTTP请求
	logx.Info("模拟调用外部接口获取主机数据")

	// 模拟HTTP请求参数
	requestBody := map[string]interface{}{
		"page":     1,
		"pageSize": 100,
		"filters":  map[string]interface{}{},
	}

	// 模拟请求过程
	_, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构造请求参数失败: %v", err)
	}

	// 模拟HTTP请求（这里返回mock数据）
	hosts := e.generateMockHosts(100)

	logx.Infof("外部接口返回%d条主机数据", len(hosts))
	return hosts, nil
}

// generateMockHosts 生成模拟主机数据
func (e *ExternalAPI) generateMockHosts(count int) []*model.HostsPool {
	var hosts []*model.HostsPool

	hostTypes := []string{"云主机", "裸金属"}

	for i := 0; i < count; i++ {
		host := &model.HostsPool{
			HostName: fmt.Sprintf("server-%03d", i+1),
			HostIp:   fmt.Sprintf("192.168.%d.%d", 1+(i/254), (i%254)+1),
			HostType: sql.NullString{
				String: hostTypes[rand.Intn(len(hostTypes))],
				Valid:  true,
			},
			H3cId: sql.NullString{
				String: fmt.Sprintf("H3C-%06d", i+1),
				Valid:  true,
			},
			H3cStatus: sql.NullString{
				String: "运行中",
				Valid:  true,
			},
			DiskSize: sql.NullInt64{
				Int64: int64(rand.Intn(1000) + 100), // 100-1100 GB
				Valid: true,
			},
			Ram: sql.NullInt64{
				Int64: int64(rand.Intn(64) + 8), // 8-72 GB
				Valid: true,
			},
			Vcpus: sql.NullInt64{
				Int64: int64(rand.Intn(16) + 2), // 2-18 cores
				Valid: true,
			},
			IfH3cSync: sql.NullString{
				String: "是",
				Valid:  true,
			},
			H3cImgId: sql.NullString{
				String: fmt.Sprintf("IMG-%06d", i+1),
				Valid:  true,
			},
			H3cHmName: sql.NullString{
				String: fmt.Sprintf("HM-%06d", i+1),
				Valid:  true,
			},
			LeafNumber: sql.NullString{
				String: fmt.Sprintf("LEAF-%02d", rand.Intn(10)+1),
				Valid:  true,
			},
			RackNumber: sql.NullString{
				String: fmt.Sprintf("RACK-%02d", rand.Intn(20)+1),
				Valid:  true,
			},
			RackHeight: sql.NullInt64{
				Int64: int64(rand.Intn(42) + 1), // 1-42U
				Valid: true,
			},
			RackStartNumber: sql.NullInt64{
				Int64: int64(rand.Intn(40) + 1), // 1-40U
				Valid: true,
			},
			FromFactor: sql.NullInt64{
				Int64: int64(rand.Intn(4) + 1), // 1-4U
				Valid: true,
			},
			SerialNumber: sql.NullString{
				String: fmt.Sprintf("SN%d%06d", time.Now().Year(), i+1),
				Valid:  true,
			},
			IsDeleted: 0,
			IsStatic:  0,
		}

		hosts = append(hosts, host)
	}

	return hosts
}

// RealExternalAPICall 真实的外部接口调用示例（未实现）
func (e *ExternalAPI) RealExternalAPICall(url string, requestBody interface{}) (*http.Response, error) {
	// 这里是真实外部接口调用的框架
	// 实际使用时需要根据具体的API文档来实现
	logx.Infof("调用外部接口: %s", url)

	// 构造HTTP请求
	// client := &http.Client{Timeout: 30 * time.Second}
	// jsonBody, _ := json.Marshal(requestBody)
	// req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	// if err != nil {
	// 	return nil, err
	// }
	// req.Header.Set("Content-Type", "application/json")
	//
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// return resp, nil

	// 目前返回nil，表示未实现
	return nil, fmt.Errorf("外部接口调用尚未实现")
}

// 外部CMDB接口相关结构体定义
type ExternalCMDBRequest struct {
	PageBeanDTO ExternalPageBeanDTO `json:"pageBeanDTO"`
	ReqDTO      ExternalReqDTO      `json:"reqDTO"`
}

type ExternalPageBeanDTO struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
}

type ExternalReqDTO struct {
	RpcLogId  string `json:"rpcLogId"`
	HostOwner int    `json:"host_owner,omitempty"`
}

type ExternalCMDBResponse struct {
	Success bool                      `json:"success"`
	Result  ExternalCMDBResponseData  `json:"result"`
}

type ExternalCMDBResponseData struct {
	PageSize    int                     `json:"pageSize"`
	Success     bool                    `json:"success"`
	TotalPage   int                     `json:"totalPage"`
	CurrentPage int                     `json:"currentPage"`
	TotalRows   int                     `json:"totalRows"`
	Result      []ExternalCMDBHostData  `json:"result"`
}

type ExternalCMDBHostData struct {
	CmdbId        string                 `json:"cmdbId"`
	DomainNum     string                 `json:"domainNum"`
	HostName      string                 `json:"hostName"`
	HostIp        string                 `json:"hostIp"`
	HostType      string                 `json:"hostType"`
	HostOwner     string                 `json:"hostOwner"`
	OpsIamCode    string                 `json:"opsIamCode"`
	OwnerGroup    string                 `json:"ownerGroup"`
	OwnerIamCode  string                 `json:"ownerIamCode"`
	H3cId         string                 `json:"h3cId"`
	H3cStatus     string                 `json:"h3cStatus"`
	CreatedAt     string                 `json:"createdAt"`
	UpdatedAt     string                 `json:"updatedAt"`
	HostExtInfo   ExternalHostExtInfo    `json:"hostExtInfo"`
	IfH3cSync     string                 `json:"ifH3cSync"`
	H3cImageId    string                 `json:"h3cImageId"`
	H3cHmName     string                 `json:"h3cHmName"`
	IsDelete      string                 `json:"isDelete"`
	DeployAppInfo ExternalDeployAppInfo  `json:"deployAppInfo"`
	DeployAppInfoList []ExternalDeployAppInfo  `json:"deployAppInfoList"`
	HostLocInfo   map[string]interface{} `json:"hostLocInfo"`
}

type ExternalHostExtInfo struct {
	H3cId  string `json:"h3cId"`
	Disk   int64  `json:"disk"`
	Ram    int64  `json:"ram"`
	Vcpus  int64  `json:"vcpus"`
}

type ExternalDeployAppInfo struct {
	AppName      string `json:"appName"`
	DataSource   string `json:"dataSource"`
	BizGroup     string `json:"bizGroup"`
	OpsBizGroup  string `json:"opsBizGroup"`
}

// FetchHostsFromExternalCMDB 从外部CMDB获取主机数据
func (e *ExternalAPI) FetchHostsFromExternalCMDB(pageSize int, hostOwner int) ([]*cmpool.ExternalCmdbHost, int, error) {
	logx.Info("开始从外部CMDB获取主机数据")

	// 外部CMDB接口配置 - 从配置文件读取
	apiURL := e.config.CmdbUrl + "/v1/host/page"
	appToken := e.config.CmdbAppCode
	accessToken := e.config.CmdbSecret

	// 构造请求体
	reqBody := ExternalCMDBRequest{
		PageBeanDTO: ExternalPageBeanDTO{
			CurrentPage: 1,
			PageSize:    pageSize,
		},
		ReqDTO: ExternalReqDTO{
			RpcLogId: accessToken,
		},
	}

	// 如果指定了HostOwner，则添加筛选条件
	if hostOwner > 0 {
		reqBody.ReqDTO.HostOwner = hostOwner
	}

	// 序列化请求体
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("序列化请求体失败: %v", err)
	}

	// 创建HTTP客户端
	client := &http.Client{Timeout: 30 * time.Second}

	// 创建HTTP请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, 0, fmt.Errorf("创建HTTP请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("X-CONTROL-ACCESS-APP", appToken)
	req.Header.Set("X-CONTROL-ACCESS-TOKEN", accessToken)
	req.Header.Set("Content-Type", "application/json")

	logx.Infof("调用外部CMDB接口: %s", apiURL)

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("调用外部CMDB接口失败: %v", err)
	}
	defer resp.Body.Close()

	// 检查响应状态码
	if resp.StatusCode != 200 {
		return nil, 0, fmt.Errorf("外部CMDB接口返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var cmdbResp ExternalCMDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&cmdbResp); err != nil {
		return nil, 0, fmt.Errorf("解析外部CMDB响应失败: %v", err)
	}

	// 检查接口返回是否成功
	if !cmdbResp.Success || !cmdbResp.Result.Success {
		return nil, 0, fmt.Errorf("外部CMDB接口返回失败")
	}

	// 转换数据格式
	var hosts []*cmpool.ExternalCmdbHost
	for _, hostData := range cmdbResp.Result.Result {
		host := &cmpool.ExternalCmdbHost{
			CmdbId:      hostData.CmdbId,
			DomainNum:   hostData.DomainNum,
			HostName:    hostData.HostName,
			HostIp:      hostData.HostIp,
			HostType:    hostData.HostType,
			HostOwner:   hostData.HostOwner,
			OpsIamCode:  hostData.OpsIamCode,
			OwnerGroup:  hostData.OwnerGroup,
			OwnerIamCode: hostData.OwnerIamCode,
			H3CId:       hostData.H3cId,
			H3CStatus:   hostData.H3cStatus,
			Disk:        hostData.HostExtInfo.Disk,
			Ram:         hostData.HostExtInfo.Ram,
			Vcpus:       hostData.HostExtInfo.Vcpus,
			CreatedAt:   hostData.CreatedAt,
			UpdatedAt:   hostData.UpdatedAt,
			IfH3CSync:   hostData.IfH3cSync,
			H3CImageId:  hostData.H3cImageId,
			H3CHmName:   hostData.H3cHmName,
			IsDelete:    hostData.IsDelete,
		}

		// 处理应用部署信息
		if hostData.DeployAppInfo.AppName != "" {
			host.AppName = hostData.DeployAppInfo.AppName
			host.DataSource = hostData.DeployAppInfo.DataSource
			host.BizGroup = hostData.DeployAppInfo.BizGroup
			host.OpsBizGroup = hostData.DeployAppInfo.OpsBizGroup
		}

		hosts = append(hosts, host)
	}

	logx.Infof("从外部CMDB获取到 %d 条主机数据，总页数: %d", len(hosts), cmdbResp.Result.TotalPage)

	return hosts, cmdbResp.Result.TotalPage, nil
}
