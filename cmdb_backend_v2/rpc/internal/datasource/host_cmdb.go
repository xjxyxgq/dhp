package datasource

import (
	"encoding/json"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"cmdb-rpc/internal/common"
	"time"
)

type Cmdb struct {
	Url                     string                           `json:"url"`
	AppCode                 string                           `json:"appCode"` // cmdb api 要求驼峰变量命名
	Secret                  string                           `json:"secret"`
	Token                   string                           `json:"token"`
	TokenTimestamp          time.Time                        `json:"token_timestamp"`            // token 目前是6小时过期，超过这个时间则需要刷新
	ReadableBizGroupMapping map[string]*BizGroupRelationship `json:"readable_biz_group_mapping"` // 用于实现CloudDB中的中文名称到CMDB系统中的BizGroup编码的对应关系
}

type BizGroupRelationship struct {
	InCloudDB string `json:"in_cloud_db"`
	InIAM     string `json:"in_iam"`
	InCMDB    string `json:"in_cmdb"`
	IAMCode   string `json:"iam_code"`
	CMDBCode  string `json:"cmdb_code"`
}

type TokensReq struct {
	Appcode string `json:"appcode"`
	Secret  string `json:"secret"`
}

type TokensResp struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

type HostData struct {
	DomainNum         string       `json:"domainNum"`
	HostName          string       `json:"hostName"`
	HostIp            string       `json:"hostIp"`
	HostType          string       `json:"hostType"`
	HostOwner         string       `json:"hostOwner"`
	H3CId             string       `json:"h3cId,omitempty"`
	H3CStatus         string       `json:"h3cStatus,omitempty"`
	HostExtInfo       *HostExtInfo `json:"hostExtInfo"`
	IfH3CSync         string       `json:"ifH3cSync,omitempty"`
	H3CImageId        string       `json:"h3cImageId,omitempty"`
	H3CHmName         string       `json:"h3cHmName,omitempty"`
	IsDelete          string       `json:"isDelete"`
	DeployAppInfoList []*AppInfo   `json:"deployAppInfoList"`
	HostLocInfo       *HostLoc     `json:"hostLocInfo"`
	BusinessNum       string       `json:"businessNum"`
}

type AppData struct {
	HostIp       string `json:"host_ip"`
	AppName      string `json:"app_name"`
	BizGroup     string `json:"bizGroup"`
	OpsBizGroup  string `json:"opsBizGroup"`
	Owner        string `json:"owner"`
	Operator     string `json:"operator"`
	SoftwareName string `json:"softwareName"`
	UpdateBy     string `json:"updateBy"`
}

type HostExtInfo struct {
	Disk  int64 `json:"disk"`
	Ram   int64 `json:"ram"`
	Vcpus int32 `json:"vcpus"`
}

type AppInfo struct {
	AppName    string `json:"appName"`
	DataSource string `json:"dataSource"`
}

type HostLoc struct {
	LeafNumber      string `json:"leafNumber"`
	RackNumber      string `json:"rackNumber"`
	RackHeight      int32  `json:"rackHeight"`
	RackStartNumber int32  `json:"rackStartNumber"`
	FromFactor      int32  `json:"formFactor"`
	SerialNumber    string `json:"serialNumber"`
}

type HostInjectRequest struct {
	RpcLogId         string   `json:"rpcLogId"`
	AppName          string   `json:"appName"`
	HostIpList       []string `json:"hostIpList"`
	BizGroup         *string  `json:"bizGroup,omitempty"`
	BizGroupReadable *string  `json:"bizGroupReadable,omitempty"`
	OpsBizGroup      *string  `json:"opsBizGroup,omitempty"`
	Owner            *string  `json:"owner,omitempty"`
	Operator         *string  `json:"operator,omitempty"`
	SoftwareName     string   `json:"softwareName,omitempty"`
	SoftwareVersion  *string  `json:"softwareVersion,omitempty"`
	UpdateBy         string   `json:"update_by,omitempty"`
}

type HostInjectResponse struct {
	Success bool `json:"success"`
	Result  bool `json:"result"`
}

type HostsRequest struct {
	PageBeanDTO *ReqPageBeanDTO `json:"pageBeanDTO"`
	ReqDTO      *ReqDTO         `json:"reqDTO"`
}

type ReqPageBeanDTO struct {
	CurrentPage int `json:"currentPage"`
	PageSize    int `json:"pageSize"`
}

type ReqDTO struct {
	RpcLogId   string   `json:"rpcLogId"`
	HostIpList []string `json:"hostIpList"`
}

