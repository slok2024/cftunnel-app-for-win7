package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

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

// CheckInstall 检查 cftunnel 是否已安装
func (a *App) CheckInstall() StatusInfo {
	out, err := exec.Command("cftunnel", "version").CombinedOutput()
	if err != nil {
		return StatusInfo{Installed: false}
	}
	return StatusInfo{
		Installed: true,
		Version:   strings.TrimSpace(string(out)),
	}
}

// GetStatus 获取隧道状态
func (a *App) GetStatus() string {
	out, err := exec.Command("cftunnel", "status").CombinedOutput()
	if err != nil {
		return "未初始化"
	}
	return strings.TrimSpace(string(out))
}

// GetRoutes 获取路由列表
func (a *App) GetRoutes() []RouteInfo {
	out, err := exec.Command("cftunnel", "list").CombinedOutput()
	if err != nil {
		return nil
	}
	return parseRoutes(string(out))
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
	out, err := runWithTimeout("cftunnel", "quick", port)
	if err != nil {
		// quick 模式是前台阻塞的，这里用后台方式
		return QuickResult{Err: err.Error()}
	}
	return QuickResult{URL: extractTunnelURL(out)}
}

// TunnelUp 启动隧道
func (a *App) TunnelUp() string {
	out, err := exec.Command("cftunnel", "up").CombinedOutput()
	if err != nil {
		return fmt.Sprintf("错误: %s", string(out))
	}
	return strings.TrimSpace(string(out))
}

// TunnelDown 停止隧道
func (a *App) TunnelDown() string {
	out, err := exec.Command("cftunnel", "down").CombinedOutput()
	if err != nil {
		return fmt.Sprintf("错误: %s", string(out))
	}
	return strings.TrimSpace(string(out))
}

// RunCommand 通用命令执行（前端可调用任意 cftunnel 子命令）
func (a *App) RunCommand(args string) string {
	parts := strings.Fields(args)
	out, err := exec.Command("cftunnel", parts...).CombinedOutput()
	if err != nil {
		return fmt.Sprintf("错误: %s\n%s", err, string(out))
	}
	return strings.TrimSpace(string(out))
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

func runWithTimeout(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
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
