package handler

import (
	"net/http"

	"cmdb-api/internal/logic"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func SyncExternalCmdbHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SyncExternalCmdbRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewSyncExternalCmdbLogic(r.Context(), svcCtx)
		resp, err := l.SyncExternalCmdb(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
