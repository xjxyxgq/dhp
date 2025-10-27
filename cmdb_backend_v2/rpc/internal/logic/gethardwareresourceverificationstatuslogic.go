package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHardwareResourceVerificationStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHardwareResourceVerificationStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHardwareResourceVerificationStatusLogic {
	return &GetHardwareResourceVerificationStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取硬件资源验证状态
func (l *GetHardwareResourceVerificationStatusLogic) GetHardwareResourceVerificationStatus(in *cmpool.GetHardwareResourceVerificationStatusReq) (*cmpool.GetHardwareResourceVerificationStatusResp, error) {
	// 查询验证记录
	records, err := l.svcCtx.HardwareResourceVerificationModel.FindByHostIpList(l.ctx, in.HostIpList, in.ResourceType)
	if err != nil {
		return &cmpool.GetHardwareResourceVerificationStatusResp{
			Success: false,
			Message: fmt.Sprintf("查询验证记录失败: %v", err),
		}, nil
	}

	// 转换为响应格式
	var verificationRecords []*cmpool.HardwareResourceVerificationStatus
	for _, record := range records {
		status := &cmpool.HardwareResourceVerificationStatus{
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
			status.StartTime = record.StartTime.Time.Format("2006-01-02 15:04:05")
		}
		if record.EndTime.Valid {
			status.EndTime = record.EndTime.Time.Format("2006-01-02 15:04:05")
		}
		if record.ExitCode.Valid {
			status.ExitCode = int32(record.ExitCode.Int64)
		}
		if record.ResultSummary.Valid {
			status.ResultSummary = record.ResultSummary.String
		}
		if record.SshError.Valid {
			status.SSHError = record.SshError.String
		}

		verificationRecords = append(verificationRecords, status)
	}

	return &cmpool.GetHardwareResourceVerificationStatusResp{
		Success:             true,
		Message:             "查询成功",
		VerificationRecords: verificationRecords,
	}, nil
}
