package main

import (
	"flag"
	"fmt"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/config"
	"cmdb-rpc/internal/global"
	"cmdb-rpc/internal/scheduler"
	"cmdb-rpc/internal/server"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/cmpool.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	// 创建并启动定时任务调度器（硬件资源验证）
	taskScheduler := scheduler.NewTaskScheduler(ctx)
	global.SetTaskScheduler(taskScheduler)
	fmt.Println("Starting task scheduler...")
	err := taskScheduler.Start()
	if err != nil {
		fmt.Printf("Failed to start task scheduler: %v\n", err)
	} else {
		fmt.Println("Task scheduler started successfully")
	}

	// 创建并启动统一外部资源同步调度器（支持 ES 和 CMSys）
	externalSyncScheduler := scheduler.NewExternalSyncScheduler(ctx)
	ctx.ExternalSyncScheduler = externalSyncScheduler // 设置到 ServiceContext
	fmt.Println("Starting External Resource Sync scheduler...")
	if err := externalSyncScheduler.Start(); err != nil {
		fmt.Printf("Failed to start External Resource Sync scheduler: %v\n", err)
	} else {
		fmt.Println("External Resource Sync scheduler started successfully")
	}

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		cmpool.RegisterCmpoolServer(grpcServer, server.NewCmpoolServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer func() {
		s.Stop()
		// 停止定时任务调度器
		fmt.Println("Stopping task scheduler...")
		taskScheduler.Stop()
		fmt.Println("Task scheduler stopped")

		// 停止统一外部资源同步调度器（支持 ES 和 CMSys）
		fmt.Println("Stopping External Resource Sync scheduler...")
		externalSyncScheduler.Stop()
		fmt.Println("External Resource Sync scheduler stopped")
	}()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
