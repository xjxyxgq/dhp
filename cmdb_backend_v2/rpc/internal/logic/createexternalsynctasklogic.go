package logic

import (
	"context"
	"database/sql"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/model"
	"cmdb-rpc/internal/svc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

type CreateExternalSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateExternalSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateExternalSyncTaskLogic {
	return &CreateExternalSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CreateExternalSyncTask 统一的外部资源同步任务创建接口（支持 ES 和 CMSys）
func (l *CreateExternalSyncTaskLogic) CreateExternalSyncTask(in *cmpool.CreateExternalSyncTaskReq) (*cmpool.CreateExternalSyncTaskResp, error) {
	// 1. 验证通用参数
	if in.DataSource == "" {
		return &cmpool.CreateExternalSyncTaskResp{
			Success: false,
			Message: "数据源类型不能为空 (elasticsearch/cmsys)",
		}, nil
	}

	if in.DataSource != "elasticsearch" && in.DataSource != "cmsys" {
		return &cmpool.CreateExternalSyncTaskResp{
			Success: false,
			Message: "数据源类型不合法，仅支持: elasticsearch, cmsys",
		}, nil
	}

	if in.TaskName == "" {
		return &cmpool.CreateExternalSyncTaskResp{
			Success: false,
			Message: "任务名称不能为空",
		}, nil
	}

	if in.CronExpression == "" {
		return &cmpool.CreateExternalSyncTaskResp{
			Success: false,
			Message: "Cron表达式不能为空",
		}, nil
	}

	// 验证Cron表达式格式（使用支持秒级精度的解析器）
	parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	if _, err := parser.Parse(in.CronExpression); err != nil {
		return &cmpool.CreateExternalSyncTaskResp{
			Success: false,
			Message: "Cron表达式格式错误: " + err.Error() + "。正确格式为6字段（秒 分 时 日 月 周），例如: '0 0 2 * * *' 表示每天凌晨2点执行",
		}, nil
	}

	// 2. 检查任务名称是否已存在
	exists, err := l.svcCtx.ExternalSyncTaskConfigModel.CheckTaskNameExists(l.ctx, in.TaskName, 0)
	if err != nil {
		l.Logger.Errorf("检查任务名称失败: %v", err)
		return &cmpool.CreateExternalSyncTaskResp{
			Success: false,
			Message: "检查任务名称失败",
		}, nil
	}
	if exists {
		return &cmpool.CreateExternalSyncTaskResp{
			Success: false,
			Message: "任务名称已存在",
		}, nil
	}

	// 3. 根据数据源类型填充默认值
	task := &model.ExternalSyncTaskConfig{
		TaskName:       in.TaskName,
		Description:    sql.NullString{String: in.Description, Valid: in.Description != ""},
		DataSource:     in.DataSource,
		CronExpression: in.CronExpression,
		IsEnabled:      1, // 默认启用
		CreatedBy:      sql.NullString{String: in.CreatedBy, Valid: in.CreatedBy != ""},
	}

	if in.DataSource == "elasticsearch" {
		// ES 数据源特定配置
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

		task.EsEndpoint = sql.NullString{String: esEndpoint, Valid: true}
		task.EsIndexPattern = sql.NullString{String: esIndexPattern, Valid: true}
		task.QueryTimeRange = queryTimeRange

	} else if in.DataSource == "cmsys" {
		// CMSys 数据源特定配置
		cmsysQuery := in.CmsysQuery
		if cmsysQuery == "" {
			// CMSys的默认查询参数可以为空或使用配置的默认值（如果有）
			cmsysQuery = ""
		}

		task.CmsysQuery = sql.NullString{String: cmsysQuery, Valid: cmsysQuery != ""}
		// CMSys 可能不使用 QueryTimeRange，或使用固定值
		task.QueryTimeRange = ""
		// CMSys endpoint 从配置中读取，不存储在任务配置中
	}

	// 4. 插入任务配置
	result, err := l.svcCtx.ExternalSyncTaskConfigModel.Insert(l.ctx, task)
	if err != nil {
		l.Logger.Errorf("插入任务配置失败: %v", err)
		return &cmpool.CreateExternalSyncTaskResp{
			Success: false,
			Message: "创建任务失败",
		}, nil
	}

	taskId, err := result.LastInsertId()
	if err != nil {
		l.Logger.Errorf("获取任务ID失败: %v", err)
		return &cmpool.CreateExternalSyncTaskResp{
			Success: false,
			Message: "创建任务失败",
		}, nil
	}

	// TODO: 5. 如果任务默认启用，根据数据源类型注册定时任务到调度器
	// if in.DataSource == "elasticsearch" && l.svcCtx.ESSyncScheduler != nil {
	//     l.svcCtx.ESSyncScheduler.RegisterTask(taskId, in.CronExpression)
	// } else if in.DataSource == "cmsys" && l.svcCtx.CMSysSyncScheduler != nil {
	//     l.svcCtx.CMSysSyncScheduler.RegisterTask(taskId, in.CronExpression)
	// }

	l.Logger.Infof("创建外部同步任务成功: ID=%d, Name=%s, DataSource=%s", taskId, in.TaskName, in.DataSource)

	return &cmpool.CreateExternalSyncTaskResp{
		Success: true,
		Message: "创建任务成功",
		TaskId:  taskId,
	}, nil
}
