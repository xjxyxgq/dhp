package logic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/datasource/cmsys"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteCmsysSyncByFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteCmsysSyncByFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteCmsysSyncByFileLogic {
	return &ExecuteCmsysSyncByFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CMSys按文件执行同步
func (l *ExecuteCmsysSyncByFileLogic) ExecuteCmsysSyncByFile(in *cmpool.ExecuteCMSysSyncByFileReq) (*cmpool.ExecuteCMSysSyncResp, error) {
	startTime := time.Now()

	l.Logger.Infof("开始执行CMSys按文件同步: 文件大小=%d bytes", len(in.FileContent))

	// 1. 验证任务名称
	taskName := in.TaskName
	if taskName == "" {
		taskName = "CMSys文件同步"
	}

	// 2. 解析文件内容，提取IP列表
	ipList := l.parseIPListFromFile(in.FileContent)
	if len(ipList) == 0 {
		return &cmpool.ExecuteCMSysSyncResp{
			Success: false,
			Message: "文件中未找到有效的IP地址",
		}, nil
	}

	l.Logger.Infof("从文件中解析出 %d 个IP地址", len(ipList))

	// 3. 创建CMSys客户端
	cmsysClient := cmsys.NewCMSysClient(
		l.svcCtx.Config.CMSysDataSource.AuthEndpoint,
		l.svcCtx.Config.CMSysDataSource.DataEndpoint,
		l.svcCtx.Config.CMSysDataSource.AppCode,
		l.svcCtx.Config.CMSysDataSource.AppSecret,
		l.svcCtx.Config.CMSysDataSource.Operator,
		l.svcCtx.Config.CMSysDataSource.TimeoutSeconds,
	)

	// 4. 使用批量查询方法查询CMSys数据（一次请求查询所有IP）
	metricsList, err := cmsysClient.QueryHostMetricsByIPs(l.ctx, ipList)
	if err != nil {
		l.Logger.Errorf("批量查询CMSys数据失败: %v", err)
		return &cmpool.ExecuteCMSysSyncResp{
			Success: false,
			Message: fmt.Sprintf("批量查询CMSys数据失败: %v", err),
		}, nil
	}

	l.Logger.Infof("从CMSys查询到 %d 条主机数据", len(metricsList))

	if len(metricsList) == 0 {
		return &cmpool.ExecuteCMSysSyncResp{
			Success: false,
			Message: "CMSys中没有查询到符合条件的数据",
		}, nil
	}

	// 5. 创建执行记录
	executionLogic := NewExecuteCmsysSyncLogic(l.ctx, l.svcCtx)
	executionId, err := executionLogic.CreateExecutionLog(taskName, len(metricsList))
	if err != nil {
		l.Logger.Errorf("创建执行记录失败: %v", err)
		return &cmpool.ExecuteCMSysSyncResp{
			Success: false,
			Message: "创建执行记录失败",
		}, nil
	}

	// 6. 并发同步数据（复用现有逻辑）
	successIps, failedIps, notInDatasourceIps := executionLogic.SyncHostsData(metricsList, executionId)

	// 7. 统计数量
	successCount := len(successIps)
	failedCount := len(failedIps)
	notInDatasourceCount := len(notInDatasourceIps)

	// 8. 更新执行记录
	duration := time.Since(startTime).Milliseconds()
	executionStatus := "success"
	if failedCount > 0 {
		if successCount > 0 {
			executionStatus = "partial"
		} else {
			executionStatus = "failed"
		}
	}

	err = executionLogic.UpdateExecutionLog(executionId, executionStatus, successCount, failedCount, notInDatasourceCount, duration)
	if err != nil {
		l.Logger.Errorf("更新执行记录失败: %v", err)
	}

	l.Logger.Infof("CMSys文件同步完成: 总数=%d, 成功=%d, 失败=%d, 数据源中不存在=%d, 耗时=%dms",
		len(metricsList), successCount, failedCount, notInDatasourceCount, duration)

	return &cmpool.ExecuteCMSysSyncResp{
		Success:      true,
		Message:      fmt.Sprintf("同步完成: 成功%d个, 失败%d个, 数据源中不存在%d个", successCount, failedCount, notInDatasourceCount),
		ExecutionId:  executionId,
		TotalHosts:   int32(len(metricsList)),
		SuccessCount: int32(successCount),
		FailedCount:  int32(failedCount),
		// CMSys特有字段
		NotInDatasourceCount:  int32(notInDatasourceCount),
		NotInDatasourceIpList: notInDatasourceIps,
		// ES字段固定为0/空（统一响应结构要求）
		NotInPoolCount:    0,
		NotInPoolIpList:   []string{},
		NewHostsCount:     0,
		NewHostIpList:     []string{},
		UpdatedHostsCount: 0,
		UpdatedHostIpList: []string{},
		// 通用字段
		SuccessIpList: successIps,
		FailedIpList:  failedIps,
	}, nil
}

// parseIPListFromFile 从文件内容中解析IP列表
func (l *ExecuteCmsysSyncByFileLogic) parseIPListFromFile(fileContent []byte) []string {
	content := string(fileContent)
	lines := strings.Split(content, "\n")

	ipList := make([]string, 0, len(lines))
	for _, line := range lines {
		ip := strings.TrimSpace(line)
		// 跳过空行和注释行
		if ip == "" || strings.HasPrefix(ip, "#") {
			continue
		}
		ipList = append(ipList, ip)
	}

	return ipList
}
