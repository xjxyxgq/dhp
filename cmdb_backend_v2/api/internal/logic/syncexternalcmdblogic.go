package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncExternalCmdbLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncExternalCmdbLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncExternalCmdbLogic {
	return &SyncExternalCmdbLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncExternalCmdbLogic) SyncExternalCmdb(req *types.SyncExternalCmdbRequest) (resp *types.SyncExternalCmdbResponse, err error) {
	// 调用RPC服务
	rpcReq := &cmpool.SyncExternalCmdbReq{
		PageSize:    int32(req.PageSize),
		HostOwner:   int32(req.HostOwner),
		ForceUpdate: req.ForceUpdate,
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.SyncExternalCmdb(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC同步外部CMDB失败: %v", err)
		return nil, err
	}

	// 转换RPC响应到API响应
	var syncResults []types.ExternalCmdbHost
	for _, result := range rpcResp.SyncResults {
		syncResult := types.ExternalCmdbHost{
			CmdbId:       result.CmdbId,
			DomainNum:    result.DomainNum,
			HostName:     result.HostName,
			HostIp:       result.HostIp,
			HostType:     result.HostType,
			HostOwner:    result.HostOwner,
			OpsIamCode:   result.OpsIamCode,
			OwnerGroup:   result.OwnerGroup,
			OwnerIamCode: result.OwnerIamCode,
			H3cId:        result.H3CId,
			H3cStatus:    result.H3CStatus,
			Disk:         result.Disk,
			Ram:          result.Ram,
			Vcpus:        result.Vcpus,
			CreatedAt:    result.CreatedAt,
			UpdatedAt:    result.UpdatedAt,
			IfH3cSync:    result.IfH3CSync,
			H3cImageId:   result.H3CImageId,
			H3cHmName:    result.H3CHmName,
			IsDelete:     result.IsDelete,
			AppName:      result.AppName,
			DataSource:   result.DataSource,
			BizGroup:     result.BizGroup,
			OpsBizGroup:  result.OpsBizGroup,
			Message:      result.Message,
			Success:      result.Success,
		}
		syncResults = append(syncResults, syncResult)
	}

	resp = &types.SyncExternalCmdbResponse{
		Success:        rpcResp.Success,
		Message:        rpcResp.Message,
		TotalPages:     int(rpcResp.TotalPages),
		ProcessedPages: int(rpcResp.ProcessedPages),
		TotalHosts:     int(rpcResp.TotalHosts),
		SyncedHosts:    int(rpcResp.SyncedHosts),
		UpdatedHosts:   int(rpcResp.UpdatedHosts),
		FailedHosts:    int(rpcResp.FailedHosts),
		SyncResults:    syncResults,
	}

	return resp, nil
}
