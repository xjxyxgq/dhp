package logic

import (
	"context"
	"fmt"
	"strings"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateExternalSyncTaskLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateExternalSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateExternalSyncTaskLogic {
	return &CreateExternalSyncTaskLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// validateDataSource 验证数据源参数是否合法
func validateDataSource(dataSource string) error {
	normalizedSource := strings.ToLower(strings.TrimSpace(dataSource))
	if normalizedSource != "elasticsearch" && normalizedSource != "es" && normalizedSource != "cmsys" {
		return fmt.Errorf("无效的数据源: %s，必须为 'elasticsearch'、'es' 或 'cmsys'", dataSource)
	}
	return nil
}

// normalizeDataSource 标准化数据源名称
func normalizeDataSource(dataSource string) string {
	normalized := strings.ToLower(strings.TrimSpace(dataSource))
	if normalized == "elasticsearch" || normalized == "es" {
		return "elasticsearch"
	}
	return "cmsys"
}

func (l *CreateExternalSyncTaskLogic) CreateExternalSyncTask(req *types.CreateExternalSyncTaskRequest) (resp *types.CreateExternalSyncTaskResponse, err error) {
	// 参数验证
	if err := validateDataSource(req.DataSource); err != nil {
		l.Logger.Errorf("参数验证失败: %v", err)
		return &types.CreateExternalSyncTaskResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	normalizedSource := normalizeDataSource(req.DataSource)
	l.Logger.Infof("统一接口: 创建外部同步任务, DataSource=%s (标准化后=%s), TaskName=%s", req.DataSource, normalizedSource, req.TaskName)

	// 调用统一的外部资源同步任务创建RPC（支持 ES 和 CMSys）
	l.Logger.Infof("数据源路由: %s -> CreateExternalSyncTask RPC方法", normalizedSource)

	rpcReq := &cmpool.CreateExternalSyncTaskReq{
		DataSource:     normalizedSource,
		TaskName:       req.TaskName,
		Description:    req.Description,
		EsEndpoint:     req.EsEndpoint,
		EsIndexPattern: req.EsIndexPattern,
		CronExpression: req.CronExpression,
		QueryTimeRange: req.QueryTimeRange,
		CmsysQuery:     req.CmsysQuery,
		CreatedBy:      req.CreatedBy,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.CreateExternalSyncTask(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用统一外部同步任务创建RPC失败: %v", err)
		return &types.CreateExternalSyncTaskResponse{
			Success: false,
			Message: fmt.Sprintf("创建外部同步任务失败: %v", err),
		}, nil
	}

	l.Logger.Infof("统一接口执行成功: TaskId=%d, DataSource=%s", rpcResp.TaskId, normalizedSource)
	return &types.CreateExternalSyncTaskResponse{
		Success: rpcResp.Success,
		Message: rpcResp.Message,
		TaskId:  rpcResp.TaskId,
	}, nil
}
