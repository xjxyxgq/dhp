package logic

import (
	"context"
	"strings"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteEsSyncByFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteEsSyncByFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteEsSyncByFileLogic {
	return &ExecuteEsSyncByFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据文件立即执行ES数据同步
func (l *ExecuteEsSyncByFileLogic) ExecuteEsSyncByFile(in *cmpool.ExecuteESSyncByFileReq) (*cmpool.ExecuteESSyncResp, error) {
	// 1. 验证参数
	if len(in.FileContent) == 0 {
		return &cmpool.ExecuteESSyncResp{
			Success: false,
			Message: "文件内容不能为空",
		}, nil
	}

	// 2. 解析文件内容，提取主机 IP 列表
	// 支持换行符分隔，过滤空行和空白字符
	fileContentStr := string(in.FileContent)
	lines := strings.Split(fileContentStr, "\n")
	var hostIpList []string
	for _, line := range lines {
		// 去除空白字符
		ip := strings.TrimSpace(line)
		// 过滤空行
		if ip != "" {
			hostIpList = append(hostIpList, ip)
		}
	}

	if len(hostIpList) == 0 {
		return &cmpool.ExecuteESSyncResp{
			Success: false,
			Message: "文件中没有有效的主机IP",
		}, nil
	}

	l.Logger.Infof("从文件解析出 %d 个主机IP", len(hostIpList))

	// 3. 调用 ExecuteEsSyncByHostList 的核心同步逻辑
	hostListLogic := NewExecuteEsSyncByHostListLogic(l.ctx, l.svcCtx)
	req := &cmpool.ExecuteESSyncByHostListReq{
		TaskName:       in.TaskName,
		HostIpList:     hostIpList,
		EsEndpoint:     in.EsEndpoint,
		QueryTimeRange: in.QueryTimeRange,
	}

	resp, err := hostListLogic.ExecuteEsSyncByHostList(req)
	if err != nil {
		l.Logger.Errorf("执行文件同步失败: %v", err)
		return &cmpool.ExecuteESSyncResp{
			Success: false,
			Message: "执行同步失败",
		}, err
	}

	return resp, nil
}
