package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHardwareResourceVerificationHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHardwareResourceVerificationHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHardwareResourceVerificationHistoryLogic {
	return &GetHardwareResourceVerificationHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取硬件资源验证历史记录
func (l *GetHardwareResourceVerificationHistoryLogic) GetHardwareResourceVerificationHistory(in *cmpool.GetHardwareResourceVerificationHistoryReq) (*cmpool.GetHardwareResourceVerificationHistoryResp, error) {
	// 验证必需参数
	if in.HostIp == "" {
		return &cmpool.GetHardwareResourceVerificationHistoryResp{
			Success: false,
			Message: "主机IP不能为空",
		}, nil
	}

	// 查询历史记录
	records, err := l.svcCtx.HardwareResourceVerificationModel.FindByHostIp(l.ctx, in.HostIp, in.ResourceType, in.Limit)
	if err != nil {
		return &cmpool.GetHardwareResourceVerificationHistoryResp{
			Success: false,
			Message: fmt.Sprintf("查询历史记录失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	var historyRecords []*cmpool.HardwareResourceVerificationHistory
	for _, record := range records {
		history := &cmpool.HardwareResourceVerificationHistory{
			Id:              int64(record.Id),
			TaskId:          record.TaskId,
			HostIp:          record.HostIp,
			ResourceType:    record.ResourceType,
			TargetPercent:   int32(record.TargetPercent),
			Duration:        int32(record.Duration),
			ExecutionStatus: record.ExecutionStatus,
			CreateTime:      record.CreateTime.Format("2006-01-02 15:04:05"),
		}

		// 处理可空字段
		if record.StartTime.Valid {
			history.StartTime = record.StartTime.Time.Format("2006-01-02 15:04:05")
		}
		if record.EndTime.Valid {
			history.EndTime = record.EndTime.Time.Format("2006-01-02 15:04:05")
		}
		if record.ExitCode.Valid {
			history.ExitCode = int32(record.ExitCode.Int64)
		}
		if record.StdoutLog.Valid {
			history.StdoutLog = record.StdoutLog.String
		}
		if record.StderrLog.Valid {
			history.StderrLog = record.StderrLog.String
		}
		if record.ResultSummary.Valid {
			history.ResultSummary = record.ResultSummary.String
		}
		if record.SshError.Valid {
			history.SSHError = record.SshError.String
		}

		historyRecords = append(historyRecords, history)
	}

	return &cmpool.GetHardwareResourceVerificationHistoryResp{
		Success:        true,
		Message:        "查询成功",
		HistoryRecords: historyRecords,
	}, nil
}
