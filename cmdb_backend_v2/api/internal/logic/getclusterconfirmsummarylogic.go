package logic

import (
	"context"
	"encoding/json"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterConfirmSummaryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetClusterConfirmSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterConfirmSummaryLogic {
	return &GetClusterConfirmSummaryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetClusterConfirmSummaryLogic) GetClusterConfirmSummary() (resp *types.ClusterConfirmSummary, err error) {
	l.Logger.Info("开始调用RPC获取集群确认摘要")

	// 调用RPC服务获取集群确认摘要
	rpcReq := &cmpool.ClusterConfirmSummaryReq{
		Days: 7, // 默认获取最近7天的数据
	}

	rpcResp, err := l.svcCtx.CmpoolRpc.GetClusterConfirmSummary(l.ctx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC获取集群确认摘要失败: %v", err)
		return nil, err
	}

	if !rpcResp.Success {
		l.Logger.Errorf("RPC返回失败: %s", rpcResp.Message)
		return &types.ClusterConfirmSummary{
			ReportFileURL: "",
			PluginResults: make(map[string]interface{}),
		}, nil
	}

	// 解析插件结果JSON字符串
	var pluginResults map[string]interface{}
	if err := json.Unmarshal([]byte(rpcResp.ClusterConfirmSummary.PluginResults), &pluginResults); err != nil {
		l.Logger.Errorf("解析插件结果JSON失败: %v", err)
		return &types.ClusterConfirmSummary{
			ReportFileURL: rpcResp.ClusterConfirmSummary.ReportFileURL,
			PluginResults: map[string]interface{}{
				"parse_error": err.Error(),
			},
		}, nil
	}

	l.Logger.Info("成功获取集群确认摘要")
	return &types.ClusterConfirmSummary{
		ReportFileURL: rpcResp.ClusterConfirmSummary.ReportFileURL,
		PluginResults: pluginResults,
	}, nil
}
