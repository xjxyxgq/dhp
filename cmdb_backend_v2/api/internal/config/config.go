package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	RpcConfig zrpc.RpcClientConf
	CorsConf  CorsConfig
	SSOAuth   AuthConfig
}

type CorsConfig struct {
	AccessControlAllowOrigin   string `json:",optional"`
	AccessControlAllowMethods  string `json:",optional"`
	AccessControlAllowHeaders  string `json:",optional"`
	AccessControlExposeHeaders string `json:",optional"`
	AccessControlMaxAge        int    `json:",optional"`
}

type AuthConfig struct {
	EnableCAS        bool
	CASServerURL     string
	ServiceURL       string
	JWTSecret        string
	TokenExpireHours int
}