type HostsResponse struct {
	Success bool `json:"success"`
	Result  struct {
		PageSize    int         `json:"pageSize"`
		Success     bool        `json:"success"`
		TotalPage   int         `json:"totalPage"`
		CurrentPage int         `json:"currentPage"`
		TotalRows   int         `json:"totalRows"`
		Result      []*HostData `json:"result"`
	} `json:"result"`
}

func NewCmdb(url, appcode, secret string) *Cmdb {
	return &Cmdb{
		Url:                     url,
		AppCode:                 appcode,
		Secret:                  secret,
		ReadableBizGroupMapping: make(map[string]*BizGroupRelationship),
	}
}

// LoadBizGroupMapping
//
//	@Description: 构建 ReadableBizGroupMapping 属性的值，形成业务名称在各个平台中的名称和编码的对应关系
//	@receiver c
//	@param data
//func (c *Cmdb) LoadBizGroupMapping(data []*model.HostsPoolBizRs) {
//	for _, bizGrpRelationship := range data {
//		if bizGrpRelationship.InClouddb.String != "" {
//			// clouddb中业务组名称为空时，即clouddb中没有属于这个业务组的机器
//			c.ReadableBizGroupMapping[bizGrpRelationship.InClouddb.String] = &BizGroupRelationship{
//				InCloudDB: bizGrpRelationship.InClouddb.String,
//				InCMDB:    bizGrpRelationship.InCmdb.String,
//				InIAM:     bizGrpRelationship.InIam.String,
//				CMDBCode:  bizGrpRelationship.CmdbCode.String,
//				IAMCode:   bizGrpRelationship.IamCode.String,
//			}
//		}
//	}
//}

// RefreshToken 公开的刷新Token方法
func (c *Cmdb) RefreshToken() error {
	return c.refreshToken()
}

// refreshToken
//
//	@Description: 刷新CMDB访问token
//	@receiver c
//	@return error
func (c *Cmdb) refreshToken() error {
	url := fmt.Sprintf("%s/v2/auth/tokens", c.Url)
	reqBody, err := json.Marshal(&TokensReq{Appcode: c.AppCode, Secret: c.Secret})
	if err != nil {
		return err
	}
	logx.Infof("request cmdb %s %s", url, string(reqBody))

	respBody, err := common.HttpPostReq(url, reqBody, nil)
	if err != nil {
		return err
	}
	tokenResp := new(TokensResp)
	err = json.Unmarshal(respBody, &tokenResp)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf(`err msg: %s, original: %s`, err.Error(), respBody))
	}
	if tokenResp.Code != "A0000" {
		return fmt.Errorf(fmt.Sprintf(tokenResp.Msg))
	}
	c.Token = tokenResp.Data
	c.TokenTimestamp = time.Now().Add(-time.Minute)
	return nil
}

// GetDataByHostIpList
//
//	@Description: 请求 CMDB 的接口，根据提供的主机列表，获取这些主机在CMDB中的数据
//	@receiver c
//	@param ipList
//	@param batchMaxSize
//	@param isFreak 是否使用模拟数据
//	@return []*HostData
//	@return error
func (c *Cmdb) GetDataByHostIpList(ipList []string, batchMaxSize int, isFreak bool) ([]*HostData, error) {
	// 如果是模拟模式，直接返回模拟数据
	if isFreak {
		return c.generateMockHostData(ipList), nil
	}

	url := fmt.Sprintf("%s/v1/host/page", c.Url)
	if len(ipList) > batchMaxSize {
		logx.Errorf("单次请求CMDB获取主机数量大于%d", batchMaxSize)
		return nil, fmt.Errorf("单次请求CMDB获取主机数量大于%d", batchMaxSize)
	}
	reqBody, err := json.Marshal(&HostsRequest{
		PageBeanDTO: &ReqPageBeanDTO{CurrentPage: 1, PageSize: len(ipList)},
		ReqDTO:      &ReqDTO{RpcLogId: c.Token, HostIpList: ipList},
	})
	if err != nil {
		return nil, err
	}
	logx.Infof("请求CMDB数据获取接口：%s %s", url, string(reqBody))

	respBody, err := common.HttpPostReq(url, reqBody, nil)
	if err != nil {
		return nil, fmt.Errorf(`err msg: %s, original: %s`, err.Error(), respBody)
	}

	hostsResp := new(HostsResponse)
	err = json.Unmarshal(respBody, &hostsResp)
	if err != nil {
		return nil, fmt.Errorf(`err msg: %s, original: %s`, err.Error(), respBody)
	}

	if !hostsResp.Success { // cmdb 接口对于失败数据的返回格式不一致，目前失败请求的响应可能并没有 success 等这些字段
		return nil, fmt.Errorf(`%s`, string(respBody))
	}
	return hostsResp.Result.Result, nil
}

