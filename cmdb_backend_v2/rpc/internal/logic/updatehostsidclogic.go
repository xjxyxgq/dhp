package logic

import (
	"context"
	"database/sql"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateHostsIdcLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateHostsIdcLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateHostsIdcLogic {
	return &UpdateHostsIdcLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量更新主机IDC信息
func (l *UpdateHostsIdcLogic) UpdateHostsIdc(in *cmpool.UpdateHostsIdcReq) (*cmpool.UpdateHostsIdcResp, error) {
	// 获取需要更新IDC信息的主机列表
	hosts, err := l.svcCtx.HostsPoolModel.FindAllHostsForIdcUpdate(l.ctx, in.HostIpList)
	if err != nil {
		l.Logger.Errorf("查询主机列表失败: %v", err)
		return &cmpool.UpdateHostsIdcResp{
			Success: false,
			Message: "查询主机列表失败",
		}, nil
	}

	totalHosts := len(hosts)
	updatedHosts := 0
	unmatchedHosts := 0

	// 遍历每个主机，尝试匹配IDC并更新
	for _, host := range hosts {
		// 匹配IDC配置
		config, err := l.svcCtx.IdcConfModel.MatchIdcByIp(l.ctx, host.HostIp)
		if err != nil {
			if err == sql.ErrNoRows {
				// 没有匹配到IDC配置
				unmatchedHosts++
				l.Logger.Infof("主机 %s 未匹配到IDC配置", host.HostIp)
				continue
			}
			
			l.Logger.Errorf("匹配主机 %s 的IDC配置失败: %v", host.HostIp, err)
			continue
		}

		// 更新主机的IDC信息
		err = l.svcCtx.HostsPoolModel.UpdateHostIdcInfo(l.ctx, host.HostIp, int64(config.Id))
		if err != nil {
			l.Logger.Errorf("更新主机 %s 的IDC信息失败: %v", host.HostIp, err)
			continue
		}

		updatedHosts++
		l.Logger.Infof("成功更新主机 %s 的IDC信息为 %s", host.HostIp, config.IdcName)
	}

	message := ""
	if totalHosts > 0 {
		message = fmt.Sprintf("处理完成，共处理%d台主机，成功更新%d台，未匹配%d台", 
			totalHosts, updatedHosts, unmatchedHosts)
	} else {
		message = "没有找到需要更新的主机"
	}

	return &cmpool.UpdateHostsIdcResp{
		Success:        true,
		Message:        message,
		TotalHosts:     int32(totalHosts),
		UpdatedHosts:   int32(updatedHosts),
		UnmatchedHosts: int32(unmatchedHosts),
	}, nil
}
