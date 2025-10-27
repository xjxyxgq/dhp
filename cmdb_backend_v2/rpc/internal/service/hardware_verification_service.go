package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cmdb-rpc/cmpool"
	"cmdb-rpc/internal/config"
	"cmdb-rpc/internal/model"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

// HardwareVerificationService 硬件验证服务接口
type HardwareVerificationService interface {
	ExecuteVerification(req *cmpool.HardwareResourceVerificationReq) (*cmpool.HardwareResourceVerificationResp, error)
}

// hardwareVerificationService 硬件验证服务实现
type hardwareVerificationService struct {
	ctx context.Context
	//db                            sqlx.SqlConn
	config                            config.Config
	hardwareResourceVerificationModel model.HardwareResourceVerificationModel
	logger                            logx.Logger
}

// NewHardwareVerificationService 创建硬件验证服务
func NewHardwareVerificationService(ctx context.Context, config config.Config, hwModel model.HardwareResourceVerificationModel) HardwareVerificationService {
	return &hardwareVerificationService{
		ctx: ctx,
		//db:                            db,
		config:                            config,
		hardwareResourceVerificationModel: hwModel,
		logger:                            logx.WithContext(ctx),
	}
}

// ExecuteVerification 执行硬件资源验证
func (s *hardwareVerificationService) ExecuteVerification(req *cmpool.HardwareResourceVerificationReq) (*cmpool.HardwareResourceVerificationResp, error) {
	// 生成唯一任务ID
	taskId := uuid.New().String()

	// 验证参数
	if len(req.HostIpList) == 0 {
		return &cmpool.HardwareResourceVerificationResp{
			Success: false,
			Message: "主机IP列表不能为空",
		}, nil
	}

	if req.ResourceType == "" {
		return &cmpool.HardwareResourceVerificationResp{
			Success: false,
			Message: "资源类型不能为空",
		}, nil
	}

	if req.ResourceType != "cpu" && req.ResourceType != "memory" && req.ResourceType != "disk" {
		return &cmpool.HardwareResourceVerificationResp{
			Success: false,
			Message: "资源类型必须是 cpu、memory 或 disk",
		}, nil
	}

	if req.TargetPercent <= 0 || req.TargetPercent > 100 {
		return &cmpool.HardwareResourceVerificationResp{
			Success: false,
			Message: "目标百分比必须在1-100之间",
		}, nil
	}

	duration := req.Duration
	if duration <= 0 {
		duration = 300 // 默认5分钟
	}

	// 检查任务冲突并处理
	processedHostList, err := s.handleTaskConflicts(req.HostIpList, req.ResourceType, req.ForceExecution)
	if err != nil {
		return &cmpool.HardwareResourceVerificationResp{
			Success: false,
			Message: fmt.Sprintf("处理任务冲突时出错: %v", err),
		}, nil
	}

	if len(processedHostList) == 0 {
		return &cmpool.HardwareResourceVerificationResp{
			Success: false,
			Message: "所有主机都有正在运行的任务，已跳过执行",
		}, nil
	}

	// 批量插入数据库记录
	for _, hostIp := range processedHostList {
		verification := &model.HardwareResourceVerification{
			TaskId:          taskId,
			HostIp:          hostIp,
			ResourceType:    req.ResourceType,
			TargetPercent:   uint64(req.TargetPercent),
			Duration:        uint64(duration),
			ScriptParams:    sql.NullString{String: req.ScriptParams, Valid: req.ScriptParams != ""},
			ExecutionStatus: "pending",
		}

		_, err := s.hardwareResourceVerificationModel.InsertVerification(s.ctx, verification)
		if err != nil {
			s.logger.Errorf("插入验证记录失败: %v", err)
			continue
		}
	}

	// 启动异步执行
	go s.executeVerificationTask(taskId, processedHostList, req.ResourceType, req.TargetPercent, duration, req.ScriptParams)

	return &cmpool.HardwareResourceVerificationResp{
		Success: true,
		Message: fmt.Sprintf("硬件资源验证任务已创建，任务ID: %s", taskId),
		TaskId:  taskId,
	}, nil
}

// executeVerificationTask 执行验证任务
func (s *hardwareVerificationService) executeVerificationTask(taskId string, hostIpList []string, resourceType string, targetPercent, duration int32, scriptParams string) {
	var wg sync.WaitGroup

	// 并发执行验证，但限制并发数量
	semaphore := make(chan struct{}, 5) // 最多同时执行5个

	for _, hostIp := range hostIpList {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			semaphore <- struct{}{}        // 获取信号量
			defer func() { <-semaphore }() // 释放信号量

			s.executeHostVerification(taskId, ip, resourceType, targetPercent, duration, scriptParams)
		}(hostIp)
	}

	wg.Wait()
	s.logger.Infof("任务 %s 的所有验证已完成", taskId)
}