// generateMockHostData 生成模拟的主机硬件数据
func (c *Cmdb) generateMockHostData(ipList []string) []*HostData {
	var mockData []*HostData
	
	// 预定义的硬件配置模板
	hardwareTemplates := []HostExtInfo{
		{Disk: 500, Ram: 16, Vcpus: 4},   // 小型配置
		{Disk: 1000, Ram: 32, Vcpus: 8},  // 中型配置
		{Disk: 2000, Ram: 64, Vcpus: 16}, // 大型配置
		{Disk: 4000, Ram: 128, Vcpus: 32}, // 超大配置
	}
	
	// 模拟主机名前缀
	hostnamePrefixes := []string{"web-server", "db-server", "app-server", "cache-server", "monitor-server"}
	
	for i, ip := range ipList {
		// 选择硬件模板（按IP循环选择）
		template := hardwareTemplates[i%len(hardwareTemplates)]
		
		// 生成主机名
		hostnamePrefix := hostnamePrefixes[i%len(hostnamePrefixes)]
		hostname := fmt.Sprintf("%s-%02d", hostnamePrefix, (i%20)+1)
		
		// 添加一些随机性到硬件配置
		diskVariation := int64(i%200 - 100) // -100 到 +100 GB的变化
		ramVariation := int64(i%8 - 4)      // -4 到 +4 GB的变化
		
		mockHost := &HostData{
			DomainNum:    fmt.Sprintf("DOM%03d", i+1),
			HostName:     hostname,
			HostIp:       ip,
			HostType:     "云主机",
			HostOwner:    "DBA",
			H3CId:        fmt.Sprintf("H3C-%06d", i+1),
			H3CStatus:    "运行中",
			HostExtInfo: &HostExtInfo{
				Disk:  template.Disk + diskVariation,
				Ram:   template.Ram + ramVariation,
				Vcpus: template.Vcpus,
			},
			IfH3CSync:    "是",
			H3CImageId:   fmt.Sprintf("IMG-%06d", i+1),
			H3CHmName:    fmt.Sprintf("HM-%s", hostname),
			IsDelete:     "否",
			BusinessNum:  fmt.Sprintf("BIZ%03d", i+1),
			HostLocInfo: &HostLoc{
				LeafNumber:      fmt.Sprintf("LEAF-%02d", (i%10)+1),
				RackNumber:      fmt.Sprintf("RACK-%02d", (i%20)+1),
				RackHeight:      int32((i%40) + 2),  // 2-42U
				RackStartNumber: int32((i%38) + 1),  // 1-39U
				FromFactor:      int32((i%4) + 1),   // 1-4U
				SerialNumber:    fmt.Sprintf("SN%d%06d", time.Now().Year(), i+1),
			},
			DeployAppInfoList: []*AppInfo{
				{
					AppName:    fmt.Sprintf("app-%s", hostnamePrefix),
					DataSource: "Mock-CMDB",
				},
			},
		}
		
		mockData = append(mockData, mockHost)
		
		logx.Infof("生成模拟数据 - IP: %s, 主机名: %s, 配置: %dGB磁盘/%dGB内存/%d核CPU", 
			ip, hostname, mockHost.HostExtInfo.Disk, mockHost.HostExtInfo.Ram, mockHost.HostExtInfo.Vcpus)
	}
	
	return mockData
}

