package handler

import (
	"net/http"

	"cmdb-api/internal/logic"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetExternalSyncExecutionLogsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetExternalSyncExecutionLogsRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetExternalSyncExecutionLogsLogic(r.Context(), svcCtx)
		resp, err := l.GetExternalSyncExecutionLogs(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
