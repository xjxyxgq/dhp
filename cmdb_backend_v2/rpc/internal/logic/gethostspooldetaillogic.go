package logic

import (
	"context"
	"database/sql"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetHostsPoolDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetHostsPoolDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetHostsPoolDetailLogic {
	return &GetHostsPoolDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询主机资源池详情，包括主机的硬件信息、软件信息，接口涉及数据量较大，一般不用于页面显示，用于向外部提供全量数据或数据导出需求
func (l *GetHostsPoolDetailLogic) GetHostsPoolDetail(in *cmpool.GetHostsPoolDetailReq) (*cmpool.GetHostsPoolDetailResp, error) {
	l.Logger.Infof("查询主机池详情，IP列表: %v", in.IpList)

	// 从数据库查询主机池详细信息
	hostsPoolDetail, err := l.getHostsPoolDetailFromDB(in.IpList)
	if err != nil {
		l.Logger.Errorf("查询主机池详情失败: %v", err)
		return &cmpool.GetHostsPoolDetailResp{
			Success: false,
			Message: fmt.Sprintf("查询主机池详情失败: %v", err),
		}, nil
	}

	l.Logger.Infof("查询到%d台主机的详细信息", len(hostsPoolDetail))
	return &cmpool.GetHostsPoolDetailResp{
		Success:         true,
		Message:         "查询成功",
		HostsPoolDetail: hostsPoolDetail,
	}, nil
}

// getHostsPoolDetailFromDB 从数据库查询主机池详细信息
func (l *GetHostsPoolDetailLogic) getHostsPoolDetailFromDB(ipList []string) ([]*cmpool.HostPoolDetail, error) {
	// 使用模型层方法查询主机信息
	hostRows, err := l.svcCtx.HostsPoolModel.FindHostsPoolDetailWithFilter(l.ctx, ipList)
	if err != nil {
		return nil, fmt.Errorf("查询主机信息失败: %w", err)
	}

	var hostsDetail []*cmpool.HostPoolDetail
	hostIdToIpMap := make(map[int64]string) // 用于后续查询应用信息

	// 转换主机信息
	for _, row := range hostRows {
		host := &cmpool.HostPoolDetail{
			Id:              row.Id,
			Hostname:        row.HostName,
			HostIp:          row.HostIp,
			Disk:            row.DiskSize,
			Ram:             row.Ram,
			VCpu:            row.Vcpus,
			IfH3CSync:       row.IfH3cSync,
			H3CImgId:        row.H3cImgId,
			H3CHmName:       row.H3cHmName,
			LeafNumber:      row.LeafNumber,
			RackNumber:      row.RackNumber,
			RackHeight:      row.RackHeight,
			RackStartNumber: row.RackStartNumber,
			FromFactor:      row.FromFactor,
			SerialNumber:    row.SerialNumber,
			Remark:          row.Remark,
			IsDeleted:       row.IsDeleted,
			IsStatic:        row.IsStatic,
			CreateTime:      row.CreateTime,
			UpdateTime:      row.UpdateTime,
			AppList:         []*cmpool.App{}, // 初始化应用列表
		}

		// 处理可为空的字段
		if row.HostType != nil {
			host.HostType = *row.HostType
		}
		if row.H3cId != nil {
			host.H3CId = *row.H3cId
		}
		if row.H3cStatus != nil {
			host.H3CStatus = *row.H3cStatus
		}

		// 查询并设置IDC信息
		l.loadIdcInfoForHost(host)

		hostsDetail = append(hostsDetail, host)
		hostIdToIpMap[host.Id] = host.HostIp
	}

	// 查询每台主机的应用信息
	if len(hostsDetail) > 0 {
		// todo. 一台台主机加载应用信息，可能会导致突发的qps上升和接口响应缓慢，后续应当优化为关联查询
		err = l.loadApplicationsForHosts(hostsDetail, hostIdToIpMap)
		if err != nil {
			l.Logger.Errorf("加载主机应用信息失败: %v", err)
			// 不返回错误，继续返回主机基本信息
		}
	}

	return hostsDetail, nil
}

// loadApplicationsForHosts 为主机加载应用信息
func (l *GetHostsPoolDetailLogic) loadApplicationsForHosts(hostsDetail []*cmpool.HostPoolDetail, hostIdToIpMap map[int64]string) error {
	if len(hostsDetail) == 0 {
		return nil
	}

	// 构建主机ID列表
	hostIds := make([]string, 0, len(hostsDetail))
	hostDetailMap := make(map[int64]*cmpool.HostPoolDetail)

	for _, host := range hostsDetail {
		hostIds = append(hostIds, fmt.Sprintf("%d", host.Id))
		hostDetailMap[host.Id] = host
	}

	// 使用模型层方法查询应用信息
	appRows, err := l.svcCtx.HostsApplicationsModel.FindByPoolIds(l.ctx, hostIds)
	if err != nil {
		return fmt.Errorf("查询应用信息失败: %w", err)
	}

	// 将应用信息添加到对应的主机中
	for _, row := range appRows {
		app := &cmpool.App{
			Aid: row.Id, // 设置应用ID
		}

		// 处理可为空的字段
		if row.ServerType.Valid {
			app.ServerType = row.ServerType.String
		}
		if row.ServerVersion.Valid {
			app.ServerVersion = row.ServerVersion.String
		}
		if row.ServerSubtitle.Valid {
			app.ServerSubtitle = row.ServerSubtitle.String
		}
		if row.ClusterName.Valid {
			app.ClusterName = row.ClusterName.String
		}
		if row.ServerProtocol.Valid {
			app.ServiceProtocol = row.ServerProtocol.String
		}
		if row.ServerAddr.Valid {
			app.ServiceAddr = row.ServerAddr.String
		}
		if row.ServerPort.Valid {
			app.ServiceAddr = fmt.Sprintf(":%d", row.ServerPort.Int32)
		}
		if row.ServerRole.Valid {
			app.ServiceRole = row.ServerRole.String
		}
		if row.DepartmentName.Valid {
			app.DepartmentName = row.DepartmentName.String
		}

		// 将应用信息添加到对应的主机中
		if host, exists := hostDetailMap[row.PoolId]; exists {
			host.AppList = append(host.AppList, app)
		}
	}

	return nil
}

// loadIdcInfoForHost 为主机加载IDC信息
func (l *GetHostsPoolDetailLogic) loadIdcInfoForHost(host *cmpool.HostPoolDetail) {
	// 通过IP地址匹配IDC信息
	idcConf, err := l.svcCtx.IdcConfModel.MatchIdcByIp(l.ctx, host.HostIp)
	if err != nil {
		if err != sql.ErrNoRows {
			l.Logger.Errorf("匹配主机%s的IDC信息失败: %v", host.HostIp, err)
		}
		return
	}

	// 填充IDC信息
	host.IdcInfo = &cmpool.IdcConf{
		Id:          int64(idcConf.Id),
		IdcName:     idcConf.IdcName,
		IdcCode:     idcConf.IdcCode,
		IdcIpRegexp: idcConf.IdcIpRegexp,
		IsActive:    idcConf.IsActive > 0,
		Priority:    int32(idcConf.Priority),
		CreatedAt:   idcConf.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   idcConf.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if idcConf.IdcLocation.Valid {
		host.IdcInfo.IdcLocation = idcConf.IdcLocation.String
	}

	if idcConf.IdcDescription.Valid {
		host.IdcInfo.IdcDescription = idcConf.IdcDescription.String
	}
}
