package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScheduledTaskExecutionDetailsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetScheduledTaskExecutionDetailsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScheduledTaskExecutionDetailsLogic {
	return &GetScheduledTaskExecutionDetailsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取定时任务执行详情
func (l *GetScheduledTaskExecutionDetailsLogic) GetScheduledTaskExecutionDetails(in *cmpool.GetScheduledTaskExecutionDetailsReq) (*cmpool.GetScheduledTaskExecutionDetailsResp, error) {
	if in.ExecutionTaskId == "" {
		l.Errorf("执行任务ID不能为空")
		return &cmpool.GetScheduledTaskExecutionDetailsResp{
			Success: false,
			Message: "参数错误：执行任务ID不能为空",
		}, nil
	}

	// 查询执行历史记录
	executionHistory, err := l.svcCtx.ScheduledTaskHistoryModel.FindByExecutionTaskId(in.ExecutionTaskId)
	if err != nil {
		l.Errorf("查询执行历史失败: %v", err)
		return &cmpool.GetScheduledTaskExecutionDetailsResp{
			Success: false,
			Message: "查询执行历史失败",
		}, nil
	}

	if executionHistory == nil {
		return &cmpool.GetScheduledTaskExecutionDetailsResp{
			Success: false,
			Message: "未找到对应的执行记录",
		}, nil
	}

	// 查询定时任务信息
	scheduledTask, err := l.svcCtx.ScheduledTaskModel.FindOne(l.ctx, executionHistory.ScheduledTaskId)
	if err != nil {
		l.Errorf("查询定时任务失败: %v", err)
		return &cmpool.GetScheduledTaskExecutionDetailsResp{
			Success: false,
			Message: "查询定时任务失败",
		}, nil
	}

	// 查询硬件资源验证记录
	verificationRecords, err := l.svcCtx.HardwareResourceVerificationModel.FindByTaskId(l.ctx, in.ExecutionTaskId)
	if err != nil {
		l.Errorf("查询硬件资源验证记录失败: %v", err)
		return &cmpool.GetScheduledTaskExecutionDetailsResp{
			Success: false,
			Message: "查询验证记录失败",
		}, nil
	}

	// 构建主机详情列表
	var hostDetails []*cmpool.ScheduledTaskExecutionDetail
	for _, record := range verificationRecords {
		hostDetail := &cmpool.ScheduledTaskExecutionDetail{
			HostIp:          record.HostIp,
			ResourceType:    record.ResourceType,
			TargetPercent:   int32(record.TargetPercent),
			Duration:        int32(record.Duration),
			ExecutionStatus: record.ExecutionStatus,
			StartTime:       record.StartTime.Time.Format("2006-01-02 15:04:05"),
			EndTime:         record.EndTime.Time.Format("2006-01-02 15:04:05"),
			ExitCode:        int32(record.ExitCode.Int64),
			StdoutLog:       record.StdoutLog.String,
			StderrLog:       record.StderrLog.String,
			ResultSummary:   record.ResultSummary.String,
			SSHError:        record.SshError.String,
			CreateTime:      record.CreateTime.Format("2006-01-02 15:04:05"),
		}
		hostDetails = append(hostDetails, hostDetail)
	}

	// 构建响应数据
	executionInfo := &cmpool.ScheduledTaskExecutionInfo{
		ScheduledTaskId: executionHistory.ScheduledTaskId,
		TaskName:        scheduledTask.TaskName,
		ExecutionTaskId: executionHistory.ExecutionTaskId,
		ExecutionTime:   executionHistory.ExecutionTime.Format("2006-01-02 15:04:05"),
		ExecutionStatus: executionHistory.ExecutionStatus,
		TotalHosts:      int32(executionHistory.TotalHosts),
		SuccessHosts:    int32(executionHistory.SuccessHosts),
		FailedHosts:     int32(executionHistory.FailedHosts),
		ResourceType:    scheduledTask.ResourceType,
		TargetPercent:   int32(scheduledTask.TargetPercent),
		Duration:        int32(scheduledTask.Duration),
		HostDetails:     hostDetails,
	}

	return &cmpool.GetScheduledTaskExecutionDetailsResp{
		Success: true,
		Message: "查询成功",
		Data:    executionInfo,
	}, nil
}
