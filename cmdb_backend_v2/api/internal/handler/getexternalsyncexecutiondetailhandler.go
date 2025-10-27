package handler

import (
	"net/http"
	"strconv"

	"cmdb-api/internal/logic"
	"cmdb-api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/pathvar"
)

func GetExternalSyncExecutionDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 提取路径参数 execution_id
		vars := pathvar.Vars(r)
		executionIdStr := vars["execution_id"]
		executionId, err := strconv.ParseInt(executionIdStr, 10, 64)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetExternalSyncExecutionDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetExternalSyncExecutionDetail(executionId)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
