package logic

import (
	"context"
	"database/sql"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type MatchIdcByIpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMatchIdcByIpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MatchIdcByIpLogic {
	return &MatchIdcByIpLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据IP匹配IDC机房
func (l *MatchIdcByIpLogic) MatchIdcByIp(in *cmpool.MatchIdcByIpReq) (*cmpool.MatchIdcByIpResp, error) {
	if in.HostIp == "" {
		return &cmpool.MatchIdcByIpResp{
			Success: false,
			Message: "主机IP不能为空",
		}, nil
	}

	// 使用模型的MatchIdcByIp方法匹配IDC
	config, err := l.svcCtx.IdcConfModel.MatchIdcByIp(l.ctx, in.HostIp)
	if err != nil {
		if err == sql.ErrNoRows {
			return &cmpool.MatchIdcByIpResp{
				Success: true,
				Message: "未匹配到任何IDC机房",
			}, nil
		}
		
		l.Logger.Errorf("匹配IDC机房失败: %v", err)
		return &cmpool.MatchIdcByIpResp{
			Success: false,
			Message: "系统错误，请稍后再试",
		}, nil
	}

	// 转换为protobuf格式
	idcConf := &cmpool.IdcConf{
		Id:         int64(config.Id),
		IdcName:    config.IdcName,
		IdcCode:    config.IdcCode,
		IdcIpRegexp: config.IdcIpRegexp,
		IsActive:   config.IsActive > 0,
		Priority:   int32(config.Priority),
		CreatedAt:  config.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  config.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	
	if config.IdcLocation.Valid {
		idcConf.IdcLocation = config.IdcLocation.String
	}
	
	if config.IdcDescription.Valid {
		idcConf.IdcDescription = config.IdcDescription.String
	}

	return &cmpool.MatchIdcByIpResp{
		Success: true,
		Message: "匹配成功",
		IdcConf: idcConf,
	}, nil
}
