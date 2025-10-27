package logic

import (
	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBackupRestoreCheckInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBackupRestoreCheckInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBackupRestoreCheckInfoLogic {
	return &GetBackupRestoreCheckInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetBackupRestoreCheckInfoLogic) GetBackupRestoreCheckInfo(in *cmpool.BackupRestoreCheckInfoReq) (*cmpool.BackupRestoreCheckInfoResp, error) {
	l.Logger.Info("开始查询备份恢复检查信息")

	// 设置默认限制
	limit := int32(100)
	if in.Limit > 0 {
		limit = in.Limit
	}

	// 从model层查询数据
	dbRecords, err := l.svcCtx.BackupRestoreCheckInfoModel.FindOrderedByTime(l.ctx, limit)
	if err != nil {
		l.Logger.Errorf("查询备份恢复检查信息失败: %v", err)
		return &cmpool.BackupRestoreCheckInfoResp{
			Success: false,
			Message: "查询失败: " + err.Error(),
		}, nil
	}

	// 转换为响应结构体
	var records []*cmpool.BackupRestoreCheckInfo
	for _, dbRecord := range dbRecords {
		record := &cmpool.BackupRestoreCheckInfo{
			Id:       dbRecord.Id,
			CheckSeq: dbRecord.CheckSeq,
			CheckDb:  dbRecord.CheckDb,
		}

		// 处理NULL值
		if dbRecord.CheckSrcIP.Valid {
			record.CheckStatus = "已完成"
		} else {
			record.CheckStatus = "未完成"
		}

		if dbRecord.BackupCheckResult.Valid {
			record.CheckResult = dbRecord.BackupCheckResult.String
		} else {
			record.CheckResult = "待检查"
		}

		if dbRecord.CreatedAt.Valid {
			record.CheckTime = dbRecord.CreatedAt.String
		} else {
			record.CheckTime = ""
		}

		records = append(records, record)
	}

	l.Logger.Infof("成功查询到%d条备份恢复检查信息", len(records))
	return &cmpool.BackupRestoreCheckInfoResp{
		Success:                true,
		Message:                "查询成功",
		BackupRestoreCheckInfo: records,
	}, nil
}
