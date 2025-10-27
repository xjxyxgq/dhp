package logic

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"
	"database/sql"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoadServerMetricsFromCsvLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoadServerMetricsFromCsvLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadServerMetricsFromCsvLogic {
	return &LoadServerMetricsFromCsvLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoadServerMetricsFromCsvLogic) LoadServerMetricsFromCsv(in *cmpool.LoadServerMetricsCSVReq, stream cmpool.Cmpool_LoadServerMetricsFromCsvServer) error {
	l.Logger.Infof("开始从CSV加载服务器监控指标数据, 文件名: %s", in.Filename)

	// 创建一个独立的上下文，不会因为客户端连接关闭而取消
	dbCtx := context.Background()

	// 检查流上下文是否已经被取消
	if stream.Context().Err() != nil {
		errorMsg := fmt.Sprintf("请求已被取消: %v", stream.Context().Err())
		l.Logger.Errorf("无法开始处理: %s", errorMsg)
		return fmt.Errorf(errorMsg)
	}

	// 创建一个CSV读取器
	reader := csv.NewReader(bytes.NewReader(in.FileContent))

	// 读取并验证标题行
	headers, err := reader.Read()
	if err != nil {
		l.Logger.Errorf("读取CSV标题行失败: %v", err)
		err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
			Success:         false,
			Message:         "读取CSV标题行失败",
			ErrorDetail:     err.Error(),
			IsCompleted:     true,
			LastUpdatedTime: time.Now().Format(time.RFC3339),
		})
		if err != nil {
			l.Logger.Errorf("发送错误信息失败: %v", err)
		}
		return fmt.Errorf("读取CSV标题行失败: %v", err)
	}

	// 验证必要的列是否存在
	requiredColumns := map[string]bool{
		"hostIP":   false,
		"hostName": false,
		"MaxCpu":   false,
		"MaxMem":   false,
		"MaxDisk":  false,
	}

	for _, header := range headers {
		if _, exists := requiredColumns[header]; exists {
			requiredColumns[header] = true
		}
	}

	// 检查所有必需列是否都存在
	missingColumns := []string{}
	for col, found := range requiredColumns {
		if !found {
			missingColumns = append(missingColumns, col)
		}
	}

	if len(missingColumns) > 0 {
		errorMsg := fmt.Sprintf("CSV文件缺少必要的列: %v", missingColumns)
		l.Logger.Errorf("CSV格式错误: %s", errorMsg)
		err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
			Success:         false,
			Message:         "CSV格式错误",
			ErrorDetail:     errorMsg,
			IsCompleted:     true,
			LastUpdatedTime: time.Now().Format(time.RFC3339),
		})
		if err != nil {
			l.Logger.Errorf("发送错误信息失败: %v", err)
		}
		return fmt.Errorf("CSV格式错误: %s", errorMsg)
	}

	// 获取各列的索引
	hostIPIndex := l.getColumnIndex(headers, "hostIP")
	hostNameIndex := l.getColumnIndex(headers, "hostName")
	maxCpuIndex := l.getColumnIndex(headers, "MaxCpu")
	maxMemIndex := l.getColumnIndex(headers, "MaxMem")
	maxDiskIndex := l.getColumnIndex(headers, "MaxDisk")

	// 读取所有行以获取总行数
	allRecords, err := reader.ReadAll()
	if err != nil {
		l.Logger.Errorf("读取CSV数据失败: %v", err)
		err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
			Success:         false,
			Message:         "读取CSV数据失败",
			ErrorDetail:     err.Error(),
			IsCompleted:     true,
			LastUpdatedTime: time.Now().Format(time.RFC3339),
		})
		if err != nil {
			l.Logger.Errorf("发送错误信息失败: %v", err)
		}
		return fmt.Errorf("读取CSV数据失败: %v", err)
	}

	totalRows := len(allRecords)
	if totalRows == 0 {
		l.Logger.Errorf("CSV文件不包含数据行")
		err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
			Success:         false,
			Message:         "CSV文件不包含数据行",
			ErrorDetail:     "没有可导入的数据",
			IsCompleted:     true,
			LastUpdatedTime: time.Now().Format(time.RFC3339),
		})
		if err != nil {
			l.Logger.Errorf("发送错误信息失败: %v", err)
		}
		return fmt.Errorf("CSV文件不包含数据行")
	}

	// 检查流上下文是否已经被取消
	if stream.Context().Err() != nil {
		errorMsg := fmt.Sprintf("处理被客户端取消: %v", stream.Context().Err())
		l.Logger.Errorf("%s，已读取CSV文件但尚未开始处理", errorMsg)
		return fmt.Errorf(errorMsg)
	}

	// 发送初始进度
	err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
		Success:            true,
		Message:            "开始处理数据",
		TotalRows:          int32(totalRows),
		ProcessedRows:      0,
		ProgressPercentage: 0,
		IsCompleted:        false,
		LastUpdatedTime:    time.Now().Format(time.RFC3339),
	})
	if err != nil {
		l.Logger.Errorf("发送进度信息失败: %v", err)
		return err
	}

	// 统计信息
	processedRows := 0
	skippedRows := 0

	// 处理每一行数据
	for i, record := range allRecords {
		// 检查当前上下文是否已取消
		if stream.Context().Err() != nil {
			errorMsg := fmt.Sprintf("处理被客户端取消: %v", stream.Context().Err())
			l.Logger.Errorf("%s，已处理 %d/%d 行", errorMsg, i, totalRows)
			return fmt.Errorf(errorMsg)
		}

		if len(record) <= maxDiskIndex || len(record) <= maxMemIndex || len(record) <= maxCpuIndex ||
			len(record) <= hostIPIndex || len(record) <= hostNameIndex {
			errorMsg := fmt.Sprintf("行 %d 的数据列数不足", i+1)
			l.Logger.Errorf("数据行格式错误: %s", errorMsg)
			err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
				Success:         false,
				Message:         "数据行格式错误",
				ErrorDetail:     "数据列数不足",
				ErrorLine:       int32(i + 1),
				IsCompleted:     true,
				LastUpdatedTime: time.Now().Format(time.RFC3339),
			})
			if err != nil {
				l.Logger.Errorf("发送错误信息失败: %v", err)
			}
			return fmt.Errorf("数据行格式错误: %s", errorMsg)
		}

		hostIP := record[hostIPIndex]
		hostName := record[hostNameIndex]
		// 如果主机名为空，使用 IP 作为主机名
		if hostName == "" {
			hostName = hostIP
		}

		// 检查主机IP是否存在于hosts_pool表中
		hostInfo, err := l.svcCtx.HostsPoolModel.FindOneByHostIp(dbCtx, hostIP)
		if err != nil {
			// 如果主机不存在，跳过此行
			if err == model.ErrNotFound {
				l.Logger.Infof("主机IP %s 不在hosts_pool表中，跳过此行", hostIP)
				skippedRows++
				continue
			}

			// 其他数据库错误
			errorMsg := fmt.Sprintf("查询主机信息失败: %v, IP: %s", err, hostIP)
			l.Logger.Errorf("%s", errorMsg)
			err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
				Success:         false,
				Message:         "查询主机信息失败",
				ErrorDetail:     errorMsg,
				ErrorLine:       int32(i + 1),
				IsCompleted:     true,
				LastUpdatedTime: time.Now().Format(time.RFC3339),
			})
			if err != nil {
				l.Logger.Errorf("发送错误信息失败: %v", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		// 解析 CPU 值
		maxCpu, err := strconv.ParseFloat(record[maxCpuIndex], 64)
		if err != nil {
			errorMsg := fmt.Sprintf("行 %d 的 MaxCpu 值无法解析: %v", i+1, err)
			l.Logger.Errorf("%s", errorMsg)
			err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
				Success:         false,
				Message:         "MaxCpu 解析失败",
				ErrorDetail:     err.Error(),
				ErrorLine:       int32(i + 1),
				IsCompleted:     true,
				LastUpdatedTime: time.Now().Format(time.RFC3339),
			})
			if err != nil {
				l.Logger.Errorf("发送错误信息失败: %v", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		// 解析内存值
		maxMem, err := strconv.ParseFloat(record[maxMemIndex], 64)
		if err != nil {
			errorMsg := fmt.Sprintf("行 %d 的 MaxMem 值无法解析: %v", i+1, err)
			l.Logger.Errorf("%s", errorMsg)
			err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
				Success:         false,
				Message:         "MaxMem 解析失败",
				ErrorDetail:     err.Error(),
				ErrorLine:       int32(i + 1),
				IsCompleted:     true,
				LastUpdatedTime: time.Now().Format(time.RFC3339),
			})
			if err != nil {
				l.Logger.Errorf("发送错误信息失败: %v", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		// 解析磁盘值
		maxDisk, err := strconv.ParseFloat(record[maxDiskIndex], 64)
		if err != nil {
			errorMsg := fmt.Sprintf("行 %d 的 MaxDisk 值无法解析: %v", i+1, err)
			l.Logger.Errorf("%s", errorMsg)
			err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
				Success:         false,
				Message:         "MaxDisk 解析失败",
				ErrorDetail:     err.Error(),
				ErrorLine:       int32(i + 1),
				IsCompleted:     true,
				LastUpdatedTime: time.Now().Format(time.RFC3339),
			})
			if err != nil {
				l.Logger.Errorf("发送错误信息失败: %v", err)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		// 创建服务器资源记录
		// CSV文件中的 MaxCpu, MaxMem, MaxDisk 为百分比值
		// 如果CSV没有avg/min列，使用max值填充
		serverResource := &model.ServerResources{
			PoolId: hostInfo.Id,                                 // 使用hosts_pool表中的ID
			Ip:     sql.NullString{String: hostIP, Valid: true}, // 主机IP

			// 百分比字段：CSV只提供max值，使用max填充avg和min
			CpuPercentMax: sql.NullFloat64{Float64: maxCpu, Valid: true},   // CPU百分比最大值
			CpuPercentAvg: sql.NullFloat64{Float64: maxCpu, Valid: true},   // CPU百分比平均值（用max填充）
			CpuPercentMin: sql.NullFloat64{Float64: maxCpu, Valid: true},   // CPU百分比最小值（用max填充）
			MemPercentMax: sql.NullFloat64{Float64: maxMem, Valid: true},   // 内存百分比最大值
			MemPercentAvg: sql.NullFloat64{Float64: maxMem, Valid: true},   // 内存百分比平均值（用max填充）
			MemPercentMin: sql.NullFloat64{Float64: maxMem, Valid: true},   // 内存百分比最小值（用max填充）
			DiskPercentMax: sql.NullFloat64{Float64: maxDisk, Valid: true}, // 磁盘百分比最大值
			DiskPercentAvg: sql.NullFloat64{Float64: maxDisk, Valid: true}, // 磁盘百分比平均值（用max填充）
			DiskPercentMin: sql.NullFloat64{Float64: maxDisk, Valid: true}, // 磁盘百分比最小值（用max填充）

			MonDate: sql.NullTime{Time: time.Now(), Valid: true}, // 当前日期
		}

		// 使用独立的上下文进行数据库插入操作
		_, err = l.svcCtx.ServerResourcesModel.Insert(dbCtx, serverResource)
		if err != nil {
			errorMsg := fmt.Sprintf("行 %d 插入数据失败: %v", i+1, err)
			l.Logger.Errorf("%s", errorMsg)

			// 检查是否为上下文取消错误
			if err.Error() == "context canceled" {
				errorMsg = fmt.Sprintf("数据库操作被中断，可能是系统重启或服务关闭导致，行号: %d", i+1)
				l.Logger.Errorf("%s", errorMsg)
			}

			// 检查流上下文是否已被取消
			if stream.Context().Err() != nil {
				l.Logger.Errorf("无法发送错误信息：客户端连接已关闭，原始错误: %v", err)
				return fmt.Errorf("客户端连接已关闭，原始错误: %v", err)
			}

			// 尝试发送错误，但不要因为发送失败而阻止函数返回
			sendErr := stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
				Success:         false,
				Message:         "插入数据失败",
				ErrorDetail:     errorMsg,
				ErrorLine:       int32(i + 1),
				IsCompleted:     true,
				LastUpdatedTime: time.Now().Format(time.RFC3339),
			})
			if sendErr != nil {
				l.Logger.Errorf("发送错误信息失败: %v", sendErr)
			}
			return fmt.Errorf("%s", errorMsg)
		}

		l.Logger.Infof("成功加载服务器监控数据: %s (%s)", hostName, hostIP)
		processedRows++

		// 每处理一定数量的行发送一次进度更新
		if (i+1)%10 == 0 || i == len(allRecords)-1 {
			progress := float32(i+1) / float32(totalRows) * 100

			// 检查是否可以发送进度
			if stream.Context().Err() != nil {
				l.Logger.Infof("无法发送进度更新：客户端连接已关闭: %v", stream.Context().Err())
				return fmt.Errorf("客户端连接已关闭: %v", stream.Context().Err())
			}

			err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
				Success:            true,
				Message:            fmt.Sprintf("已处理 %d/%d 行, 已跳过 %d 行", i+1, totalRows, skippedRows),
				TotalRows:          int32(totalRows),
				ProcessedRows:      int32(i + 1),
				ProgressPercentage: progress,
				IsCompleted:        false,
				LastUpdatedTime:    time.Now().Format(time.RFC3339),
			})
			if err != nil {
				l.Logger.Errorf("发送进度信息失败: %v", err)
				return err
			}
		}
	}

	// 发送完成信息
	if stream.Context().Err() != nil {
		l.Logger.Infof("无法发送完成信息：客户端连接已关闭: %v", stream.Context().Err())
		return fmt.Errorf("客户端连接已关闭: %v", stream.Context().Err())
	}

	// 构建完成消息，包含导入和跳过的行数
	completionMessage := fmt.Sprintf("数据处理完成，共导入 %d 条记录，跳过 %d 条记录（主机IP不存在）", processedRows, skippedRows)

	err = stream.Send(&cmpool.LoadServerMetricsCSVProgressResp{
		Success:            true,
		Message:            completionMessage,
		TotalRows:          int32(totalRows),
		ProcessedRows:      int32(processedRows),
		ProgressPercentage: 100,
		IsCompleted:        true,
		LastUpdatedTime:    time.Now().Format(time.RFC3339),
	})
	if err != nil {
		l.Logger.Errorf("发送完成信息失败: %v", err)
		return err
	}

	l.Logger.Infof("CSV数据加载完成，共处理 %d 条记录，跳过 %d 条记录", processedRows, skippedRows)
	return nil
}

// 获取列索引
func (l *LoadServerMetricsFromCsvLogic) getColumnIndex(headers []string, columnName string) int {
	for i, header := range headers {
		if header == columnName {
			return i
		}
	}
	return -1
}
