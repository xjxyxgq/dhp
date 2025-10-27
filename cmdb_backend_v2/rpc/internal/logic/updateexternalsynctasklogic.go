package logic

import (
	"context"
	"database/sql"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/robfig/cron/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateExternalSyncTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateExternalSyncTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateExternalSyncTaskLogic {
	return &UpdateExternalSyncTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateExternalSyncTask 更新外部同步任务配置
func (l *UpdateExternalSyncTaskLogic) UpdateExternalSyncTask(in *cmpool.UpdateExternalSyncTaskReq) (*cmpool.UpdateExternalSyncTaskResp, error) {
	// 1. 验证任务ID
	if in.Id <= 0 {
		return &cmpool.UpdateExternalSyncTaskResp{
			Success: false,
			Message: "任务ID不能为空",
		}, nil
	}

	// 2. 查询任务是否存在
	existingTask, err := l.svcCtx.ExternalSyncTaskConfigModel.FindOne(l.ctx, uint64(in.Id))
	if err != nil {
		if err == sql.ErrNoRows {
			return &cmpool.UpdateExternalSyncTaskResp{
				Success: false,
				Message: "任务不存在",
			}, nil
		}
		l.Logger.Errorf("查询任务失败: %v", err)
		return &cmpool.UpdateExternalSyncTaskResp{
			Success: false,
			Message: "查询任务失败",
		}, nil
	}

	// 3. 验证数据源类型（如果提供）
	dataSource := in.DataSource
	if dataSource == "" {
		dataSource = existingTask.DataSource // 保持原数据源
	} else if dataSource != "elasticsearch" && dataSource != "cmsys" {
		return &cmpool.UpdateExternalSyncTaskResp{
			Success: false,
			Message: "数据源类型不合法，仅支持: elasticsearch, cmsys",
		}, nil
	}

	// 4. 检查任务名是否与其他任务重复
	if in.TaskName != "" && in.TaskName != existingTask.TaskName {
		exists, err := l.svcCtx.ExternalSyncTaskConfigModel.CheckTaskNameExists(l.ctx, in.TaskName, uint64(in.Id))
		if err != nil {
			l.Logger.Errorf("检查任务名称失败: %v", err)
			return &cmpool.UpdateExternalSyncTaskResp{
				Success: false,
				Message: "检查任务名称失败",
			}, nil
		}
		if exists {
			return &cmpool.UpdateExternalSyncTaskResp{
				Success: false,
				Message: "任务名称已存在",
			}, nil
		}
		existingTask.TaskName = in.TaskName
	}

	// 5. 更新通用字段
	if in.Description != "" {
		existingTask.Description = sql.NullString{String: in.Description, Valid: true}
	}
	if in.CronExpression != "" {
		// 验证Cron表达式格式（使用支持秒级精度的解析器）
		parser := cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
		if _, err := parser.Parse(in.CronExpression); err != nil {
			return &cmpool.UpdateExternalSyncTaskResp{
				Success: false,
				Message: "Cron表达式格式错误: " + err.Error() + "。正确格式为6字段（秒 分 时 日 月 周），例如: '0 0 2 * * *' 表示每天凌晨2点执行",
			}, nil
		}
		existingTask.CronExpression = in.CronExpression
	}

	// 6. 根据数据源类型更新特定字段
	existingTask.DataSource = dataSource

	if dataSource == "elasticsearch" {
		// 更新 ES 特定配置
		if in.EsEndpoint != "" {
			existingTask.EsEndpoint = sql.NullString{String: in.EsEndpoint, Valid: true}
		}
		if in.EsIndexPattern != "" {
			existingTask.EsIndexPattern = sql.NullString{String: in.EsIndexPattern, Valid: true}
		}
		if in.QueryTimeRange != "" {
			existingTask.QueryTimeRange = in.QueryTimeRange
		}
		// 清空 CMSys 字段
		existingTask.CmsysQuery = sql.NullString{Valid: false}

	} else if dataSource == "cmsys" {
		// 更新 CMSys 特定配置
		if in.CmsysQuery != "" {
			existingTask.CmsysQuery = sql.NullString{String: in.CmsysQuery, Valid: true}
		}
		// CMSys 不使用 QueryTimeRange，但保留字段
		existingTask.QueryTimeRange = ""
		// 清空 ES 字段
		existingTask.EsEndpoint = sql.NullString{Valid: false}
		existingTask.EsIndexPattern = sql.NullString{Valid: false}
	}

	// 7. 更新数据库
	err = l.svcCtx.ExternalSyncTaskConfigModel.Update(l.ctx, existingTask)
	if err != nil {
		l.Logger.Errorf("更新任务配置失败: %v", err)
		return &cmpool.UpdateExternalSyncTaskResp{
			Success: false,
			Message: "更新任务失败",
		}, nil
	}

	// TODO: 8. 如果 Cron 表达式改变，更新调度器
	// if in.CronExpression != "" && in.CronExpression != originalCronExpression {
	//     if dataSource == "elasticsearch" && l.svcCtx.ESSyncScheduler != nil {
	//         l.svcCtx.ESSyncScheduler.UpdateTask(in.Id, in.CronExpression)
	//     } else if dataSource == "cmsys" && l.svcCtx.CMSysSyncScheduler != nil {
	//         l.svcCtx.CMSysSyncScheduler.UpdateTask(in.Id, in.CronExpression)
	//     }
	// }

	l.Logger.Infof("更新外部同步任务成功: TaskID=%d, Name=%s, DataSource=%s", in.Id, existingTask.TaskName, dataSource)

	return &cmpool.UpdateExternalSyncTaskResp{
		Success: true,
		Message: "更新任务成功",
	}, nil
}
