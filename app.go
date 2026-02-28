package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// AppVersion 由编译时的 ldflags 注入，默认为 dev
var AppVersion = "dev"

type App struct {
	ctx      context.Context
	quickMu  sync.Mutex
	quickCmd *exec.Cmd
	quickURL string
}

func NewApp() *App {
	return &App{}
}

// startup: 程序启动时调用
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 环境预检查：静默杀掉可能存在的残留进程
	killProcessByName("cftunnel.exe")
	killProcessByName("cloudflared.exe")
}

// shutdown: 程序关闭时调用
func (a *App) shutdown(ctx context.Context) {
	a.quickMu.Lock()
	if a.quickCmd != nil && a.quickCmd.Process != nil {
		_ = quickProcessKill(a.quickCmd.Process.Pid)
	}
	a.quickMu.Unlock()

	// 退出清理
	killProcessByName("cftunnel.exe")
	killProcessByName("cloudflared.exe")
}

// --- 核心静默工具函数：解决黑窗口闪烁 ---

func hideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	// Windows下隐藏窗口的核心设置
	cmd.SysProcAttr.HideWindow = true
	// CREATE_NO_WINDOW: 0x08000000
	// 在Win7中，如果依然无法启动，可尝试将此值设为 0
	cmd.SysProcAttr.CreationFlags = 0x08000000
}

func killProcessByName(imageName string) {
	cmd := exec.Command("taskkill", "/F", "/T", "/IM", imageName)
	hideWindow(cmd)
	_ = cmd.Run()
}

func quickProcessKill(pid int) error {
	if pid <= 0 {
		return nil
	}
	cmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	hideWindow(cmd)
	return cmd.Run()
}

func quickProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

// --- 业务逻辑 ---

func (a *App) QuickStop() string {
	a.quickMu.Lock()
	cmd := a.quickCmd
	a.quickMu.Unlock()

	_ = os.Remove(quickURLPath())

	var targetPid int
	if cmd != nil && cmd.Process != nil {
		targetPid = cmd.Process.Pid
	} else {
		pidData, err := os.ReadFile(quickPIDPath())
		if err == nil {
			pid, _ := strconv.Atoi(strings.TrimSpace(string(pidData)))
			targetPid = pid
		}
	}

	if targetPid > 0 {
		_ = quickProcessKill(targetPid)
		for i := 0; i < 10; i++ {
			if !quickProcessAlive(targetPid) {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	_ = os.Remove(quickPIDPath())
	_ = os.Remove(quickURLPath())

	a.quickMu.Lock()
	a.quickCmd = nil
	a.quickURL = ""
	a.quickMu.Unlock()

	return "隧道已停止"
}

type StatusInfo struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Output    string `json:"output"`
}

type RouteInfo struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Service  string `json:"service"`
}

type QuickResult struct {
	URL string `json:"url"`
	Err string `json:"err"`
}

var cftunnelBin string

// 修复点：增强 Win7 下的路径查找逻辑
func findCftunnel() string {
	if cftunnelBin != "" {
		return cftunnelBin
	}

	// 1. 尝试获取当前运行程序目录下的绝对路径
	if exePath, err := os.Executable(); err == nil {
		dir, _ := filepath.Abs(filepath.Dir(exePath))
		p := filepath.Join(dir, "cftunnel.exe")
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			cftunnelBin = p
			return p
		}
	}

	// 2. 尝试从系统 PATH 中查找
	if p, err := exec.LookPath("cftunnel.exe"); err == nil {
		absP, _ := filepath.Abs(p)
		cftunnelBin = absP
		return absP
	}

	// 3. 兜底返回，交给 exec.Command 报错
	return "cftunnel.exe"
}

func runCftunnel(args ...string) (string, error) {
	bin := findCftunnel()
	cmd := exec.Command(bin, args...)
	
	// 显式设置工作目录为内核所在目录
	// 这能保证内核里的 "." 永远指向它自己所在的文件夹
	cmd.Dir = filepath.Dir(bin) 
	
	hideWindow(cmd)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (a *App) CheckInstall() StatusInfo {
	out, err := runCftunnel("version")
	if err != nil {
		// 修复点：返回具体错误信息，方便在 Win7 UI 上排查
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}
		return StatusInfo{
			Installed: false,
			Output:    fmt.Sprintf("Exec error: %s, Output: %s", errMsg, out),
		}
	}
	return StatusInfo{
		Installed: true,
		Version:   strings.TrimSpace(out),
	}
}

