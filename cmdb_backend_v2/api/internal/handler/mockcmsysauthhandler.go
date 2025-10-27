package handler

import (
	"encoding/json"
	"net/http"

	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// MockCMSysAuthHandler 模拟CMSys认证接口，用于开发测试
func MockCMSysAuthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logx.Info("收到Mock CMSys认证请求")

		// 解析请求体
		var reqBody struct {
			AppCode string `json:"appCode"`
			Secret  string `json:"secret"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			httpx.Error(w, err)
			return
		}

		logx.Infof("Mock CMSys认证 - AppCode: %s", reqBody.AppCode)

		// 生成模拟token（固定值，方便测试）
		mockToken := "mock-cmsys-token-eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"

		// 返回响应
		response := map[string]interface{}{
			"code": "A0000",
			"msg":  "success",
			"data": mockToken,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