// executeHostVerification 执行单个主机的验证
func (s *hardwareVerificationService) executeHostVerification(taskId, hostIp, resourceType string, targetPercent, duration int32, scriptParams string) {
	// 查找对应的数据库记录
	records, err := s.hardwareResourceVerificationModel.FindByTaskId(s.ctx, taskId)
	if err != nil {
		s.logger.Errorf("查找验证记录失败: %v", err)
		return
	}

	var record *model.HardwareResourceVerification
	for _, r := range records {
		if r.HostIp == hostIp && r.ResourceType == resourceType {
			record = r
			break
		}
	}

	if record == nil {
		s.logger.Errorf("未找到主机 %s 的验证记录", hostIp)
		return
	}

	startTime := time.Now().Format("2006-01-02 15:04:05")

	// 更新状态为running
	err = s.hardwareResourceVerificationModel.UpdateVerificationStatus(s.ctx,
		int64(record.Id), "running", startTime, "", sql.NullInt64{},
		sql.NullString{}, sql.NullString{}, sql.NullString{}, sql.NullString{})
	if err != nil {
		s.logger.Errorf("更新验证状态失败: %v", err)
		return
	}

	// 执行SSH命令
	exitCode, stdout, stderr, sshError := s.executeSSHCommand(hostIp, resourceType, targetPercent, duration, scriptParams)

	endTime := time.Now().Format("2006-01-02 15:04:05")
	executionStatus := "completed"
	if exitCode != 0 || sshError != "" {
		executionStatus = "failed"
	}

	// 解析结果摘要
	resultSummary := s.parseResultSummary(stdout, stderr, exitCode)

	// 更新最终状态
	err = s.hardwareResourceVerificationModel.UpdateVerificationStatus(
		s.ctx,
		int64(record.Id), executionStatus, startTime, endTime,
		sql.NullInt64{Int64: int64(exitCode), Valid: true},
		sql.NullString{String: stdout, Valid: true},
		sql.NullString{String: stderr, Valid: true},
		sql.NullString{String: resultSummary, Valid: true},
		sql.NullString{String: sshError, Valid: sshError != ""})

	if err != nil {
		s.logger.Errorf("更新最终验证状态失败: %v", err)
	}

	s.logger.Infof("主机 %s 的 %s 资源验证已完成，状态: %s", hostIp, resourceType, executionStatus)
}

// executeSSHCommand 执行SSH命令
func (s *hardwareVerificationService) executeSSHCommand(hostIp, resourceType string, targetPercent, duration int32, scriptParams string) (int, string, string, string) {
	// 确定脚本名称
	var scriptName string
	switch resourceType {
	case "cpu":
		scriptName = "cpu_load_limit.sh"
	case "memory":
		scriptName = "memory_usage_limit.sh"
	case "disk":
		scriptName = "disk_usage_limit.sh"
	default:
		return -1, "", fmt.Sprintf("不支持的资源类型: %s", resourceType), "参数错误"
	}

	// 从配置文件获取脚本路径
	scriptBasePath := s.config.HardwareVerification.ScriptBasePath
	if scriptBasePath == "" {
		scriptBasePath = "/tmp/scripts" // 默认路径
	}
	localScriptPath := filepath.Join(scriptBasePath, scriptName)
	remoteScriptPath := fmt.Sprintf("/tmp/%s", scriptName)

	// 获取远程用户配置
	remoteUser := s.config.HardwareVerification.RemoteUser
	if remoteUser == "" {
		remoteUser = "root" // 默认用户
	}

	// 构建SSH目标地址
	sshTarget := fmt.Sprintf("%s@%s", remoteUser, hostIp)

	// 1. 复制脚本到远程主机
	scpCmd := exec.Command("scp", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		localScriptPath, fmt.Sprintf("%s:%s", sshTarget, remoteScriptPath))
	scpOutput, scpErr := scpCmd.CombinedOutput()
	if scpErr != nil {
		sshError := fmt.Sprintf("SCP复制脚本失败: %v, 输出: %s", scpErr, string(scpOutput))
		return -1, "", string(scpOutput), sshError
	}

	// 2. 设置脚本执行权限
	chmodCmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		sshTarget, fmt.Sprintf("chmod +x %s", remoteScriptPath))
	chmodOutput, chmodErr := chmodCmd.CombinedOutput()
	if chmodErr != nil {
		sshError := fmt.Sprintf("设置脚本权限失败: %v, 输出: %s", chmodErr, string(chmodOutput))
		return -1, "", string(chmodOutput), sshError
	}

	// 3. 构建执行命令
	var cmdArgs []string
	switch resourceType {
	case "cpu":
		cmdArgs = []string{fmt.Sprintf("%d", targetPercent), fmt.Sprintf("%d", duration), "auto", "19"}
	case "memory":
		cmdArgs = []string{fmt.Sprintf("%d", targetPercent), fmt.Sprintf("%d", duration), "19"}
	case "disk":
		cmdArgs = []string{fmt.Sprintf("%d", targetPercent), fmt.Sprintf("%d", duration), "/tmp", "19"}
	}

	// 解析额外参数
	if scriptParams != "" {
		var params map[string]interface{}
		if err := json.Unmarshal([]byte(scriptParams), &params); err == nil {
			// 根据参数覆盖默认值
			if val, ok := params["nice_level"]; ok {
				if strVal, ok := val.(string); ok && len(cmdArgs) > 3 {
					cmdArgs[3] = strVal
				}
			}
			// 可以根据需要添加更多参数处理
		}
	}

	// 根据配置决定是否使用sudo
	useSudo := s.config.HardwareVerification.UseSudo
	var execCommand string
	if useSudo {
		// 使用sudo -n (non-interactive)避免密码提示，要求配置NOPASSWD
		execCommand = fmt.Sprintf("sudo -n %s %s", remoteScriptPath, strings.Join(cmdArgs, " "))
	} else {
		execCommand = fmt.Sprintf("%s %s", remoteScriptPath, strings.Join(cmdArgs, " "))
	}

	// 4. 执行脚本
	sshCmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		sshTarget, execCommand)
	output, err := sshCmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// SSH连接错误
			sshError := fmt.Sprintf("SSH连接失败: %v", err)
			return -1, "", string(output), sshError
		}
	}

	// 5. 清理远程脚本
	cleanupCmd := exec.Command("ssh", "-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null",
		sshTarget, fmt.Sprintf("rm -f %s", remoteScriptPath))
	cleanupCmd.Run() // 忽略清理错误

	// 分离标准输出和标准错误（这里简化处理，实际可能需要更复杂的解析）
	outputStr := string(output)
	stdout := outputStr
	stderr := ""

	if exitCode != 0 {
		// 如果有错误，将输出作为stderr
		stderr = outputStr
		stdout = ""
	}

	return exitCode, stdout, stderr, ""
}

