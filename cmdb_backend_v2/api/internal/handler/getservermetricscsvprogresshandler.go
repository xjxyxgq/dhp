package handler

import (
	"net/http"

	"cmdb-api/internal/logic"
	"cmdb-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetServerMetricsCSVProgressHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetServerMetricsCSVProgressLogic(r.Context(), svcCtx)
		resp, err := l.GetServerMetricsCSVProgress()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
