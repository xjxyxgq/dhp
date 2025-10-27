package logic

import (
	"context"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetIdcConfListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetIdcConfListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetIdcConfListLogic {
	return &GetIdcConfListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// IDC机房配置管理相关方法
func (l *GetIdcConfListLogic) GetIdcConfList(in *cmpool.GetIdcConfListReq) (*cmpool.GetIdcConfListResp, error) {
	// 从数据库获取IDC配置列表
	configs, err := l.svcCtx.IdcConfModel.FindAllByPriority(l.ctx, in.ActiveOnly)
	if err != nil {
		l.Logger.Errorf("查询IDC配置列表失败: %v", err)
		return &cmpool.GetIdcConfListResp{
			Success: false,
			Message: "查询IDC配置列表失败",
		}, nil
	}

	// 转换为protobuf格式
	var idcConfList []*cmpool.IdcConf
	for _, config := range configs {
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
		
		idcConfList = append(idcConfList, idcConf)
	}

	return &cmpool.GetIdcConfListResp{
		Success:     true,
		Message:     "查询成功",
		IdcConfList: idcConfList,
	}, nil
}
