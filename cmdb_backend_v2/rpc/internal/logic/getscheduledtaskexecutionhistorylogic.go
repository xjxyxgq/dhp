package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetScheduledTaskExecutionHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetScheduledTaskExecutionHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetScheduledTaskExecutionHistoryLogic {
	return &GetScheduledTaskExecutionHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取定时任务执行历史
func (l *GetScheduledTaskExecutionHistoryLogic) GetScheduledTaskExecutionHistory(in *cmpool.GetScheduledTaskExecutionHistoryReq) (*cmpool.GetScheduledTaskExecutionHistoryResp, error) {
	// 参数校验
	if in.ScheduledTaskId <= 0 {
		l.Errorf("定时任务ID不能为空或小于等于0")
		return &cmpool.GetScheduledTaskExecutionHistoryResp{
			Success: false,
			Message: "参数错误：任务ID不能为空",
		}, nil
	}

	// 限制数量默认值
	limit := in.Limit
	if limit <= 0 || limit > 100 {
		limit = 20 // 默认返回20条记录
	}

	// 查询执行历史记录
	historyRecords, err := l.svcCtx.ScheduledTaskHistoryModel.FindByTaskId(in.ScheduledTaskId, int32(limit))
	if err != nil {
		l.Errorf("查询定时任务执行历史失败: %v", err)
		return &cmpool.GetScheduledTaskExecutionHistoryResp{
			Success: false,
			Message: "查询执行历史失败",
		}, nil
	}

	// 转换为响应格式
	var responseHistories []*cmpool.ScheduledTaskExecutionHistory
	for _, history := range historyRecords {
		// 获取该次执行的主机详细信息
		hostDetails, err := l.getHostExecutionDetails(history.ExecutionTaskId)
		if err != nil {
			l.Errorf("获取主机执行详情失败: %v", err)
			// 不阻断整个查询，继续处理下一条记录
		}

		responseHistory := &cmpool.ScheduledTaskExecutionHistory{
			Id:              history.Id,
			ScheduledTaskId: history.ScheduledTaskId,
			ExecutionTaskId: history.ExecutionTaskId,
			ExecutionTime:   history.ExecutionTime.Format("2006-01-02 15:04:05"),
			ExecutionStatus: history.ExecutionStatus,
			TotalHosts:      int32(history.TotalHosts),
			SuccessHosts:    int32(history.SuccessHosts),
			FailedHosts:     int32(history.FailedHosts),
			ErrorMessage:    history.ErrorMessage.String,
			HostDetails:     hostDetails,
		}
		responseHistories = append(responseHistories, responseHistory)
	}

	return &cmpool.GetScheduledTaskExecutionHistoryResp{
		Success:        true,
		Message:        "查询成功",
		HistoryRecords: responseHistories,
	}, nil
}

// getHostExecutionDetails 获取主机执行详情，包含完整的hardware_resource_verification信息
func (l *GetScheduledTaskExecutionHistoryLogic) getHostExecutionDetails(executionTaskId string) ([]*cmpool.HostExecutionDetail, error) {
	// 通过HardwareResourceVerificationModel查询主机执行详情
	verificationRecords, err := l.svcCtx.HardwareResourceVerificationModel.FindByTaskId(l.ctx, executionTaskId)
	if err != nil {
		return nil, err
	}

	var hostDetails []*cmpool.HostExecutionDetail
	for _, record := range verificationRecords {
		// 处理空值和NULL值
		startTime := ""
		if record.StartTime.Valid {
			startTime = record.StartTime.Time.Format("2006-01-02 15:04:05")
		}

		endTime := ""
		if record.EndTime.Valid {
			endTime = record.EndTime.Time.Format("2006-01-02 15:04:05")
		}

		stdoutLog := ""
		if record.StdoutLog.Valid {
			stdoutLog = record.StdoutLog.String
		}

		stderrLog := ""
		if record.StderrLog.Valid {
			stderrLog = record.StderrLog.String
		}

		resultSummary := ""
		if record.ResultSummary.Valid {
			resultSummary = record.ResultSummary.String
		}

		sshError := ""
		if record.SshError.Valid {
			sshError = record.SshError.String
		}

		exitCode := int32(0)
		if record.ExitCode.Valid {
			exitCode = int32(record.ExitCode.Int64)
		}

		createTime := ""
		if !record.CreateTime.IsZero() {
			createTime = record.CreateTime.Format("2006-01-02 15:04:05")
		}

		hostDetail := &cmpool.HostExecutionDetail{
			HostIp:          record.HostIp,
			ResourceType:    record.ResourceType,
			TargetPercent:   int32(record.TargetPercent),
			Duration:        int32(record.Duration),
			ExecutionStatus: record.ExecutionStatus,
			StartTime:       startTime,
			EndTime:         endTime,
			ExitCode:        exitCode,
			StdoutLog:       stdoutLog,
			StderrLog:       stderrLog,
			ResultSummary:   resultSummary,
			SSHError:        sshError,
			CreateTime:      createTime,
		}
		hostDetails = append(hostDetails, hostDetail)
	}

	return hostDetails, nil
}
