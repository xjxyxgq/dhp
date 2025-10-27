package logic

import (
	"context"
	"database/sql"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateEsSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateEsSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateEsSyncTaskLogic {
	return &CreateEsSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateEsSyncTask ES数据同步任务配置管理相关方法
func (l *CreateEsSyncTaskLogic) CreateEsSyncTask(in *cmpool.CreateESSyncTaskReq) (*cmpool.CreateESSyncTaskResp, error) {
	// 1. 验证参数
	if in.TaskName == "" {
		return &cmpool.CreateESSyncTaskResp{
			Success: false,
			Message: "任务名称不能为空",
		}, nil
	}

	if in.CronExpression == "" {
		return &cmpool.CreateESSyncTaskResp{
			Success: false,
			Message: "Cron表达式不能为空",
		}, nil
	}

	// 2. 检查任务名称是否已存在
	exists, err := l.svcCtx.ExternalSyncTaskConfigModel.CheckTaskNameExists(l.ctx, in.TaskName, 0)
	if err != nil {
		l.Logger.Errorf("检查任务名称失败: %v", err)
		return &cmpool.CreateESSyncTaskResp{
			Success: false,
			Message: "检查任务名称失败",
		}, nil
	}
	if exists {
		return &cmpool.CreateESSyncTaskResp{
			Success: false,
			Message: "任务名称已存在",
		}, nil
	}

	// 3. 使用默认值填充可选参数
	esEndpoint := in.EsEndpoint
	if esEndpoint == "" {
		esEndpoint = l.svcCtx.Config.ESDataSource.DefaultEndpoint
	}

	esIndexPattern := in.EsIndexPattern
	if esIndexPattern == "" {
		esIndexPattern = l.svcCtx.Config.ESDataSource.DefaultIndexPattern
	}

	queryTimeRange := in.QueryTimeRange
	if queryTimeRange == "" {
		queryTimeRange = "30d"
	}

	// 4. 插入任务配置
	task := &model.ExternalSyncTaskConfig{
		TaskName:       in.TaskName,
		Description:    sql.NullString{String: in.Description, Valid: in.Description != ""},
		DataSource:     "elasticsearch", // ES任务的数据源是elasticsearch
		EsEndpoint:     sql.NullString{String: esEndpoint, Valid: true},
		EsIndexPattern: sql.NullString{String: esIndexPattern, Valid: true},
		CronExpression: in.CronExpression,
		QueryTimeRange: queryTimeRange,
		IsEnabled:      1, // 默认启用
		CreatedBy:      sql.NullString{String: in.CreatedBy, Valid: in.CreatedBy != ""},
	}

	result, err := l.svcCtx.ExternalSyncTaskConfigModel.Insert(l.ctx, task)
	if err != nil {
		l.Logger.Errorf("插入任务配置失败: %v", err)
		return &cmpool.CreateESSyncTaskResp{
			Success: false,
			Message: "创建任务失败",
		}, nil
	}

	taskId, err := result.LastInsertId()
	if err != nil {
		l.Logger.Errorf("获取任务ID失败: %v", err)
		return &cmpool.CreateESSyncTaskResp{
			Success: false,
			Message: "创建任务失败",
		}, nil
	}

	// TODO: 5. 如果任务默认启用，注册定时任务到调度器
	// if l.svcCtx.ESSyncScheduler != nil {
	//     l.svcCtx.ESSyncScheduler.RegisterTask(taskId, in.CronExpression)
	// }

	l.Logger.Infof("创建ES同步任务成功: ID=%d, Name=%s", taskId, in.TaskName)

	return &cmpool.CreateESSyncTaskResp{
		Success: true,
		Message: "创建任务成功",
		TaskId:  taskId,
	}, nil
}