// parseResultSummary 解析执行结果摘要
func (s *hardwareVerificationService) parseResultSummary(stdout, stderr string, exitCode int) string {
	summary := map[string]interface{}{
		"exit_code": exitCode,
		"success":   exitCode == 0,
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
	}

	// 尝试从输出中提取关键信息
	if exitCode == 0 && stdout != "" {
		// 解析成功的执行结果
		lines := strings.Split(stdout, "\n")
		for _, line := range lines {
			if strings.Contains(line, "已达标") {
				summary["target_achieved"] = true
			}
			if strings.Contains(line, "压测成功完成") || strings.Contains(line, "程序成功完成") {
				summary["test_completed"] = true
			}
		}
	}

	if stderr != "" {
		summary["has_error"] = true
		summary["error_summary"] = stderr[:min(200, len(stderr))] // 截取前200字符作为错误摘要
	}

	jsonBytes, _ := json.Marshal(summary)
	return string(jsonBytes)
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// handleTaskConflicts 处理任务冲突
func (s *hardwareVerificationService) handleTaskConflicts(hostIpList []string, resourceType string, forceExecution bool) ([]string, error) {
	var processedHosts []string
	var conflictingHosts []string

	for _, hostIp := range hostIpList {
		// 检查该主机是否有正在运行的任务
		hasRunningTask, err := s.hasRunningTask(hostIp, resourceType)
		if err != nil {
			s.logger.Errorf("检查主机 %s 运行任务时出错: %v", hostIp, err)
			continue
		}

		if hasRunningTask {
			conflictingHosts = append(conflictingHosts, hostIp)
			if forceExecution {
				// 强制执行：终止旧任务
				err := s.terminateRunningTasks(hostIp, resourceType)
				if err != nil {
					s.logger.Errorf("终止主机 %s 运行任务时出错: %v", hostIp, err)
					continue
				}
				s.logger.Infof("已强制终止主机 %s 的运行任务，将执行新任务", hostIp)
				processedHosts = append(processedHosts, hostIp)
			} else {
				// 跳过执行
				s.logger.Infof("主机 %s 有正在运行的任务，跳过执行", hostIp)
			}
		} else {
			// 没有冲突，正常添加
			processedHosts = append(processedHosts, hostIp)
		}
	}

	if len(conflictingHosts) > 0 {
		if forceExecution {
			s.logger.Infof("检测到 %d 个主机有冲突任务，已强制终止并执行新任务", len(conflictingHosts))
		} else {
			s.logger.Infof("检测到 %d 个主机有冲突任务，已跳过执行", len(conflictingHosts))
		}
	}

	return processedHosts, nil
}

// hasRunningTask 检查主机是否有正在运行的任务
func (s *hardwareVerificationService) hasRunningTask(hostIp, resourceType string) (bool, error) {
	if found, err := s.hardwareResourceVerificationModel.HasRunningTask(s.ctx, hostIp, resourceType); err != nil {
		return false, err
	} else {
		return found, nil
	}
}

// terminateRunningTasks 终止正在运行的任务
func (s *hardwareVerificationService) terminateRunningTasks(hostIp, resourceType string) error {
	err := s.hardwareResourceVerificationModel.TerminateRunningTasks(s.ctx, hostIp, resourceType)
	if err != nil {
		return fmt.Errorf("更新终止任务状态失败: %v", err)
	}

	// 尝试终止远程进程（这里可以根据实际情况实现进程终止逻辑）
	// 由于我们无法直接管理远程进程，这里只是标记任务为终止状态
	// 实际的进程终止可能需要通过SSH连接或其他方式实现

	s.logger.Infof("已标记主机 %s 的 %s 验证任务为终止状态", hostIp, resourceType)
	return nil
}
