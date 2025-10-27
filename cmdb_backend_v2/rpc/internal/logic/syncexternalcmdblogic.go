package logic

import (
	"context"
	"database/sql"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncExternalCmdbLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncExternalCmdbLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncExternalCmdbLogic {
	return &SyncExternalCmdbLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 从外部CMDB同步完整主机信息到hosts_pool表
func (l *SyncExternalCmdbLogic) SyncExternalCmdb(in *cmpool.SyncExternalCmdbReq) (*cmpool.SyncExternalCmdbResp, error) {
	l.Logger.Infof("开始从外部CMDB同步主机数据，PageSize: %d, HostOwner: %d, ForceUpdate: %v",
		in.PageSize, in.HostOwner, in.ForceUpdate)

	// 设置默认页大小
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 100
	}

	var totalHosts, syncedHosts, updatedHosts, failedHosts int32
	var syncResults []*cmpool.ExternalCmdbHost

	// 调用外部CMDB接口获取主机数据
	externalHosts, totalPages, err := l.svcCtx.ExternalAPI.FetchHostsFromExternalCMDB(int(pageSize), int(in.HostOwner))
	if err != nil {
		l.Logger.Errorf("从外部CMDB获取主机数据失败: %v", err)
		return &cmpool.SyncExternalCmdbResp{
			Success: false,
			Message: fmt.Sprintf("从外部CMDB获取主机数据失败: %v", err),
		}, nil
	}

	totalHosts = int32(len(externalHosts))
	l.Logger.Infof("从外部CMDB获取到 %d 条主机数据", totalHosts)

	// 逐条处理主机数据
	for _, host := range externalHosts {
		// 转换外部CMDB数据到本地模型
		hostPool, hostApp, syncResult := l.convertExternalHostToLocal(host)
		syncResults = append(syncResults, syncResult)

		// 检查主机是否已存在
		existingHost, err := l.svcCtx.HostsPoolModel.FindOneByHostIp(l.ctx, hostPool.HostIp)
		if err != nil && err != model.ErrNotFound {
			l.Logger.Errorf("查询主机 %s 失败: %v", host.HostIp, err)
			failedHosts++
			syncResult.Success = false
			syncResult.Message = fmt.Sprintf("查询主机失败: %v", err)
			continue
		}

		if existingHost != nil {
			// 主机已存在
			if in.ForceUpdate {
				// 强制更新
				hostPool.Id = existingHost.Id
				err = l.svcCtx.HostsPoolModel.Update(l.ctx, hostPool)
				if err != nil {
					l.Logger.Errorf("更新主机 %s 失败: %v", host.HostIp, err)
					failedHosts++
					syncResult.Success = false
					syncResult.Message = fmt.Sprintf("更新主机失败: %v", err)
					continue
				}
				updatedHosts++
				syncResult.Message = "主机信息已更新"
				l.Logger.Infof("成功更新主机: %s (%s)", host.HostName, host.HostIp)
			} else {
				// 跳过已存在的主机
				syncResult.Message = "主机已存在，跳过"
				l.Logger.Infof("主机已存在，跳过: %s (%s)", host.HostName, host.HostIp)
				continue
			}
		} else {
			// 插入新主机
			result, err := l.svcCtx.HostsPoolModel.Insert(l.ctx, hostPool)
			if err != nil {
				l.Logger.Errorf("插入主机 %s 失败: %v", host.HostIp, err)
				failedHosts++
				syncResult.Success = false
				syncResult.Message = fmt.Sprintf("插入主机失败: %v", err)
				continue
			}

			// 获取插入的主机ID
			hostId, err := result.LastInsertId()
			if err != nil {
				l.Logger.Errorf("获取主机ID失败: %v", err)
			} else {
				// 如果有应用信息，插入应用部署信息
				if hostApp != nil && hostApp.ServerType.Valid {
					hostApp.PoolId = uint64(hostId)
					_, err = l.svcCtx.HostsApplicationsModel.Insert(l.ctx, hostApp)
					if err != nil {
						l.Logger.Errorf("插入主机应用信息失败: %v", err)
					}
				}
			}

			syncedHosts++
			syncResult.Message = "主机信息已同步"
			l.Logger.Infof("成功同步新主机: %s (%s)", host.HostName, host.HostIp)
		}
	}

	successMsg := fmt.Sprintf("外部CMDB同步完成。总数: %d, 新增: %d, 更新: %d, 失败: %d",
		totalHosts, syncedHosts, updatedHosts, failedHosts)

	l.Logger.Info(successMsg)

	return &cmpool.SyncExternalCmdbResp{
		Success:       true,
		Message:       successMsg,
		TotalPages:    int32(totalPages),
		ProcessedPages: 1, // 目前处理一页
		TotalHosts:    totalHosts,
		SyncedHosts:   syncedHosts,
		UpdatedHosts:  updatedHosts,
		FailedHosts:   failedHosts,
		SyncResults:   syncResults,
	}, nil
}

// convertExternalHostToLocal 将外部CMDB主机数据转换为本地模型
func (l *SyncExternalCmdbLogic) convertExternalHostToLocal(host *cmpool.ExternalCmdbHost) (*model.HostsPool, *model.HostsApplications, *cmpool.ExternalCmdbHost) {
	// 创建hosts_pool记录
	hostPool := &model.HostsPool{
		HostName: host.HostName,
		HostIp:   host.HostIp,
		HostType: sql.NullString{
			String: host.HostType,
			Valid:  host.HostType != "",
		},
		H3cId: sql.NullString{
			String: host.H3CId,
			Valid:  host.H3CId != "",
		},
		H3cStatus: sql.NullString{
			String: host.H3CStatus,
			Valid:  host.H3CStatus != "",
		},
		DiskSize: sql.NullInt64{
			Int64: host.Disk,
			Valid: host.Disk > 0,
		},
		Ram: sql.NullInt64{
			Int64: host.Ram / 1024, // 转换MB到GB
			Valid: host.Ram > 0,
		},
		Vcpus: sql.NullInt64{
			Int64: host.Vcpus,
			Valid: host.Vcpus > 0,
		},
		IfH3cSync: sql.NullString{
			String: host.IfH3CSync,
			Valid:  host.IfH3CSync != "",
		},
		H3cImgId: sql.NullString{
			String: host.H3CImageId,
			Valid:  host.H3CImageId != "",
		},
		H3cHmName: sql.NullString{
			String: host.H3CHmName,
			Valid:  host.H3CHmName != "",
		},
		IsDeleted: 0, // 新同步的主机默认未删除
		IsStatic:  0, // 新同步的主机默认非静态
	}

	// 创建hosts_applications记录（如果有应用信息）
	var hostApp *model.HostsApplications
	if host.AppName != "" {
		hostApp = &model.HostsApplications{
			ServerType: sql.NullString{
				String: "other", // 默认类型，可根据实际需要调整
				Valid:  true,
			},
			ClusterName: sql.NullString{
				String: host.AppName,
				Valid:  true,
			},
			ServerAddr: sql.NullString{
				String: host.HostIp,
				Valid:  true,
			},
			ServerPort:     8080, // 默认端口，可根据实际需要调整
			DepartmentName: sql.NullString{
				String: host.BizGroup,
				Valid:  host.BizGroup != "",
			},
		}
	}

	// 创建同步结果记录
	syncResult := &cmpool.ExternalCmdbHost{
		CmdbId:      host.CmdbId,
		DomainNum:   host.DomainNum,
		HostName:    host.HostName,
		HostIp:      host.HostIp,
		HostType:    host.HostType,
		HostOwner:   host.HostOwner,
		OpsIamCode:  host.OpsIamCode,
		OwnerGroup:  host.OwnerGroup,
		OwnerIamCode: host.OwnerIamCode,
		H3CId:       host.H3CId,
		H3CStatus:   host.H3CStatus,
		Disk:        host.Disk,
		Ram:         host.Ram,
		Vcpus:       host.Vcpus,
		CreatedAt:   host.CreatedAt,
		UpdatedAt:   host.UpdatedAt,
		IfH3CSync:   host.IfH3CSync,
		H3CImageId:  host.H3CImageId,
		H3CHmName:   host.H3CHmName,
		AppName:     host.AppName,
		DataSource:  host.DataSource,
		BizGroup:    host.BizGroup,
		OpsBizGroup: host.OpsBizGroup,
		Message:     "", // 初始化为空，后面会设置
		Success:     true, // 默认成功，失败时会设置为false
	}

	return hostPool, hostApp, syncResult
}
