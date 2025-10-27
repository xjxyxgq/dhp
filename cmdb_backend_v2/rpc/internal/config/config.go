package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	DataSource           string
	HardwareVerification HardwareVerificationConfig
	SSOAuth              AuthConfig
	ExternalAPI          ExternalAPIConfig
	ESDataSource         ESDataSourceConfig
	CMSysDataSource      CMSysDataSourceConfig
}

type HardwareVerificationConfig struct {
	ScriptBasePath string
	RemoteUser     string
	UseSudo        bool
}

type AuthConfig struct {
	EnableCAS        bool
	CASServerURL     string
	ServiceURL       string
	JWTSecret        string
	TokenExpireHours int
}

type ExternalAPIConfig struct {
	CmdbUrl     string
	CmdbAppCode string
	CmdbSecret  string
	Freak       bool
}

type ESDataSourceConfig struct {
	DefaultEndpoint     string
	DefaultIndexPattern string
	TimeoutSeconds      int
}

type CMSysDataSourceConfig struct {
	AuthEndpoint string // 认证接口地址
	DataEndpoint string // 数据接口地址
	AppCode      string // 应用代码
	AppSecret    string // 应用密钥
	Operator     string // 操作员标识
	TimeoutSeconds int  // 超时时间（秒）
}
