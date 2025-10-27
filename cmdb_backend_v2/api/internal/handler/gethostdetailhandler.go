package handler

import (
	"net/http"

	"cmdb-api/internal/logic"
	"cmdb-api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetHostDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetHostDetailLogic(r.Context(), svcCtx)
		resp, err := l.GetHostDetail()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
