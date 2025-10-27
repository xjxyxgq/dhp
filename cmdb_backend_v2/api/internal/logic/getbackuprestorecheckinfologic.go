package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBackupRestoreCheckInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBackupRestoreCheckInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBackupRestoreCheckInfoLogic {
	return &GetBackupRestoreCheckInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBackupRestoreCheckInfoLogic) GetBackupRestoreCheckInfo() (resp *types.BackupRestoreCheckInfoListResponse, err error) {
	l.Logger.Info("开始调用RPC获取备份恢复检查信息")

	// 调用RPC服务获取备份恢复检查信息
	rpcReq := &cmpool.BackupRestoreCheckInfoReq{
		Limit: 100, // 默认获取100条记录
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetBackupRestoreCheckInfo(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC获取备份恢复检查信息失败: %v", err)
		return nil, err
	}

	if !rpcResp.Success {
		l.Logger.Errorf("RPC返回失败: %s", rpcResp.Message)
		return &types.BackupRestoreCheckInfoListResponse{
			List: []types.BackupRestoreCheckInfo{},
		}, nil
	}

	// 转换RPC响应到API响应类型
	var records []types.BackupRestoreCheckInfo
	for _, rpcRecord := range rpcResp.BackupRestoreCheckInfo {
		record := types.BackupRestoreCheckInfo{
			ID:          int(rpcRecord.Id),
			CheckSeq:    rpcRecord.CheckSeq,
			CheckDB:     rpcRecord.CheckDb,
			CheckStatus: rpcRecord.CheckStatus,
			CheckResult: rpcRecord.CheckResult,
			CheckTime:   rpcRecord.CheckTime,
		}
		records = append(records, record)
	}

	l.Logger.Infof("成功获取%d条备份恢复检查信息", len(records))
	return &types.BackupRestoreCheckInfoListResponse{
		List: records,
	}, nil
}
