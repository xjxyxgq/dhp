package logic

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExecuteExternalSyncByFileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteExternalSyncByFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteExternalSyncByFileLogic {
	return &ExecuteExternalSyncByFileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ExecuteExternalSyncByFile 统一的按文件执行同步接口（支持 ES 和 CMSys）
func (l *ExecuteExternalSyncByFileLogic) ExecuteExternalSyncByFile(in *cmpool.ExecuteExternalSyncByFileReq) (*cmpool.ExecuteExternalSyncResp, error) {
	// 1. 验证 data_source 参数
	if in.DataSource == "" {
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: "数据源类型不能为空 (elasticsearch/cmsys)",
		}, nil
	}

	// 2. 验证文件内容
	if len(in.FileContent) == 0 {
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: "文件内容不能为空",
		}, nil
	}

	// 3. 解析文件内容，提取IP列表
	hostIpList, err := l.parseIPListFromFile(in.FileContent)
	if err != nil {
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: fmt.Sprintf("解析文件失败: %v", err),
		}, nil
	}

	if len(hostIpList) == 0 {
		return &cmpool.ExecuteExternalSyncResp{
			Success: false,
			Message: "文件中没有找到有效的IP地址",
		}, nil
	}

	l.Logger.Infof("从文件 %s 解析到 %d 个主机IP", in.Filename, len(hostIpList))

	// 4. 转换为按主机列表执行的请求
	hostListReq := &cmpool.ExecuteExternalSyncByHostListReq{
		DataSource:     in.DataSource,
		HostIpList:     hostIpList,
		TaskName:       in.TaskName,
		QueryTimeRange: in.QueryTimeRange,
		EsEndpoint:     in.EsEndpoint,
		CmsysQuery:     in.CmsysQuery,
	}

	// 5. 调用按主机列表执行的逻辑
	hostListLogic := NewExecuteExternalSyncByHostListLogic(l.ctx, l.svcCtx)
	return hostListLogic.ExecuteExternalSyncByHostList(hostListReq)
}

// parseIPListFromFile 从文件内容中解析IP列表
func (l *ExecuteExternalSyncByFileLogic) parseIPListFromFile(fileContent []byte) ([]string, error) {
	var ipList []string
	ipMap := make(map[string]bool) // 用于去重

	reader := bytes.NewReader(fileContent)
	scanner := bufio.NewScanner(reader)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// 简单验证IP格式（支持IPv4）
		ip := line
		// 如果行中包含空格或逗号，取第一个字段
		if strings.Contains(line, " ") {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				ip = fields[0]
			}
		} else if strings.Contains(line, ",") {
			fields := strings.Split(line, ",")
			if len(fields) > 0 {
				ip = strings.TrimSpace(fields[0])
			}
		}

		// 基本的IP格式验证（简单检查）
		if l.isValidIP(ip) {
			// 去重
			if !ipMap[ip] {
				ipList = append(ipList, ip)
				ipMap[ip] = true
			}
		} else {
			l.Logger.Infof("第%d行包含无效的IP地址: %s (已跳过)", lineNum, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %v", err)
	}

	return ipList, nil
}

// isValidIP 简单的IP格式验证
func (l *ExecuteExternalSyncByFileLogic) isValidIP(ip string) bool {
	if ip == "" {
		return false
	}

	// 基本的IPv4格式检查
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if part == "" || len(part) > 3 {
			return false
		}
		// 检查是否全是数字
		for _, c := range part {
			if c < '0' || c > '9' {
				return false
			}
		}
	}

	return true
}
