package logic

import (
	"errors"
	"strings"

	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"cmdb-api/cmpool"

	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
)

type GetScheduledTaskExecutionDetailsLogic struct {
	logx.Logger
	r      *http.Request
	svcCtx *svc.ServiceContext
}

func NewGetScheduledTaskExecutionDetailsLogic(r *http.Request, svcCtx *svc.ServiceContext) *GetScheduledTaskExecutionDetailsLogic {
	return &GetScheduledTaskExecutionDetailsLogic{
		Logger: logx.WithContext(r.Context()),
		r:      r,
		svcCtx: svcCtx,
	}
}

func (l *GetScheduledTaskExecutionDetailsLogic) GetScheduledTaskExecutionDetails() (resp *types.GetScheduledTaskExecutionDetailsResponse, err error) {
	executionTaskId := l.r.URL.Query().Get("execution_task_id")
	if executionTaskId == "" {
		// 从URL路径获取参数
		path := l.r.URL.Path
		parts := strings.Split(path, "/")
		if len(parts) > 0 {
			executionTaskId = parts[len(parts)-1]
		}
	}
	
	l.Logger.Infof("获取执行详情，URL: %s, 提取的execution_task_id: %s", l.r.URL.Path, executionTaskId)
	
	if executionTaskId == "" {
		return nil, errors.New("执行任务ID不能为空")
	}

	// 调用RPC服务获取执行详情
	result, err := l.svcCtx.CmpoolRpc.GetScheduledTaskExecutionDetails(l.r.Context(), &cmpool.GetScheduledTaskExecutionDetailsReq{
		ExecutionTaskId: executionTaskId,
	})
	if err != nil {
		l.Logger.Errorf("获取执行任务详情失败: %v", err)
		return nil, err
	}

	if !result.Success {
		return &types.GetScheduledTaskExecutionDetailsResponse{
			Success: false,
			Message: result.Message,
		}, nil
	}

	// 转换数据结构
	var hostDetails []types.ScheduledTaskExecutionDetail
	for _, detail := range result.Data.HostDetails {
		hostDetails = append(hostDetails, types.ScheduledTaskExecutionDetail{
			HostIp:          detail.HostIp,
			ResourceType:    detail.ResourceType,
			TargetPercent:   int(detail.TargetPercent),
			Duration:        int(detail.Duration),
			ExecutionStatus: detail.ExecutionStatus,
			StartTime:       detail.StartTime,
			EndTime:         detail.EndTime,
			ExitCode:        int(detail.ExitCode),
			StdoutLog:       detail.StdoutLog,
			StderrLog:       detail.StderrLog,
			ResultSummary:   detail.ResultSummary,
			SSHError:        detail.SSHError,
			CreateTime:      detail.CreateTime,
		})
	}

	return &types.GetScheduledTaskExecutionDetailsResponse{
		Success: true,
		Message: "查询成功",
		Data: types.ScheduledTaskExecutionInfo{
			ScheduledTaskId: int(result.Data.ScheduledTaskId),
			TaskName:        result.Data.TaskName,
			ExecutionTaskId: result.Data.ExecutionTaskId,
			ExecutionTime:   result.Data.ExecutionTime,
			ExecutionStatus: result.Data.ExecutionStatus,
			TotalHosts:      int(result.Data.TotalHosts),
			SuccessHosts:    int(result.Data.SuccessHosts),
			FailedHosts:     int(result.Data.FailedHosts),
			ResourceType:    result.Data.ResourceType,
			TargetPercent:   int(result.Data.TargetPercent),
			Duration:        int(result.Data.Duration),
			HostDetails:     hostDetails,
		},
	}, nil
}
