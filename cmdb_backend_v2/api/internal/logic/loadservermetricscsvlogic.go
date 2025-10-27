package logic

import (
	"context"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"sync"
	"time"

	"cmdb-api/cmpool"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// 用于存储CSV处理进度的全局变量
var (
	progressMutex sync.Mutex
	csvProgress   *types.LoadServerMetricsCSVProgressResponse
)

func init() {
	// 初始化进度信息
	csvProgress = &types.LoadServerMetricsCSVProgressResponse{
		Success:            true,
		Message:            "暂无处理任务",
		TotalRows:          0,
		ProcessedRows:      0,
		ProgressPercentage: 0,
		IsCompleted:        true,
	}
}

type LoadServerMetricsCSVLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoadServerMetricsCSVLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoadServerMetricsCSVLogic {
	return &LoadServerMetricsCSVLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoadServerMetricsCSVLogic) LoadServerMetricsCSV(req *types.LoadServerMetricsCSVRequest, file multipart.File, header *multipart.FileHeader) (*types.LoadServerMetricsCSVProgressResponse, error) {
	l.Logger.Infof("开始处理CSV文件上传: %s, 大小: %d字节", header.Filename, header.Size)

	// 重置进度信息
	progressMutex.Lock()
	now := time.Now()
	csvProgress = &types.LoadServerMetricsCSVProgressResponse{
		Success:            true,
		Message:            "正在准备处理数据...",
		TotalRows:          0,
		ProcessedRows:      0,
		ProgressPercentage: 0,
		IsCompleted:        false,
		LastUpdatedTime:    now.Format(time.RFC3339),
	}
	progressMutex.Unlock()

	// 读取上传的文件内容
	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		l.Logger.Errorf("读取文件内容失败: %v", err)
		progressMutex.Lock()
		now := time.Now()
		csvProgress = &types.LoadServerMetricsCSVProgressResponse{
			Success:         false,
			Message:         "读取文件内容失败",
			ErrorDetail:     err.Error(),
			IsCompleted:     true,
			LastUpdatedTime: now.Format(time.RFC3339),
		}
		progressMutex.Unlock()
		return nil, fmt.Errorf("读取文件内容失败: %v", err)
	}

	// 创建一个独立的上下文，不受API请求超时的影响
	// 这样即使API请求结束，RPC调用也能继续执行
	rpcCtx := context.Background()

	// 调用RPC服务处理CSV文件
	client := l.svcCtx.CmpoolRpc
	rpcReq := &cmpool.LoadServerMetricsCSVReq{
		FileContent: fileContent,
		Filename:    header.Filename,
	}

	// 使用流式RPC调用，注意方法名使用Csv而不是CSV
	// 使用独立的上下文，防止API请求结束导致RPC调用被取消
	stream, err := client.LoadServerMetricsFromCsv(rpcCtx, rpcReq)
	if err != nil {
		l.Logger.Errorf("调用RPC服务失败: %v", err)
		progressMutex.Lock()
		now := time.Now()
		csvProgress = &types.LoadServerMetricsCSVProgressResponse{
			Success:         false,
			Message:         "服务调用失败",
			ErrorDetail:     err.Error(),
			IsCompleted:     true,
			LastUpdatedTime: now.Format(time.RFC3339),
		}
		progressMutex.Unlock()
		return nil, fmt.Errorf("调用RPC服务失败: %v", err)
	}

	// 获取第一个响应以处理可能的初始错误
	initialProgress, err := stream.Recv()
	if err != nil {
		l.Logger.Errorf("接收初始进度信息失败: %v", err)
		progressMutex.Lock()
		now := time.Now()
		csvProgress = &types.LoadServerMetricsCSVProgressResponse{
			Success:         false,
			Message:         "接收进度信息失败",
			ErrorDetail:     err.Error(),
			IsCompleted:     true,
			LastUpdatedTime: now.Format(time.RFC3339),
		}
		progressMutex.Unlock()
		return nil, fmt.Errorf("接收进度信息失败: %v", err)
	}

	// 检查初始响应是否包含错误
	if !initialProgress.Success {
		l.Logger.Errorf("CSV处理初始错误: %s, 详情: %s", initialProgress.Message, initialProgress.ErrorDetail)
		progressMutex.Lock()
		now := time.Now()
		csvProgress = &types.LoadServerMetricsCSVProgressResponse{
			Success:         false,
			Message:         initialProgress.Message,
			ErrorDetail:     initialProgress.ErrorDetail,
			ErrorLine:       int(initialProgress.ErrorLine),
			IsCompleted:     true,
			LastUpdatedTime: now.Format(time.RFC3339),
		}
		progressMutex.Unlock()
		return nil, fmt.Errorf("CSV处理错误: %s, 详情: %s", initialProgress.Message, initialProgress.ErrorDetail)
	}

	// 更新进度信息
	progressMutex.Lock()
	now = time.Now()
	csvProgress.Message = initialProgress.Message
	csvProgress.TotalRows = int(initialProgress.TotalRows)
	csvProgress.ProcessedRows = int(initialProgress.ProcessedRows)
	csvProgress.ProgressPercentage = initialProgress.ProgressPercentage
	csvProgress.LastUpdatedTime = now.Format(time.RFC3339)
	progressMutex.Unlock()

	// 创建一个带超时的上下文，设置5分钟超时
	ctx, cancel := context.WithTimeout(rpcCtx, 5*time.Minute)

	// 启动后台协程处理后续的流式响应
	go func() {
		defer cancel() // 确保协程结束时取消上下文

		// 创建一个计时器用于超时检测
		timeoutTimer := time.NewTimer(5 * time.Minute)
		defer timeoutTimer.Stop()

		for {
			// 使用select来处理超时和接收数据
			select {
			case <-ctx.Done():
				// 上下文被取消（超时或者其他原因）
				if ctx.Err() == context.DeadlineExceeded {
					// 超时情况
					l.Logger.Error("CSV处理超时，已超过5分钟限制")
					progressMutex.Lock()
					now := time.Now()
					csvProgress.Success = false
					csvProgress.Message = "处理超时"
					csvProgress.ErrorDetail = "CSV处理操作已超过5分钟限制，自动终止"
					csvProgress.IsCompleted = true
					csvProgress.LastUpdatedTime = now.Format(time.RFC3339)
					progressMutex.Unlock()
				}
				return

			default:
				// 尝试接收流数据
				progress, err := stream.Recv()
				if err != nil {
					l.Logger.Errorf("接收进度信息失败: %v", err)
					progressMutex.Lock()
					now := time.Now()
					csvProgress.Success = false
					csvProgress.Message = "接收进度信息失败"
					csvProgress.ErrorDetail = err.Error()
					csvProgress.IsCompleted = true
					csvProgress.LastUpdatedTime = now.Format(time.RFC3339)
					progressMutex.Unlock()
					return
				}

				// 重置超时计时器
				if !timeoutTimer.Stop() {
					<-timeoutTimer.C
				}
				timeoutTimer.Reset(5 * time.Minute)

				// 更新进度信息
				progressMutex.Lock()
				now := time.Now()
				csvProgress.Success = progress.Success
				csvProgress.Message = progress.Message
				csvProgress.TotalRows = int(progress.TotalRows)
				csvProgress.ProcessedRows = int(progress.ProcessedRows)
				csvProgress.ProgressPercentage = progress.ProgressPercentage
				csvProgress.IsCompleted = progress.IsCompleted
				csvProgress.LastUpdatedTime = now.Format(time.RFC3339)

				if !progress.Success {
					csvProgress.ErrorDetail = progress.ErrorDetail
					csvProgress.ErrorLine = int(progress.ErrorLine)
				}
				progressMutex.Unlock()

				if progress.IsCompleted {
					l.Logger.Infof("CSV处理完成: %s", progress.Message)
					return
				}
			}
		}
	}()

	// 返回初始进度信息
	return &types.LoadServerMetricsCSVProgressResponse{
		Success:            true,
		Message:            "文件上传成功，开始处理数据",
		ProgressPercentage: initialProgress.ProgressPercentage,
		TotalRows:          int(initialProgress.TotalRows),
		ProcessedRows:      int(initialProgress.ProcessedRows),
		IsCompleted:        false,
	}, nil
}