func (a *App) GetStatus() string {
	out, err := runCftunnel("status")
	if err != nil {
		return "未初始化"
	}
	return strings.TrimSpace(out)
}

func (a *App) GetRoutes() []RouteInfo {
	out, err := runCftunnel("list")
	if err != nil {
		return nil
	}
	return parseRoutes(out)
}

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

func (a *App) StartQuick(port string) QuickResult {
	a.quickMu.Lock()
	if a.quickCmd != nil && a.quickCmd.Process != nil {
		a.quickMu.Unlock()
		return QuickResult{Err: "隧道已在运行，请先停止"}
	}
	a.quickMu.Unlock()

	var binPath string
	// 优先找同级目录，并转为绝对路径
	if exePath, err := os.Executable(); err == nil {
		dir, _ := filepath.Abs(filepath.Dir(exePath))
		p := filepath.Join(dir, "cloudflared.exe")
		if _, err := os.Stat(p); err == nil {
			binPath = p
		}
	}
	
	if binPath == "" {
		home, _ := os.UserHomeDir()
		binPath = filepath.Join(home, ".cftunnel", "cloudflared.exe")
	}

	cmd := exec.Command(binPath, "tunnel", "--url", "http://localhost:"+port)
	hideWindow(cmd)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return QuickResult{Err: "创建管道失败: " + err.Error()}
	}

	if err := cmd.Start(); err != nil {
		return QuickResult{Err: "启动失败: " + err.Error()}
	}

	a.quickMu.Lock()
	a.quickCmd = cmd
	a.quickURL = ""
	a.quickMu.Unlock()

	pidPath := quickPIDPath()
	home, _ := os.UserHomeDir()
	_ = os.MkdirAll(filepath.Join(home, ".cftunnel"), 0700)
	_ = os.WriteFile(pidPath, []byte(strconv.Itoa(cmd.Process.Pid)), 0600)

	go a.scanQuickURL(stderr)

	go func() {
		_ = cmd.Wait()
		a.quickMu.Lock()
		a.quickCmd = nil
		a.quickURL = ""
		a.quickMu.Unlock()
		_ = os.Remove(pidPath)
		_ = os.Remove(quickURLPath())
	}()

	for i := 0; i < 15; i++ { // Win7 启动较慢，增加等待时间
		time.Sleep(500 * time.Millisecond)
		a.quickMu.Lock()
		url := a.quickURL
		a.quickMu.Unlock()
		if url != "" {
			return QuickResult{URL: url}
		}
	}
	return QuickResult{URL: ""}
}

func quickURLPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cftunnel", "quick.url")
}

func quickPIDPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cftunnel", "quick.pid")
}

func (a *App) scanQuickURL(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "trycloudflare.com") {
			url := extractTunnelURL(line + "\n")
			if url != "" {
				a.quickMu.Lock()
				a.quickURL = url
				a.quickMu.Unlock()
				_ = os.WriteFile(quickURLPath(), []byte(url), 0600)
			}
		}
	}
}

func (a *App) QuickRunning() bool {
	a.quickMu.Lock()
	running := a.quickCmd != nil
	a.quickMu.Unlock()
	if running {
		return true
	}
	pidData, err := os.ReadFile(quickPIDPath())
	if err != nil {
		return false
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(string(pidData)))
	return quickProcessAlive(pid)
}

func (a *App) QuickURL() string {
	a.quickMu.Lock()
	u := a.quickURL
	a.quickMu.Unlock()
	if u != "" {
		return u
	}
	data, err := os.ReadFile(quickURLPath())
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data))
	}
	return ""
}

func (a *App) TunnelUp() string {
	out, err := runCftunnel("up")
	if err != nil {
		return fmt.Sprintf("错误: %s", out)
	}
	return strings.TrimSpace(out)
}

func (a *App) TunnelDown() string {
	out, err := runCftunnel("down")
	if err != nil {
		return fmt.Sprintf("错误: %s", out)
	}
	return strings.TrimSpace(out)
}

