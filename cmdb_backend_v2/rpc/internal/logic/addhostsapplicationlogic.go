package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddHostsApplicationLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddHostsApplicationLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddHostsApplicationLogic {
	return &AddHostsApplicationLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 手动添加资源池主机应用信息
func (l *AddHostsApplicationLogic) AddHostsApplication(in *cmpool.AddHostsAppReq) (*cmpool.AddHostsAppResp, error) {
	if len(in.Data) == 0 {
		return &cmpool.AddHostsAppResp{
			Success: false,
			Message: "应用数据不能为空",
		}, nil
	}

	successCount := 0
	var errors []string

	for _, app := range in.Data {
		// 验证必要字段
		if app.HostId <= 0 {
			errors = append(errors, fmt.Sprintf("主机ID无效: %d", app.HostId))
			continue
		}

		if app.ServerType == "" {
			errors = append(errors, "服务类型不能为空")
			continue
		}

		// 使用 UpsertApplication 方法插入或更新应用信息
		err := l.svcCtx.HostsApplicationsModel.UpsertApplication(
			l.ctx,
			int64(app.HostId),
			app.ServerType,
			app.ServerVersion,
			app.ClusterName,
			app.ServerProtocol,
			app.ServerAddr,
			app.ServerRole,
			app.ServerStatus,
			app.DepartmentName,
			app.ServerPort,
		)

		if err != nil {
			l.Errorf("添加应用失败，主机ID: %d, 服务类型: %s, 错误: %v", app.HostId, app.ServerType, err)
			errors = append(errors, fmt.Sprintf("主机ID %d 添加失败: %v", app.HostId, err))
		} else {
			successCount++
		}
	}

	// 构造响应消息
	message := fmt.Sprintf("成功添加 %d 个应用", successCount)
	if len(errors) > 0 {
		message += fmt.Sprintf("，失败 %d 个", len(errors))
	}

	return &cmpool.AddHostsAppResp{
		Success: len(errors) == 0,
		Message: message,
	}, nil
}
