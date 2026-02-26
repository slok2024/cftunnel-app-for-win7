package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AppVersion 客户端版本（构建时通过 ldflags 注入）
var AppVersion = "dev"

// App 桌面客户端主结构
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// StatusInfo 隧道状态
type StatusInfo struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Output    string `json:"output"`
}

// RouteInfo 路由信息
type RouteInfo struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Service  string `json:"service"`
}

// QuickResult quick 模式结果
type QuickResult struct {
	URL string `json:"url"`
	Err string `json:"err"`
}

// runCftunnel 执行 cftunnel 子命令（Windows 隐藏窗口）
func runCftunnel(args ...string) (string, error) {
	cmd := exec.Command("cftunnel", args...)
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// CheckInstall 检查 cftunnel 是否已安装
func (a *App) CheckInstall() StatusInfo {
	out, err := runCftunnel("version")
	if err != nil {
		return StatusInfo{Installed: false}
	}
	return StatusInfo{
		Installed: true,
		Version:   strings.TrimSpace(out),
	}
}

// GetStatus 获取隧道状态
func (a *App) GetStatus() string {
	out, err := runCftunnel("status")
	if err != nil {
		return "未初始化"
	}
	return strings.TrimSpace(out)
}

// GetRoutes 获取路由列表
func (a *App) GetRoutes() []RouteInfo {
	out, err := runCftunnel("list")
	if err != nil {
		return nil
	}
	return parseRoutes(out)
}

// parseRoutes 解析 cftunnel list 输出
func parseRoutes(output string) []RouteInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var routes []RouteInfo
	for i, line := range lines {
		if i == 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			routes = append(routes, RouteInfo{
				Name:     fields[0],
				Hostname: fields[1],
				Service:  fields[2],
			})
		}
	}
	return routes
}

// StartQuick 启动免域名模式
func (a *App) StartQuick(port string) QuickResult {
	out, err := runCftunnel("quick", port)
	if err != nil {
		return QuickResult{Err: err.Error()}
	}
	return QuickResult{URL: extractTunnelURL(out)}
}

// TunnelUp 启动隧道
func (a *App) TunnelUp() string {
	out, err := runCftunnel("up")
	if err != nil {
		return fmt.Sprintf("错误: %s", out)
	}
	return strings.TrimSpace(out)
}

// TunnelDown 停止隧道
func (a *App) TunnelDown() string {
	out, err := runCftunnel("down")
	if err != nil {
		return fmt.Sprintf("错误: %s", out)
	}
	return strings.TrimSpace(out)
}

// RunCommand 通用命令执行（前端可调用任意 cftunnel 子命令）
func (a *App) RunCommand(args string) string {
	parts := strings.Fields(args)
	out, err := runCftunnel(parts...)
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

// SelectDirectory 打开目录选择对话框
func (a *App) SelectDirectory() string {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择目录",
	})
	if err != nil {
		return ""
	}
	return dir
}

func extractTunnelURL(output string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "trycloudflare.com") {
			for _, word := range strings.Fields(line) {
				if strings.HasPrefix(word, "https://") {
					return word
				}
			}
		}
	}
	return ""
}

// ==================== Relay 模式 ====================

// RelayRuleInfo 中继规则
type RelayRuleInfo struct {
	Name       string `json:"name"`
	Proto      string `json:"proto"`
	LocalPort  int    `json:"local_port"`
	RemotePort int    `json:"remote_port"`
	Domain     string `json:"domain"`
}

// RelayStatusInfo 中继状态
type RelayStatusInfo struct {
	Server  string `json:"server"`
	Running bool   `json:"running"`
	PID     string `json:"pid"`
	Rules   int    `json:"rules"`
}

// GetRelayStatus 获取中继状态
func (a *App) GetRelayStatus() RelayStatusInfo {
	out, err := runCftunnel("relay", "status")
	if err != nil {
		return RelayStatusInfo{}
	}
	return parseRelayStatus(out)
}

// GetRelayRules 获取中继规则列表
func (a *App) GetRelayRules() []RelayRuleInfo {
	out, err := runCftunnel("relay", "list")
	if err != nil {
		return nil
	}
	return parseRelayRules(out)
}

// RelayUp 启动中继
func (a *App) RelayUp() string {
	out, err := runCftunnel("relay", "up")
	if err != nil {
		return fmt.Sprintf("错误: %s", out)
	}
	return strings.TrimSpace(out)
}

// RelayDown 停止中继
func (a *App) RelayDown() string {
	out, err := runCftunnel("relay", "down")
	if err != nil {
		return fmt.Sprintf("错误: %s", out)
	}
	return strings.TrimSpace(out)
}