func (a *App) RunCommand(args string) string {
	parts := strings.Fields(args)
	out, err := runCftunnel(parts...)
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

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

// ==================== Relay 相关 ====================

type RelayRuleInfo struct {
	Name       string `json:"name"`
	Proto      string `json:"proto"`
	LocalPort  int    `json:"local_port"`
	RemotePort int    `json:"remote_port"`
	Domain     string `json:"domain"`
}

type RelayStatusInfo struct {
	Server  string `json:"server"`
	Running bool   `json:"running"`
	PID     string `json:"pid"`
	Rules   int    `json:"rules"`
}

func (a *App) GetRelayStatus() RelayStatusInfo {
	out, err := runCftunnel("relay", "status")
	if err != nil {
		return RelayStatusInfo{}
	}
	return parseRelayStatus(out)
}

func (a *App) GetRelayRules() []RelayRuleInfo {
	out, err := runCftunnel("relay", "list")
	if err != nil {
		return nil
	}
	return parseRelayRules(out)
}

func (a *App) RelayUp() string {
	out, err := runCftunnel("relay", "up")
	if err != nil {
		return fmt.Sprintf("错误: %s", out)
	}
	return strings.TrimSpace(out)
}

func (a *App) RelayDown() string {
	out, err := runCftunnel("relay", "down")
	if err != nil {
		return fmt.Sprintf("错误: %s", out)
	}
	return strings.TrimSpace(out)
}

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

func (a *App) RelayRemoveRule(name string) string {
	out, err := runCftunnel("relay", "remove", name)
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

func (a *App) RelayInit(server, token string) string {
	out, err := runCftunnel("relay", "init", "--server", server, "--token", token)
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

func (a *App) RelayInstallService() string {
	out, err := runCftunnel("relay", "install")
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

func (a *App) RelayUninstallService() string {
	out, err := runCftunnel("relay", "uninstall")
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

func (a *App) GetRelayLogs() string {
	out, err := runCftunnel("relay", "logs")
	if err != nil {
		return fmt.Sprintf("暂无日志\n%s", strings.TrimSpace(out))
	}
	return strings.TrimSpace(out)
}

func (a *App) RelayServerSetup(host string, port int, user, keyPath, password string, frpsPort int) string {
	args := []string{"relay", "server", "setup", "--host", host, "-p", fmt.Sprintf("%d", port), "--user", user, "--frps-port", fmt.Sprintf("%d", frpsPort)}
	if password != "" {
		args = append(args, "--pass", password)
	} else if keyPath != "" {
		args = append(args, "--key", keyPath)
	}
	out, err := runCftunnel(args...)
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, out)
	}
	return strings.TrimSpace(out)
}

func (a *App) RelayCheck() CheckResultInfo {
	out, err := runCftunnel("relay", "check", "--json")
	if err != nil {
		return CheckResultInfo{}
	}
	var result CheckResultInfo
	_ = json.Unmarshal([]byte(out), &result)
	return result
}

func parseRelayStatus(output string) RelayStatusInfo {
	info := RelayStatusInfo{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "服务器:") || strings.HasPrefix(line, "服务器：") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) < 2 {
				parts = strings.SplitN(line, "：", 2)
			}
			if len(parts) >= 2 {
				info.Server = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "状态:") || strings.HasPrefix(line, "状态：") {
			if strings.Contains(line, "运行中") {
				info.Running = true
				if idx := strings.Index(line, "PID:"); idx >= 0 {
					pid := strings.TrimSpace(line[idx+4:])
					pid = strings.TrimRight(pid, ")")
					info.PID = pid
				}
			}
		} else if strings.HasPrefix(line, "规则数:") || strings.HasPrefix(line, "规则数：") {
			_, _ = fmt.Sscanf(line, "规则数: %d", &info.Rules)
			if info.Rules == 0 {
				_, _ = fmt.Sscanf(line, "规则数：%d", &info.Rules)
			}
		}
	}
	return info
}

func parseRelayRules(output string) []RelayRuleInfo {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || strings.Contains(output, "暂无中继规则") {
		return nil
	}
	var rules []RelayRuleInfo
	for i, line := range lines {
		if i < 2 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		var localPort, remotePort int
		_, _ = fmt.Sscanf(fields[2], "%d", &localPort)
		if fields[3] != "-" {
			_, _ = fmt.Sscanf(fields[3], "%d", &remotePort)
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

// ==================== 更新与版本 ====================

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

type UpdateInfo struct {
	CurrentVersion string `json:"current_version"`
	LatestVersion  string `json:"latest_version"`
	HasUpdate      bool   `json:"has_update"`
	ReleaseURL     string `json:"release_url"`
	Err            string `json:"err,omitempty"`
}

func (a *App) GetAppVersion() string {
	return AppVersion
}

func (a *App) CheckAppUpdate() UpdateInfo {
	info := UpdateInfo{CurrentVersion: AppVersion}
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.github.com/repos/qingchencloud/cftunnel-app/releases/latest")
	if err != nil {
		info.Err = "网络请求失败"
		return info
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
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
	info.HasUpdate = false 
	return info
}