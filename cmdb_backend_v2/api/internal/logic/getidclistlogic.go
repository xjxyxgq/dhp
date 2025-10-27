package logic

import (
	"context"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetIdcListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetIdcListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetIdcListLogic {
	return &GetIdcListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetIdcListLogic) GetIdcList() (resp *types.IdcListResponse, err error) {
	// 调用RPC服务获取IDC列表
	rpcReq := &cmpool.GetIdcConfListReq{}
	
	rpcResp, err := l.svcCtx.CmpoolRpc.GetIdcConfList(l.ctx, rpcReq)
	if err != nil {
		logx.Errorf("调用RPC服务失败: %v", err)
		return &types.IdcListResponse{
			Success: false,
			Message: "服务调用失败",
			List:    []types.IdcInfo{},
		}, nil
	}

	if !rpcResp.Success {
		logx.Errorf("RPC返回失败: %s", rpcResp.Message)
		return &types.IdcListResponse{
			Success: false,
			Message: rpcResp.Message,
			List:    []types.IdcInfo{},
		}, nil
	}

	// 转换IDC信息列表
	var idcList []types.IdcInfo
	for _, rpcIdc := range rpcResp.IdcConfList {
		idc := types.IdcInfo{
			ID:             int(rpcIdc.Id),
			IdcName:        rpcIdc.IdcName,
			IdcCode:        rpcIdc.IdcCode,
			IdcLocation:    rpcIdc.IdcLocation,
			IdcDescription: rpcIdc.IdcDescription,
		}
		idcList = append(idcList, idc)
	}

	return &types.IdcListResponse{
		Success: true,
		Message: "查询成功",
		List:    idcList,
	}, nil
}