// UpdateHosts
//
//	@Description: 调用 CMDB 的接口向它批量写入数据
//	@receiver c
//	@param hostsApp
//	@param opGrpId
//	@return *HostInjectResponse
//	@return error
//
// func (c *Cmdb) UpdateHosts(hostsApp []*AppData) (*HostInjectResponse, error) {
//func (c *Cmdb) UpdateHosts(hostsApp []*model.HostPoolElementDetail, opGrpId string) (*HostInjectResponse, error) {
//	hostInjectResp := new(HostInjectResponse)
//	if len(hostsApp) > 1000 {
//		logx.Statf("单次更新主机信息数量大于1000")
//	} else if len(hostsApp) == 0 {
//		logx.Statf("更新主机信息数量为0")
//		hostInjectResp.Result = true
//		hostInjectResp.Success = true
//		return hostInjectResp, nil
//	}
//
//	// 必要的参数初始化
//	if opGrpId == "" {
//		opGrpId = "23" // 23 在 cmdb 系统中目前代表DBA组，应当通过正确参数`配置/传递`来实现，而不是通过硬编码，这里是为了处理忘记配置该参数的情况
//	}
//
//	url := fmt.Sprintf("%s/hostapp/updateHostAppByDeploy", c.Url)
//
//	valueOfSqlNullString := func(v sql.NullString) *string {
//		if v.Valid {
//			return &v.String
//		}
//		return nil
//	}
//	// cmdb 接口参数格式比较特别，导致必须重新组织数据，再调用接口，实现批量写入
//	// 目前并未对调用cmdb接口写入的每一批次的写入量做控制或检测，目前CMDB的api提供方认为由于我们目前的服务器数量有限，因此这不是问题
//	appDataInjectMap, err := func(hostApps []*model.HostPoolElementDetail) (map[string]*HostInjectRequest, error) { // 获得相同 appName 的所有主机，一次性更新，符合 api 接口请求格式
//		resMap := make(map[string]*HostInjectRequest)
//		var resKey string
//		for _, app := range hostApps {
//			if !c.validateApp(app) {
//				logx.Infof("发现无效的主机app信息，将被忽略，建议对后端数据做排查: Hostname: %s, HostIp: %s, Type: %s", app.HostName, app.HostIp, app.HostType)
//				continue
//			}
//
//			resKey = fmt.Sprintf("%s-%s", app.AppName, app.ClusterGroupName.String)
//
//			if _, found := resMap[resKey]; found {
//				resMap[resKey].HostIpList = append(resMap[resKey].HostIpList, app.HostIp)
//			} else {
//				bizGrpId, found := c.ReadableBizGroupMapping[app.DepartmentLineName.String]
//				if !found {
//					return nil, errorx.NewErrCodeMsg(errorx.ERROR_CMDB_BIZ_GRP_RELATIONSHIP,
//						fmt.Sprintf("异常业务组名称：%s", app.DepartmentLineName.String))
//				}
//
//				resMap[resKey] = new(HostInjectRequest)
//				resMap[resKey].AppName = resKey
//				resMap[resKey].RpcLogId = c.Token
//				resMap[resKey].HostIpList = append(resMap[resKey].HostIpList, app.HostIp)
//				resMap[resKey].BizGroup = &bizGrpId.CMDBCode
//				resMap[resKey].BizGroupReadable = valueOfSqlNullString(app.DepartmentLineName)
//				//resMap[resKey].OpsBizGroup = valueOfSqlNullString(app.ProjectName)
//				resMap[resKey].OpsBizGroup = &opGrpId // 所有DBA写回的机器运维组都是23
//				resMap[resKey].Owner = valueOfSqlNullString(app.Developer)
//				resMap[resKey].Operator = valueOfSqlNullString(app.Dba)
//				resMap[resKey].SoftwareName = app.AppName
//				resMap[resKey].SoftwareVersion = valueOfSqlNullString(app.Version)
//				resMap[resKey].UpdateBy = opGrpId
//			}
//		}
//		return resMap, nil
//	}(hostsApp)
//
//	if err != nil {
//		return nil, errorx.NewErrCodeMsg(errorx.ERROR_CMDB_INJECT, fmt.Sprintf(`err msg: %s`, err.Error()))
//	}
//
//	for _, appData := range appDataInjectMap {
//		reqBody, err := json.Marshal(&appData)
//		if err != nil {
//			return nil, err
//		}
//		logx.Infof("请求CMDB数据注入接口：%s %s", url, string(reqBody))
//
//		respBody, err := common.HttpPostReq(url, reqBody, nil)
//		if err != nil {
//			return nil, errorx.NewErrCodeMsg(errorx.ERROR_CMDB_INJECT, fmt.Sprintf(`err msg: %s, original: %s`, err.Error(), respBody))
//		}
//		err = json.Unmarshal(respBody, &hostInjectResp)
//		if err != nil {
//			return nil, errorx.NewErrCodeMsg(errorx.ERROR_CMDB_INJECT, fmt.Sprintf(`err msg: %s, original: %s`, err.Error(), respBody))
//		}
//
//		if !hostInjectResp.Success { // cmdb 接口对于失败数据的返回格式不一致，目前失败请求的响应可能并没有 success 等这些字段
//			return nil, errorx.NewErrCodeMsg(errorx.ERROR_CMDB_INJECT, string(respBody))
//		}
//	}
//	return hostInjectResp, nil
//}

// validateApp
//
//	@Description: 验证主机 app 数据是否合法，例如主机不能没有有效的ip地址
//	@receiver c
//	@param app
//	@return bool
//func (c *Cmdb) validateApp(app *model.HostPoolElementDetail) bool {
//	if net.ParseIP(app.HostIp) == nil {
//		return false
//	}
//
//	return true
//}
