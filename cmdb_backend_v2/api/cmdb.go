package main

import (
	"flag"
	"fmt"
	"net/http"

	"cmdb-api/internal/config"
	"cmdb-api/internal/handler"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/cmdb-api.yaml", "the config file")

// CORS中间件
func corsMiddleware(corsConf config.CorsConfig) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 设置CORS头，使用配置文件中的值或默认值
			origin := corsConf.AccessControlAllowOrigin
			if origin == "" {
				origin = "*"
			}
			methods := corsConf.AccessControlAllowMethods
			if methods == "" {
				methods = "GET, POST, PUT, DELETE, OPTIONS"
			}
			headers := corsConf.AccessControlAllowHeaders
			if headers == "" {
				headers = "Content-Type, Authorization, X-Requested-With"
			}
			exposeHeaders := corsConf.AccessControlExposeHeaders
			if exposeHeaders == "" {
				exposeHeaders = "Content-Length"
			}
			maxAge := corsConf.AccessControlMaxAge
			if maxAge == 0 {
				maxAge = 86400
			}

			// 记录请求信息
			logx.Infof("CORS请求: %s %s from %s, Content-Type: %s", r.Method, r.URL.Path, r.Header.Get("Origin"), r.Header.Get("Content-Type"))

			// 始终设置CORS头
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)
			w.Header().Set("Access-Control-Expose-Headers", exposeHeaders)
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", maxAge))

			// 处理预检请求
			if r.Method == "OPTIONS" {
				logx.Infof("处理OPTIONS预检请求: %s", r.URL.Path)
				w.WriteHeader(http.StatusOK)
				return
			}

			// 继续处理正常请求
			next(w, r)
		}
	}
}


func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	// 添加CORS中间件（必须在路由注册之前）
	server.Use(corsMiddleware(c.CorsConf))

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
