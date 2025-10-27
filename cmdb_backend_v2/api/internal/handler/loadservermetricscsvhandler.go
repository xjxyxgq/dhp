package handler

import (
	"net/http"

	"cmdb-api/internal/logic"
	"cmdb-api/internal/svc"
	"cmdb-api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoadServerMetricsCSVHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoadServerMetricsCSVRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		// 获取上传文件
		file, header, err := r.FormFile("file")
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		defer file.Close()

		// 将文件对象传递给logic层处理
		l := logic.NewLoadServerMetricsCSVLogic(r.Context(), svcCtx)
		resp, err := l.LoadServerMetricsCSV(&req, file, header)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
