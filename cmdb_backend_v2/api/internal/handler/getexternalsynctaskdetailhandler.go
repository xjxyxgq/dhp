package handler

import (
	"net/http"
	"strconv"

	"cmdb-api/internal/logic"
	"cmdb-api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetExternalSyncTaskDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 从URL路径参数中获取task_id
		taskIdStr := r.URL.Query().Get("task_id")
		if taskIdStr == "" {
			// 如果query参数没有，尝试从路径参数获取
			taskIdStr = r.PathValue("task_id")
		}

		taskId, err := strconv.ParseInt(taskIdStr, 10, 64)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetExternalSyncTaskDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetExternalSyncTaskDetail(taskId)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
