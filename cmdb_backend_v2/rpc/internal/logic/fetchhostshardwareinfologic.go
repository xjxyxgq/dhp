package logic

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/datasource"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"
)

type FetchHostsHardwareInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFetchHostsHardwareInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FetchHostsHardwareInfoLogic {
	return &FetchHostsHardwareInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FetchHostsHardwareInfoLogic) FetchHostsHardwareInfo(in *cmpool.FetchHostsHardwareInfoReq) (*cmpool.FetchHostsHardwareInfoResp, error) {
	// 获取配置
	cmdbUrl := l.svcCtx.Config.ExternalAPI.CmdbUrl
	appCode := l.svcCtx.Config.ExternalAPI.CmdbAppCode
	secret := l.svcCtx.Config.ExternalAPI.CmdbSecret
	isFreak := l.svcCtx.Config.ExternalAPI.Freak

	// 如果不是模拟模式，检查CMDB配置
	if !isFreak && (cmdbUrl == "" || appCode == "" || secret == "") {
		logx.Error("CMDB配置信息不完整")
		return &cmpool.FetchHostsHardwareInfoResp{
			Success: false,
			Message: "CMDB配置信息不完整",
		}, nil
	}

	// 初始化CMDB客户端
	cmdbClient := datasource.NewCmdb(cmdbUrl, appCode, secret)

	// 如果不是模拟模式，刷新Token
	//if !isFreak {
	//	err := cmdbClient.RefreshToken()
	//	if err != nil {
	//		logx.Errorf("刷新CMDB Token失败: %v", err)
	//		return &cmpool.FetchHostsHardwareInfoResp{
	//			Success: false,
	//			Message: fmt.Sprintf("刷新CMDB Token失败: %v", err),
	//		}, nil
	//	}
	//}

	var hostIpList []string
	if len(in.HostIpList) > 0 {
		// 使用传入的IP列表
		hostIpList = in.HostIpList
	} else {
		// 获取hosts_pool表中所有主机IP
		var err error
		hostIpList, err = l.svcCtx.HostsPoolModel.FindAllHostIPs(l.ctx)
		if err != nil {
			logx.Errorf("查询hosts_pool表失败: %v", err)
			return &cmpool.FetchHostsHardwareInfoResp{
				Success: false,
				Message: fmt.Sprintf("查询hosts_pool表失败: %v", err),
			}, nil
		}
	}

	if len(hostIpList) == 0 {
		return &cmpool.FetchHostsHardwareInfoResp{
			Success:    true,
			Message:    "没有需要更新的主机",
			TotalHosts: 0,
		}, nil
	}

	if isFreak {
		logx.Infof("模拟模式：开始获取%d台主机的硬件信息", len(hostIpList))
	} else {
		logx.Infof("开始获取%d台主机的硬件信息", len(hostIpList))
	}

	var hardwareInfoList []*cmpool.HostHardwareInfo
	var updatedCount, failedCount int32
	batchSize := 50 // 每批次处理的主机数量

	// 分批处理主机列表
	for i := 0; i < len(hostIpList); i += batchSize {
		end := i + batchSize
		if end > len(hostIpList) {
			end = len(hostIpList)
		}

		batch := hostIpList[i:end]
		logx.Infof("处理第%d批次，主机数量: %d", i/batchSize+1, len(batch))

		// 从CMDB获取硬件信息（传递isFreak参数）
		hostDataList, err := cmdbClient.GetDataByHostIpList(batch, batchSize, isFreak)
		if err != nil {
			logx.Errorf("从CMDB获取硬件信息失败: %v", err)
			// 对于失败的批次，将所有主机标记为失败
			for _, ip := range batch {
				hardwareInfoList = append(hardwareInfoList, &cmpool.HostHardwareInfo{
					HostIp:  ip,
					Success: false,
					Message: fmt.Sprintf("获取CMDB数据失败: %v", err),
				})
				failedCount++
			}
			continue
		}

		// 处理获取到的硬件信息
		for _, hostData := range hostDataList {
			hardwareInfo := &cmpool.HostHardwareInfo{
				HostIp:   hostData.HostIp,
				HostName: hostData.HostName,
				Success:  true,
				Message:  "成功获取硬件信息",
			}

			// 提取硬件信息
			if hostData.HostExtInfo != nil {
				hardwareInfo.Disk = hostData.HostExtInfo.Disk
				hardwareInfo.Ram = hostData.HostExtInfo.Ram
				hardwareInfo.Vcpus = int64(hostData.HostExtInfo.Vcpus)
			}

			// 更新hosts_pool表
			var diskSize, ram, vcpus int32
			if hostData.HostExtInfo != nil {
				diskSize = int32(hostData.HostExtInfo.Disk)
				ram = int32(hostData.HostExtInfo.Ram)
				vcpus = int32(hostData.HostExtInfo.Vcpus)
			}
			err := l.svcCtx.HostsPoolModel.UpdateHostHardwareInfo(l.ctx, &model.HostsPool{
				HostIp:   hostData.HostIp,
				HostName: hostData.HostName,
				DiskSize: sql.NullInt64{Int64: int64(diskSize), Valid: diskSize > 0},
				Ram:      sql.NullInt64{Int64: int64(ram), Valid: ram > 0},
				Vcpus:    sql.NullInt64{Int64: int64(vcpus), Valid: vcpus > 0},
			})
			if err != nil {
				logx.Errorf("更新主机%s硬件信息失败: %v", hostData.HostIp, err)
				hardwareInfo.Success = false
				hardwareInfo.Message = fmt.Sprintf("更新数据库失败: %v", err)
				failedCount++
			} else {
				// 硬件信息更新成功后，尝试更新IDC信息
				l.updateHostIdcInfo(hostData.HostIp)
				updatedCount++
			}

			hardwareInfoList = append(hardwareInfoList, hardwareInfo)
		}

		// 检查是否有主机在CMDB中找不到
		cmdbHostIps := make(map[string]bool)
		for _, hostData := range hostDataList {
			cmdbHostIps[hostData.HostIp] = true
		}

		for _, ip := range batch {
			if !cmdbHostIps[ip] {
				hardwareInfoList = append(hardwareInfoList, &cmpool.HostHardwareInfo{
					HostIp:  ip,
					Success: false,
					Message: "在CMDB中未找到该主机",
				})
				failedCount++
			}
		}
	}

	totalHosts := int32(len(hostIpList))
	modeStr := ""
	if isFreak {
		modeStr = "（模拟模式）"
	}
	logx.Infof("硬件信息获取完成%s，总计: %d, 成功: %d, 失败: %d", modeStr, totalHosts, updatedCount, failedCount)

	return &cmpool.FetchHostsHardwareInfoResp{
		Success:          true,
		Message:          fmt.Sprintf("硬件信息获取完成%s，总计: %d, 成功: %d, 失败: %d", modeStr, totalHosts, updatedCount, failedCount),
		TotalHosts:       totalHosts,
		UpdatedHosts:     updatedCount,
		FailedHosts:      failedCount,
		HardwareInfoList: hardwareInfoList,
	}, nil
}

// updateHostIdcInfo 更新主机的IDC信息
func (l *FetchHostsHardwareInfoLogic) updateHostIdcInfo(hostIp string) {
	// 匹配IDC配置
	idcConf, err := l.svcCtx.IdcConfModel.MatchIdcByIp(l.ctx, hostIp)
	if err != nil {
		if err == sql.ErrNoRows {
			logx.Infof("主机%s未匹配到IDC配置", hostIp)
		} else {
			logx.Errorf("匹配主机%s的IDC配置失败: %v", hostIp, err)
		}
		return
	}

	// 更新主机的IDC信息
	err = l.svcCtx.HostsPoolModel.UpdateHostIdcInfo(l.ctx, hostIp, int64(idcConf.Id))
	if err != nil {
		logx.Errorf("更新主机%s的IDC信息失败: %v", hostIp, err)
	} else {
		logx.Infof("成功更新主机%s的IDC信息为%s", hostIp, idcConf.IdcName)
	}
}
