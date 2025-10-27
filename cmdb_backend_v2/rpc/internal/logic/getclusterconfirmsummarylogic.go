package logic

import (
	"context"
	"encoding/json"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetClusterConfirmSummaryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetClusterConfirmSummaryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetClusterConfirmSummaryLogic {
	return &GetClusterConfirmSummaryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetClusterConfirmSummaryLogic) GetClusterConfirmSummary(in *cmpool.ClusterConfirmSummaryReq) (*cmpool.ClusterConfirmSummaryResp, error) {
	l.Logger.Info("开始生成集群确认摘要")

	// 设置默认天数
	days := int32(7)
	if in.Days > 0 {
		days = in.Days
	}

	// 查询最新的插件执行记录
	dbRecords, err := l.svcCtx.PluginExecutionRecordsModel.FindRecentRecords(l.ctx, days)
	if err != nil {
		l.Logger.Errorf("查询插件执行记录失败: %v", err)
		return &cmpool.ClusterConfirmSummaryResp{
			Success: false,
			Message: "查询失败: " + err.Error(),
		}, nil
	}

	pluginResults := make(map[string]interface{})
	checkSequences := make(map[string]bool)

	for _, record := range dbRecords {
		checkSequences[record.CheckSeq] = true

		// 解析JSON结果
		var resultData map[string]interface{}
		if err := json.Unmarshal([]byte(record.Result), &resultData); err != nil {
			l.Logger.Errorf("解析插件结果JSON失败: %v", err)
			resultData = map[string]interface{}{
				"raw_result":  record.Result,
				"parse_error": err.Error(),
			}
		}

		// 按插件类型分组结果
		if pluginResults[record.PluginName] == nil {
			pluginResults[record.PluginName] = make([]map[string]interface{}, 0)
		}

		if pluginList, ok := pluginResults[record.PluginName].([]map[string]interface{}); ok {
			pluginResults[record.PluginName] = append(pluginList, map[string]interface{}{
				"check_seq": record.CheckSeq,
				"result":    resultData,
			})
		}
	}

	// 生成摘要统计
	summary := map[string]interface{}{
		"total_checks":       len(checkSequences),
		"plugin_types":       len(pluginResults),
		"report_time":        fmt.Sprintf("最近%d天", days),
		"summary_statistics": l.generateSummaryStats(pluginResults),
	}

	pluginResults["summary"] = summary

	// 将结果转换为JSON字符串
	pluginResultsJSON, err := json.Marshal(pluginResults)
	if err != nil {
		l.Logger.Errorf("序列化插件结果失败: %v", err)
		return &cmpool.ClusterConfirmSummaryResp{
			Success: false,
			Message: "序列化失败: " + err.Error(),
		}, nil
	}

	// 生成报告文件URL
	reportFileURL := "/api/reports/cluster-summary-" + l.generateReportID() + ".json"

	l.Logger.Infof("成功生成集群确认摘要，包含%d种插件类型的结果", len(pluginResults)-1)
	return &cmpool.ClusterConfirmSummaryResp{
		Success: true,
		Message: "生成成功",
		ClusterConfirmSummary: &cmpool.ClusterConfirmSummary{
			ReportFileURL: reportFileURL,
			PluginResults: string(pluginResultsJSON),
		},
	}, nil
}

// generateSummaryStats 生成摘要统计信息
func (l *GetClusterConfirmSummaryLogic) generateSummaryStats(pluginResults map[string]interface{}) map[string]interface{} {
	stats := map[string]interface{}{
		"backup_checks": map[string]int{
			"total":   0,
			"success": 0,
			"failed":  0,
		},
		"performance_monitors": map[string]int{
			"total":   0,
			"normal":  0,
			"warning": 0,
		},
		"security_audits": map[string]int{
			"total":  0,
			"passed": 0,
			"failed": 0,
		},
	}

	// 统计各类插件的执行结果
	for pluginName, results := range pluginResults {
		if resultList, ok := results.([]map[string]interface{}); ok {
			switch pluginName {
			case "mysql_backup_checker", "mssql_backup_checker":
				backupStats := stats["backup_checks"].(map[string]int)
				backupStats["total"] += len(resultList)
				for _, result := range resultList {
					if resultData, ok := result["result"].(map[string]interface{}); ok {
						if status, exists := resultData["status"]; exists && status == "success" {
							backupStats["success"]++
						} else {
							backupStats["failed"]++
						}
					}
				}
			case "performance_monitor":
				perfStats := stats["performance_monitors"].(map[string]int)
				perfStats["total"] += len(resultList)
				for _, result := range resultList {
					if resultData, ok := result["result"].(map[string]interface{}); ok {
						if diskIO, exists := resultData["disk_io"]; exists && diskIO == "normal" {
							perfStats["normal"]++
						} else {
							perfStats["warning"]++
						}
					}
				}
			case "security_audit":
				secStats := stats["security_audits"].(map[string]int)
				secStats["total"] += len(resultList)
				for _, result := range resultList {
					if resultData, ok := result["result"].(map[string]interface{}); ok {
						if status, exists := resultData["status"]; exists && status == "passed" {
							secStats["passed"]++
						} else {
							secStats["failed"]++
						}
					}
				}
			}
		}
	}

	return stats
}

// generateReportID 生成报告ID
func (l *GetClusterConfirmSummaryLogic) generateReportID() string {
	return "20240109-cluster-summary"
}
