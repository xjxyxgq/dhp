package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteExternalSyncFullSyncLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteExternalSyncFullSyncLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteExternalSyncFullSyncLogic {
	return &ExecuteExternalSyncFullSyncLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ExecuteExternalSyncFullSync 统一的全量同步接口（支持 ES 和 CMSys）
func (l *ExecuteExternalSyncFullSyncLogic) ExecuteExternalSyncFullSync(in *cmpool.ExecuteExternalSyncFullSyncReq) (*cmpool.ExecuteExternalSyncResp, error) {
	// 1. 验证 data_source 参数
	if in.DataSource == "" {
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: "数据源类型不能为空 (elasticsearch/cmsys)",
		}, nil
	}

	l.Logger.Infof("开始统一执行外部全量同步: 数据源=%s, 任务名=%s", in.DataSource, in.TaskName)

	// 2. 根据数据源路由到对应实现
	switch in.DataSource {
	case "elasticsearch", "es":
		return l.executeFromES(in)
	case "cmsys":
		return l.executeFromCMSys(in)
	default:
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: fmt.Sprintf("不支持的数据源类型: %s (支持: elasticsearch/es, cmsys)", in.DataSource),
		}, nil
	}
}

// executeFromES 调用 ES 全量同步逻辑
func (l *ExecuteExternalSyncFullSyncLogic) executeFromES(in *cmpool.ExecuteExternalSyncFullSyncReq) (*cmpool.ExecuteExternalSyncResp, error) {
	// 转换为 ES 特定请求
	esReq := &cmpool.ExecuteESSyncFullSyncReq{
		GroupName:      in.GroupName,
		QueryTimeRange: in.QueryTimeRange,
		EsEndpoint:     in.EsEndpoint,
		TaskName:       in.TaskName,
	}

	// 调用 ES 全量同步 Logic
	esLogic := NewExecuteEsSyncFullSyncLogic(l.ctx, l.svcCtx)
	esResp, err := esLogic.ExecuteEsSyncFullSync(esReq)
	if err != nil {
		return nil, err
	}

	// 转换为统一响应格式
	return &cmpool.ExecuteExternalSyncResp{
		Success:       esResp.Success,
		Message:       esResp.Message,
		ExecutionId:   esResp.ExecutionId,
		TotalHosts:    esResp.TotalHosts,
		SuccessCount:  esResp.UpdatedHostsCount, // ES全量同步的"更新数"对应统一接口的"成功数"
		FailedCount:   esResp.FailedCount,
		SuccessIpList: append(esResp.NewHostIpList, esResp.UpdatedHostIpList...), // 成功IP = 新增 + 更新
		FailedIpList:  esResp.FailedIpList,
		// ES 全量同步特有字段
		NewHostsCount:     esResp.NewHostsCount,
		NewHostIpList:     esResp.NewHostIpList,
		UpdatedHostsCount: esResp.UpdatedHostsCount,
		UpdatedHostIpList: esResp.UpdatedHostIpList,
		// CMSys 字段固定为0/空
		NotInDatasourceCount:  0,
		NotInDatasourceIpList: []string{},
		NotInPoolCount:        0,
		NotInPoolIpList:       []string{},
	}, nil
}

// executeFromCMSys 调用 CMSys 全量同步逻辑
func (l *ExecuteExternalSyncFullSyncLogic) executeFromCMSys(in *cmpool.ExecuteExternalSyncFullSyncReq) (*cmpool.ExecuteExternalSyncResp, error) {
	// 转换为 CMSys 特定请求
	cmsysReq := &cmpool.ExecuteCMSysSyncFullSyncReq{
		Query:    in.CmsysQuery,
		TaskName: in.TaskName,
	}

	// 调用 CMSys 全量同步 Logic
	cmsysLogic := NewExecuteCmsysSyncFullSyncLogic(l.ctx, l.svcCtx)
	cmsysResp, err := cmsysLogic.ExecuteCmsysSyncFullSync(cmsysReq)
	if err != nil {
		return nil, err
	}

	// 转换为统一响应格式
	return &cmpool.ExecuteExternalSyncResp{
		Success:       cmsysResp.Success,
		Message:       cmsysResp.Message,
		ExecutionId:   cmsysResp.ExecutionId,
		TotalHosts:    cmsysResp.TotalHosts,
		SuccessCount:  cmsysResp.SuccessCount,
		FailedCount:   cmsysResp.FailedCount,
		SuccessIpList: cmsysResp.SuccessIpList,
		FailedIpList:  cmsysResp.FailedIpList,
		// CMSys 全量同步特有字段
		NotInDatasourceCount:  cmsysResp.NotInDatasourceCount,
		NotInDatasourceIpList: cmsysResp.NotInDatasourceIpList,
		NewHostsCount:         cmsysResp.NewHostsCount,
		NewHostIpList:         cmsysResp.NewHostIpList,
		UpdatedHostsCount:     cmsysResp.UpdatedHostsCount,
		UpdatedHostIpList:     cmsysResp.UpdatedHostIpList,
		// ES 字段固定为0/空
		NotInPoolCount:  0,
		NotInPoolIpList: []string{},
	}, nil
}