// RelayAddRule 添加中继规则
func (a *App) RelayAddRule(name, proto string, localPort, remotePort int, domain string) string {
	args := []string{"relay", "add", name, "--proto", proto, "--local", fmt.Sprintf("%d", localPort)}
	if remotePort > 0 {
		args = append(args, "--remote", fmt.Sprintf("%d", remotePort))
	}
	if domain != "" {
		args = append(args, "--domain", domain)
	}
	out, err := runCftunnel(args...)
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

// RelayRemoveRule 删除中继规则
func (a *App) RelayRemoveRule(name string) string {
	out, err := runCftunnel("relay", "remove", name)
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

// RelayInit 初始化中继配置
func (a *App) RelayInit(server, token string) string {
	out, err := runCftunnel("relay", "init", "--server", server, "--token", token)
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

// RelayInstallService 注册中继系统服务
func (a *App) RelayInstallService() string {
	out, err := runCftunnel("relay", "install")
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

// RelayUninstallService 卸载中继系统服务
func (a *App) RelayUninstallService() string {
	out, err := runCftunnel("relay", "uninstall")
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

// GetRelayLogs 获取中继日志（最后 100 行）
func (a *App) GetRelayLogs() string {
	out, err := runCftunnel("relay", "logs")
	if err != nil {
		return fmt.Sprintf("暂无日志\n%s", strings.TrimSpace(out))
	}
	return strings.TrimSpace(out)
}

// RelayServerSetup 远程部署 frps 服务端（仅支持密钥认证）
func (a *App) RelayServerSetup(host string, port int, user, keyPath string, frpsPort int) string {
	args := []string{"relay", "server", "setup", "--host", host, "-p", fmt.Sprintf("%d", port), "--user", user, "--key", keyPath, "--frps-port", fmt.Sprintf("%d", frpsPort)}
	out, err := runCftunnel(args...)
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

// CheckResultInfo 链路检测结果
type CheckResultInfo struct {
	Server        string          `json:"server"`
	ServerOK      bool            `json:"server_ok"`
	ServerLatency int64           `json:"server_latency_ms"`
	FrpcRunning   bool            `json:"frpc_running"`
	FrpcPID       int             `json:"frpc_pid"`
	Rules         []RuleCheckInfo `json:"rules"`
	Total         int             `json:"total"`
	Passed        int             `json:"passed"`
	Failed        int             `json:"failed"`
}

// RuleCheckInfo 单条规则检测结果
type RuleCheckInfo struct {
	Name       string `json:"name"`
	Proto      string `json:"proto"`
	LocalPort  int    `json:"local_port"`
	RemotePort int    `json:"remote_port"`
	LocalOK    bool   `json:"local_ok"`
	RemoteOK   bool   `json:"remote_ok"`
	LatencyMS  int64  `json:"latency_ms"`
	LocalErr   string `json:"local_err"`
	RemoteErr  string `json:"remote_err"`
}

// RelayCheck 执行链路检测
func (a *App) RelayCheck() CheckResultInfo {
	out, err := runCftunnel("relay", "check", "--json")
	if err != nil {
		return CheckResultInfo{}
	}
	var result CheckResultInfo
	if json.Unmarshal([]byte(out), &result) != nil {
		return CheckResultInfo{}
	}
	return result
}

// parseRelayStatus 解析 relay status 输出
func parseRelayStatus(output string) RelayStatusInfo {
	info := RelayStatusInfo{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "服务器:") || strings.HasPrefix(line, "服务器：") {
			info.Server = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			if strings.Contains(line, "：") {
				info.Server = strings.TrimSpace(strings.SplitN(line, "：", 2)[1])
			}
		} else if strings.HasPrefix(line, "状态:") || strings.HasPrefix(line, "状态：") {
			if strings.Contains(line, "运行中") {
				info.Running = true
				// 提取 PID
				if idx := strings.Index(line, "PID:"); idx >= 0 {
					pid := strings.TrimSpace(line[idx+4:])
					pid = strings.TrimRight(pid, ")")
					info.PID = pid
				}
			}
		} else if strings.HasPrefix(line, "规则数:") || strings.HasPrefix(line, "规则数：") {
			fmt.Sscanf(line, "规则数: %d", &info.Rules)
			if info.Rules == 0 {
				fmt.Sscanf(line, "规则数：%d", &info.Rules)
			}
		}
	}
	return info
}

// parseRelayRules 解析 relay list 输出
func parseRelayRules(output string) []RelayRuleInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || strings.Contains(output, "暂无中继规则") {
		return nil
	}
	var rules []RelayRuleInfo
	for i, line := range lines {
		// 跳过表头和分隔线
		if i < 2 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		var localPort, remotePort int
		fmt.Sscanf(fields[2], "%d", &localPort)
		if fields[3] != "-" {
			fmt.Sscanf(fields[3], "%d", &remotePort)
		}
		domain := ""
		if len(fields) >= 5 && fields[4] != "-" {
			domain = fields[4]
		}
		rules = append(rules, RelayRuleInfo{
			Name:       fields[0],
			Proto:      fields[1],
			LocalPort:  localPort,
			RemotePort: remotePort,
			Domain:     domain,
		})
	}
	return rules
}

// UpdateInfo 更新检测结果
type UpdateInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	ReleaseURL     string `json:"release_url"`
	Err            string `json:"err,omitempty"`
}

// GetAppVersion 获取客户端版本
func (a *App) GetAppVersion() string {
	return AppVersion
}

// CheckAppUpdate 检查客户端更新
func (a *App) CheckAppUpdate() UpdateInfo {
	info := UpdateInfo{CurrentVersion: AppVersion}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/qingchencloud/cftunnel-app/releases/latest")
	if err != nil {
		info.Err = "网络请求失败"
		return info
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		info.Err = "读取响应失败"
		return info
	}
	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if json.Unmarshal(body, &release) != nil {
		info.Err = "解析响应失败"
		return info
	}
	info.LatestVersion = strings.TrimPrefix(release.TagName, "v")
	info.ReleaseURL = release.HTMLURL
	info.HasUpdate = info.LatestVersion != "" && info.LatestVersion != strings.TrimPrefix(AppVersion, "v")
	return info
}
