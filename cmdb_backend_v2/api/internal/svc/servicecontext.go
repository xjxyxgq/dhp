package svc

import (
	"cmdb-api/cmpool"
	"cmdb-api/internal/config"
	"cmdb-api/internal/middleware"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config    config.Config
	CmpoolRpc cmpool.CmpoolClient
	Auth      rest.Middleware
}

func NewServiceContext(c config.Config) *ServiceContext {
	rpcClient := zrpc.MustNewClient(c.RpcConfig)

	ctx := &ServiceContext{
		Config:    c,
		CmpoolRpc: cmpool.NewCmpoolClient(rpcClient.Conn()),
	}
	
	ctx.Auth = middleware.NewAuthMiddleware(cmpool.NewCmpoolClient(rpcClient.Conn())).Handle

	return ctx
}
