package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetExternalSyncExecutionDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetExternalSyncExecutionDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetExternalSyncExecutionDetailLogic {
	return &GetExternalSyncExecutionDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取外部数据同步执行详情
func (l *GetExternalSyncExecutionDetailLogic) GetExternalSyncExecutionDetail(in *cmpool.GetExternalSyncExecutionDetailReq) (*cmpool.GetExternalSyncExecutionDetailResp, error) {
	// 1. 验证参数
	if in.ExecutionId == 0 {
		return &cmpool.GetExternalSyncExecutionDetailResp{
			Success: false,
			Message: "执行记录ID不能为空",
		}, nil
	}

	// 2. 查询执行记录基本信息
	executionLog, err := l.svcCtx.ExternalSyncExecutionLogModel.FindOne(l.ctx, uint64(in.ExecutionId))
	if err != nil {
		l.Logger.Errorf("查询执行记录失败: %v", err)
		return &cmpool.GetExternalSyncExecutionDetailResp{
			Success: false,
			Message: "执行记录不存在或查询失败",
		}, nil
	}

	// 3. 查询执行详情列表
	details, err := l.svcCtx.ExternalSyncExecutionDetailModel.FindByExecutionId(l.ctx, uint64(in.ExecutionId))
	if err != nil {
		l.Logger.Errorf("查询执行详情失败: %v", err)
		return &cmpool.GetExternalSyncExecutionDetailResp{
			Success: false,
			Message: "查询执行详情失败",
		}, nil
	}

	// 4. 转换执行记录为 Proto 格式
	protoLog := &cmpool.ExternalSyncExecutionLog{
		Id:              int64(executionLog.Id),
		TaskId:          int64(executionLog.TaskId),
		TaskName:        executionLog.TaskName,
		ExecutionTime:   executionLog.ExecutionTime.Format("2006-01-02 15:04:05"),
		ExecutionStatus: executionLog.ExecutionStatus,
		TotalHosts:      int32(executionLog.TotalHosts),
		SuccessCount:    int32(executionLog.SuccessCount),
		FailedCount:     int32(executionLog.FailedCount),
		NotInPoolCount:  int32(executionLog.NotInPoolCount),
		ErrorMessage:    executionLog.ErrorMessage.String,
		DurationMs:      int32(executionLog.DurationMs.Int64),
		QueryTimeRange:  executionLog.QueryTimeRange.String,
		CreatedAt:       executionLog.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 5. 转换详情列表为 Proto 格式
	var detailList []*cmpool.ExternalSyncExecutionDetail
	for _, detail := range details {
		protoDetail := &cmpool.ExternalSyncExecutionDetail{
			HostIp:         detail.HostIp,
			HostName:       detail.HostName.String,
			SyncStatus:     detail.SyncStatus,
			ErrorMessage:   detail.ErrorMessage.String,
			MaxCpu:         detail.MaxCpu.Float64,
			AvgCpu:         detail.AvgCpu.Float64,
			MaxMemory:      detail.MaxMemory.Float64,
			AvgMemory:      detail.AvgMemory.Float64,
			MaxDisk:        detail.MaxDisk.Float64,
			AvgDisk:        detail.AvgDisk.Float64,
			DataPointCount: int32(detail.DataPointCount.Int64),
			CreatedAt:      detail.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		detailList = append(detailList, protoDetail)
	}

	l.Logger.Infof("查询执行详情成功: ExecutionId=%d, 共%d条详情记录", in.ExecutionId, len(detailList))

	// 6. 构造响应 - 使用 Data 字段包装
	return &cmpool.GetExternalSyncExecutionDetailResp{
		Success: true,
		Message: "查询成功",
		Data: &cmpool.ExternalSyncExecutionInfo{
			ExecutionLog: protoLog,
			Details:      detailList,
		},
	}, nil
}
