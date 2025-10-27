package logic

import (
	"context"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyMonitoringDataLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewVerifyMonitoringDataLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyMonitoringDataLogic {
	return &VerifyMonitoringDataLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 监控数据核对
func (l *VerifyMonitoringDataLogic) VerifyMonitoringData(in *cmpool.VerifyMonitoringDataReq) (*cmpool.VerifyMonitoringDataResp, error) {
	// 验证输入参数
	if in.StartTime == "" || in.EndTime == "" {
		return &cmpool.VerifyMonitoringDataResp{
			Success: false,
			Message: "时间范围不能为空",
		}, nil
	}

	// 获取hosts_pool表中的所有主机
	allHosts, err := l.svcCtx.HostsPoolModel.FindAllActiveHosts(l.ctx)
	if err != nil {
		l.Logger.Errorf("查询hosts_pool表失败: %v", err)
		return &cmpool.VerifyMonitoringDataResp{
			Success: false,
			Message: fmt.Sprintf("查询主机列表失败: %v", err),
		}, nil
	}

	totalHosts := len(allHosts)
	if totalHosts == 0 {
		return &cmpool.VerifyMonitoringDataResp{
			Success:                    true,
			Message:                    "没有找到主机数据",
			TotalHosts:                 0,
			HostsWithMonitoring:        0,
			HostsWithoutMonitoring:     0,
			MonitoringCoverage:         0.0,
			HostsWithoutMonitoringList: []*cmpool.HostWithoutMonitoring{},
		}, nil
	}

	// 查询在指定时间范围内有监控数据的主机
	monitoringIPsList, err := l.svcCtx.ServerResourcesModel.FindDistinctIPsInTimeRange(l.ctx, in.StartTime, in.EndTime)
	if err != nil {
		l.Logger.Errorf("查询监控数据失败: %v", err)
		return &cmpool.VerifyMonitoringDataResp{
			Success: false,
			Message: fmt.Sprintf("查询监控数据失败: %v", err),
		}, nil
	}

	// 创建一个map来存储有监控数据的主机IP
	hostsWithMonitoring := make(map[string]bool)
	for _, monitoringIP := range monitoringIPsList {
		hostsWithMonitoring[monitoringIP.IP] = true
	}

	// 找出没有监控数据的主机
	var hostsWithoutMonitoringList []*cmpool.HostWithoutMonitoring
	hostsWithMonitoringCount := 0

	for _, host := range allHosts {
		if hostsWithMonitoring[host.HostIp] {
			hostsWithMonitoringCount++
		} else {
			hostsWithoutMonitoringList = append(hostsWithoutMonitoringList, &cmpool.HostWithoutMonitoring{
				HostIp:     host.HostIp,
				HostName:   host.HostName,
				PoolName:   host.PoolName,
				CreateTime: host.CreateTime,
			})
		}
	}

	hostsWithoutMonitoringCount := totalHosts - hostsWithMonitoringCount
	monitoringCoverage := float32(hostsWithMonitoringCount) / float32(totalHosts)

	// 构造响应消息
	message := fmt.Sprintf("监控数据核对完成。总主机数: %d，有监控数据: %d，无监控数据: %d，监控覆盖率: %.1f%%",
		totalHosts, hostsWithMonitoringCount, hostsWithoutMonitoringCount, monitoringCoverage*100)

	return &cmpool.VerifyMonitoringDataResp{
		Success:                    true,
		Message:                    message,
		TotalHosts:                 int32(totalHosts),
		HostsWithMonitoring:        int32(hostsWithMonitoringCount),
		HostsWithoutMonitoring:     int32(hostsWithoutMonitoringCount),
		MonitoringCoverage:         monitoringCoverage,
		HostsWithoutMonitoringList: hostsWithoutMonitoringList,
	}, nil
}
