package logic

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDiskPredictionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDiskPredictionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDiskPredictionLogic {
	return &GetDiskPredictionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取磁盘预测数据（优化版本：使用关联查询）
func (l *GetDiskPredictionLogic) GetDiskPrediction(in *cmpool.DiskPredictionReq) (*cmpool.DiskPredictionResp, error) {
	// 设置默认时间范围（最近3个月）
	endTime := in.EndTime
	beginTime := in.BeginTime

	if endTime == "" {
		endTime = time.Now().Format("2006-01-02 15:04:05")
	}

	if beginTime == "" {
		// 默认3个月前
		threemonthsAgo := time.Now().AddDate(0, -3, 0)
		beginTime = threemonthsAgo.Format("2006-01-02 15:04:05")
	}

	// 构建查询条件
	whereClause := "sr.is_deleted = 0 AND sr.mon_date BETWEEN ? AND ?"
	var args []interface{}
	args = append(args, beginTime, endTime)

	// 添加IP列表查询条件
	if len(in.IpList) > 0 {
		placeholders := make([]string, len(in.IpList))
		for i, ip := range in.IpList {
			placeholders[i] = "?"
			args = append(args, ip)
		}
		whereClause += " AND sr.ip IN (" + strings.Join(placeholders, ",") + ")"
	}

	// 添加集群查询条件
	if in.Cluster != "" {
		whereClause += " AND ha.cluster_name = ?"
		args = append(args, in.Cluster)
	}

	// 移除未使用的查询代码，现在使用 model 层方法

	// 使用统一的过滤器函数查询磁盘预测数据
	filter := &model.DiskPredictionFilter{
		BeginTime:      beginTime,
		EndTime:        endTime,
		IPList:         in.IpList,
		ClusterName:    in.Cluster,
		DepartmentName: in.DepartmentName,
	}

	diskResults, err := l.svcCtx.ServerResourcesModel.FindDiskPredictionDataWithFilter(l.ctx, filter)

	if err != nil {
		l.Logger.Errorf("Failed to query disk data: %v", err)
		return &cmpool.DiskPredictionResp{
			Success: false,
			Message: fmt.Sprintf("查询磁盘数据失败: %v", err),
		}, nil
	}

	// 按IP分组处理数据并收集集群信息
	ipDataMap := make(map[string][]*model.DiskPredictionData)
	ipClusterMap := make(map[string]map[string]*cmpool.HostClusterInfo) // IP -> clusters map

	for _, diskData := range diskResults {
		ipDataMap[diskData.Ip] = append(ipDataMap[diskData.Ip], diskData)

		// 收集集群信息
		if diskData.ClusterName != "" {
			if ipClusterMap[diskData.Ip] == nil {
				ipClusterMap[diskData.Ip] = make(map[string]*cmpool.HostClusterInfo)
			}
			clusterKey := diskData.ClusterName + "|" + diskData.GroupName
			if ipClusterMap[diskData.Ip][clusterKey] == nil {
				ipClusterMap[diskData.Ip][clusterKey] = &cmpool.HostClusterInfo{
					ClusterName:      diskData.ClusterName,
					ClusterGroupName: diskData.GroupName,
					DepartmentName:   diskData.DepartmentName,
				}
			}
		}
	}

	var predictions []*cmpool.DiskPrediction

	for ip, dataList := range ipDataMap {
		var prediction *cmpool.DiskPrediction

		if len(dataList) < 2 {
			// 数据点太少，无法预测，但仍显示当前状态
			if len(dataList) == 1 {
				prediction = l.createBasicPrediction(dataList[0])
			}
		} else {
			// 计算磁盘增长率
			prediction = l.calculateDiskPrediction(dataList)
		}

		if prediction != nil {
			// 添加集群信息
			if clusters, exists := ipClusterMap[ip]; exists {
				for _, cluster := range clusters {
					prediction.Clusters = append(prediction.Clusters, cluster)
				}
			}
			predictions = append(predictions, prediction)
		}
	}

	return &cmpool.DiskPredictionResp{
		Success:        true,
		Message:        "磁盘预测数据获取成功",
		DiskPrediction: predictions,
	}, nil
}

// 创建基础预测（仅有当前状态，无增长预测）
func (l *GetDiskPredictionLogic) createBasicPrediction(data *model.DiskPredictionData) *cmpool.DiskPrediction {
	// 直接使用数据库中的百分比字段
	currentUsagePercent := data.DiskPercentMax

	now := time.Now()
	prediction := &cmpool.DiskPrediction{
		Id:                      data.ID,
		Ip:                      data.Ip,
		CurrentDiskUsagePercent: currentUsagePercent,
		TotalDisk:               data.TotalDisk,
		UsedDisk:                data.UsedDisk,
		DailyGrowthRate:         0,
		PredictedFullDate:       "数据不足，无法预测",
		DaysUntilFull:           -1,
		IsHighRisk:              false,
		CreateAt:                now.Format("2006-01-02 15:04:05"),
		UpdateAt:                now.Format("2006-01-02 15:04:05"),
	}

	// 填充 IDC 信息
	if data.IdcId.Valid && data.IdcId.Int64 > 0 {
		prediction.IdcInfo = &cmpool.IdcConf{
			Id:             data.IdcId.Int64,
			IdcName:        data.IdcName.String,
			IdcCode:        data.IdcCode.String,
			IdcLocation:    data.IdcLocation.String,
			IdcDescription: data.IdcDescription.String,
		}
	}

	return prediction
}

func (l *GetDiskPredictionLogic) calculateDiskPrediction(dataList []*model.DiskPredictionData) *cmpool.DiskPrediction {
	if len(dataList) < 2 {
		return nil
	}

	// 获取最新的数据点
	latestData := dataList[len(dataList)-1]

	// 直接使用数据库中的百分比字段
	currentUsagePercent := latestData.DiskPercentMax

	// 计算磁盘使用百分比增长率（%/天）
	dailyGrowthRate := l.calculateGrowthRate(dataList)

	// 预测磁盘满的时间（基于百分比）
	predictedDate, daysUntilFull := l.predictFullDate(latestData, dailyGrowthRate)

	// 判断是否高风险（30天内磁盘满）
	isHighRisk := daysUntilFull <= 30 && daysUntilFull > 0

	now := time.Now()
	prediction := &cmpool.DiskPrediction{
		Id:                      latestData.ID,
		Ip:                      latestData.Ip,
		CurrentDiskUsagePercent: currentUsagePercent,
		TotalDisk:               latestData.TotalDisk,
		UsedDisk:                latestData.UsedDisk,
		DailyGrowthRate:         dailyGrowthRate,
		PredictedFullDate:       predictedDate,
		DaysUntilFull:           int32(daysUntilFull),
		IsHighRisk:              isHighRisk,
		CreateAt:                now.Format("2006-01-02 15:04:05"),
		UpdateAt:                now.Format("2006-01-02 15:04:05"),
	}

	// 填充 IDC 信息
	if latestData.IdcId.Valid && latestData.IdcId.Int64 > 0 {
		prediction.IdcInfo = &cmpool.IdcConf{
			Id:             latestData.IdcId.Int64,
			IdcName:        latestData.IdcName.String,
			IdcCode:        latestData.IdcCode.String,
			IdcLocation:    latestData.IdcLocation.String,
			IdcDescription: latestData.IdcDescription.String,
		}
	}

	return prediction
}

// 计算每日磁盘使用百分比增长率（%/天）
func (l *GetDiskPredictionLogic) calculateGrowthRate(dataList []*model.DiskPredictionData) float32 {
	if len(dataList) < 2 {
		return 0
	}

	// 使用简单的线性回归计算趋势
	var sumX, sumY, sumXY, sumX2 float64
	n := float64(len(dataList))

	for i, data := range dataList {
		x := float64(i) // 时间索引
		y := float64(data.DiskPercentMax) // 使用百分比字段

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// 计算斜率（每个时间单位的百分比增长量）
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0
	}

	slope := (n*sumXY - sumX*sumY) / denominator

	// 将斜率转换为每日百分比增长率
	timeSpan := l.calculateTimeSpan(dataList)
	if timeSpan <= 0 {
		return 0
	}

	dailyGrowthRate := float32(slope / timeSpan)

	return dailyGrowthRate
}

// 计算数据时间跨度（天）
func (l *GetDiskPredictionLogic) calculateTimeSpan(dataList []*model.DiskPredictionData) float64 {
	if len(dataList) < 2 {
		return 0
	}

	firstTime, err1 := time.Parse("2006-01-02", dataList[0].MonDate)
	lastTime, err2 := time.Parse("2006-01-02", dataList[len(dataList)-1].MonDate)

	if err1 != nil || err2 != nil {
		// 如果时间解析失败，假设数据点之间间隔1天
		return float64(len(dataList) - 1)
	}

	duration := lastTime.Sub(firstTime)
	days := duration.Hours() / 24

	if days <= 0 {
		return 1 // 至少1天
	}

	return days
}

// 预测磁盘满的时间（基于百分比）
func (l *GetDiskPredictionLogic) predictFullDate(latestData *model.DiskPredictionData, dailyGrowthRate float32) (string, int) {
	if dailyGrowthRate <= 0 {
		return time.Now().AddDate(100, 0, 0).Format("2006-01-02"), 0
	}

	// 计算当前使用百分比
	currentPercent := latestData.DiskPercentMax

	// 计算剩余百分比（距离100%）
	remainingPercent := 100.0 - currentPercent

	if remainingPercent <= 0 {
		return time.Now().Format("2006-01-02"), 0
	}

	// 计算达到100%需要的天数（基于百分比增长率）
	daysUntilFull := float64(remainingPercent) / float64(dailyGrowthRate)

	if daysUntilFull < 0 {
		return time.Now().Format("2006-01-02"), 0
	}

	if math.IsInf(daysUntilFull, 1) || daysUntilFull > 3650 { // 超过10年
		return time.Now().AddDate(100, 0, 0).Format("2006-01-02"), 0
	}

	// 计算预测日期
	predictedDate := time.Now().AddDate(0, 0, int(daysUntilFull))

	return predictedDate.Format("2006-01-02"), int(daysUntilFull)
}
