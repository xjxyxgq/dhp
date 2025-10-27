package logic

import (
	"context"
	"time"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetServerMetricsCSVProgressLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetServerMetricsCSVProgressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetServerMetricsCSVProgressLogic {
	return &GetServerMetricsCSVProgressLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetServerMetricsCSVProgressLogic) GetServerMetricsCSVProgress() (*types.LoadServerMetricsCSVProgressResponse, error) {
	// 加锁获取当前进度
	progressMutex.Lock()
	defer progressMutex.Unlock()

	// 检查最后一次更新时间，如果超过一定时间没有更新且任务未完成，认为处理已经出现问题
	if csvProgress.LastUpdatedTime != "" {
		// 解析最后更新时间
		if lastUpdated, err := time.Parse(time.RFC3339, csvProgress.LastUpdatedTime); err == nil {
			// 如果任务未完成且超过30秒没有更新，判断为超时
			if !csvProgress.IsCompleted && time.Since(lastUpdated) > 30*time.Second {
				l.Logger.Errorf("CSV处理似乎已停止: 最后更新时间是 %v，已超过30秒", lastUpdated)

				// 更新为错误状态
				csvProgress.Success = false
				csvProgress.Message = "CSV处理可能已停止响应"
				csvProgress.ErrorDetail = "服务器超过30秒未更新处理进度，请检查服务器日志或重新上传"
				csvProgress.IsCompleted = true
			}
		}
	}

	// 复制一份当前进度信息并返回
	progress := &types.LoadServerMetricsCSVProgressResponse{
		Success:            csvProgress.Success,
		Message:            csvProgress.Message,
		TotalRows:          csvProgress.TotalRows,
		ProcessedRows:      csvProgress.ProcessedRows,
		ProgressPercentage: csvProgress.ProgressPercentage,
		IsCompleted:        csvProgress.IsCompleted,
		ErrorDetail:        csvProgress.ErrorDetail,
		ErrorLine:          csvProgress.ErrorLine,
	}

	return progress, nil
}
