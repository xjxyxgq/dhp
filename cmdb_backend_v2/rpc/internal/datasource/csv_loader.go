package datasource

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"cmdb-rpc/internal/model"

	"database/sql"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CSVLoader struct {
	conn                sqlx.SqlConn
	serverResourceModel model.ServerResourcesModel
}

func NewCSVLoader(conn sqlx.SqlConn) *CSVLoader {
	return &CSVLoader{
		conn:                conn,
		serverResourceModel: model.NewServerResourcesModel(conn),
	}
}

// LoadServerMetrics 从CSV文件加载服务器监控指标数据
func (l *CSVLoader) LoadServerMetrics(csvFilePath string) error {
	logx.Infof("开始从CSV文件加载服务器监控指标数据: %s", csvFilePath)

	file, err := os.Open(csvFilePath)
	if err != nil {
		return fmt.Errorf("打开CSV文件失败: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("读取CSV文件失败: %v", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("CSV文件为空")
	}

	// 跳过标题行
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 5 {
			logx.Errorf("第%d行数据不完整，跳过: %v", i+1, record)
			continue
		}

		// 解析CSV数据
		hostIP := record[0]
		hostName := record[1]

		maxCpu, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			logx.Errorf("第%d行MaxCpu解析失败，跳过: %v", i+1, err)
			continue
		}

		maxMem, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			logx.Errorf("第%d行MaxMem解析失败，跳过: %v", i+1, err)
			continue
		}

		maxDisk, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			logx.Errorf("第%d行MaxDisk解析失败，跳过: %v", i+1, err)
			continue
		}

		// 创建服务器资源记录
		serverResource := &model.ServerResources{
			PoolId:         1, // 默认主机池ID，后续可以根据需要调整
			Ip:             sql.NullString{String: hostIP, Valid: true},
			CpuPercentMax:  sql.NullFloat64{Float64: maxCpu, Valid: true},
			MemPercentMax:  sql.NullFloat64{Float64: maxMem, Valid: true},
			DiskPercentMax: sql.NullFloat64{Float64: maxDisk, Valid: true},
			MonDate:        sql.NullTime{Time: time.Now(), Valid: true},
		}

		// 插入数据库
		_, err = l.serverResourceModel.Insert(context.Background(), serverResource)
		if err != nil {
			logx.Errorf("插入服务器资源数据失败: %v", err)
			continue
		}

		logx.Infof("成功加载服务器监控数据: %s (%s)", hostName, hostIP)
	}

	logx.Infof("CSV文件加载完成，共处理%d条记录", len(records)-1)
	return nil
}
