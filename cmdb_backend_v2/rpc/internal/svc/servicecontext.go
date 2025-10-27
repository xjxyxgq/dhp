package svc

import (
	"cmdb-rpc/internal/config"
	"cmdb-rpc/internal/datasource"
	"cmdb-rpc/internal/model"

	_ "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config      config.Config
	DB          sqlx.SqlConn
	CSVLoader   *datasource.CSVLoader
	ExternalAPI *datasource.ExternalAPI
	DataSync    *datasource.DataSync

	// 数据库模型
	HostsPoolModel                    model.HostsPoolModel
	HostsApplicationsModel            model.HostsApplicationsModel
	IdcConfModel                      model.IdcConfModel
	MysqlClusterModel                 model.MysqlClusterModel
	MysqlClusterInstanceModel         model.MysqlClusterInstanceModel
	MssqlClusterModel                 model.MssqlClusterModel
	MssqlClusterInstanceModel         model.MssqlClusterInstanceModel
	TidbClusterModel                  model.TidbClusterModel
	TidbClusterInstanceModel          model.TidbClusterInstanceModel
	GoldendbClusterModel              model.GoldendbClusterModel
	GoldendbClusterInstanceModel      model.GoldendbClusterInstanceModel
	DbLineModel                       model.DbLineModel
	ServerResourcesModel              model.ServerResourcesModel
	ClusterGroupsModel                model.ClusterGroupsModel
	BackupRestoreCheckInfoModel       model.BackupRestoreCheckInfoModel
	PluginExecutionRecordsModel       model.PluginExecutionRecordsModel
	ResourceAnalysisReportsModel      model.ResourceAnalysisReportsModel
	HardwareResourceVerificationModel model.HardwareResourceVerificationModel
	ScheduledTaskModel                model.ScheduledHardwareVerificationModel
	ScheduledTaskHistoryModel         model.ScheduledTaskExecutionHistoryModel
	UserModel                         model.UserModel
	UserSessionModel                  model.UserSessionModel

	// 统一外部资源同步相关模型
	ExternalSyncTaskConfigModel      model.ExternalSyncTaskConfigModel
	ExternalSyncExecutionLogModel    model.ExternalSyncExecutionLogModel
	ExternalSyncExecutionDetailModel model.ExternalSyncExecutionDetailModel

	// ES数据同步调度器（使用 interface{} 避免循环依赖，实际类型为 *scheduler.EsSyncScheduler）
	EsSyncScheduler interface{}

	// 统一外部资源同步调度器（使用 interface{} 避免循环依赖，实际类型为 *scheduler.ExternalSyncScheduler）
	ExternalSyncScheduler interface{}
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewMysql(c.DataSource)

	svcCtx := &ServiceContext{
		Config: c,
		//DB:          db,
		CSVLoader:   datasource.NewCSVLoader(db),
		ExternalAPI: datasource.NewExternalAPI(db, &c.ExternalAPI),
		DataSync:    datasource.NewDataSync(db),

		// 初始化所有数据库模型
		HostsPoolModel:                    model.NewHostsPoolModel(db),
		HostsApplicationsModel:            model.NewHostsApplicationsModel(db),
		IdcConfModel:                      model.NewIdcConfModel(db),
		MysqlClusterModel:                 model.NewMysqlClusterModel(db),
		MysqlClusterInstanceModel:         model.NewMysqlClusterInstanceModel(db),
		MssqlClusterModel:                 model.NewMssqlClusterModel(db),
		MssqlClusterInstanceModel:         model.NewMssqlClusterInstanceModel(db),
		TidbClusterModel:                  model.NewTidbClusterModel(db),
		TidbClusterInstanceModel:          model.NewTidbClusterInstanceModel(db),
		GoldendbClusterModel:              model.NewGoldendbClusterModel(db),
		GoldendbClusterInstanceModel:      model.NewGoldendbClusterInstanceModel(db),
		DbLineModel:                       model.NewDbLineModel(db),
		ServerResourcesModel:              model.NewServerResourcesModel(db),
		ClusterGroupsModel:                model.NewClusterGroupsModel(db),
		BackupRestoreCheckInfoModel:       model.NewBackupRestoreCheckInfoModel(db),
		PluginExecutionRecordsModel:       model.NewPluginExecutionRecordsModel(db),
		ResourceAnalysisReportsModel:      model.NewResourceAnalysisReportsModel(db),
		HardwareResourceVerificationModel: model.NewHardwareResourceVerificationModel(db),
		ScheduledTaskModel:                model.NewScheduledHardwareVerificationModel(db),
		ScheduledTaskHistoryModel:         model.NewScheduledTaskExecutionHistoryModel(db),
		UserModel:                         model.NewUserModel(db),
		UserSessionModel:                  model.NewUserSessionModel(db),

		// 统一外部资源同步相关模型初始化
		ExternalSyncTaskConfigModel:      model.NewExternalSyncTaskConfigModel(db),
		ExternalSyncExecutionLogModel:    model.NewExternalSyncExecutionLogModel(db),
		ExternalSyncExecutionDetailModel: model.NewExternalSyncExecutionDetailModel(db),
	}

	// 注意：EsSyncScheduler 和 ExternalSyncScheduler 将在 main 函数中创建和设置，以避免循环依赖

	return svcCtx
}
