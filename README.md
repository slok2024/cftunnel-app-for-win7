# cftunnel-app

[![GitHub release](https://img.shields.io/github/v/release/qingchencloud/cftunnel-app)](https://github.com/qingchencloud/cftunnel-app/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**cftunnel 桌面客户端** — 基于 [Wails](https://wails.io) 构建的跨平台 GUI。

[cftunnel CLI](https://github.com/qingchencloud/cftunnel) 的可视化管理界面，让 Cloudflare Tunnel 内网穿透操作更直观。

## 功能

- **仪表盘** — 隧道状态一目了然，一键启停
- **免域名模式** — 输入端口即可生成 `*.trycloudflare.com` 临时公网地址
- **路由管理** — 可视化添加/删除路由，自动创建 DNS 记录
- **内置终端** — 直接执行 cftunnel 子命令，无需切换窗口

## 截图

> 深色主题 · macOS 毛玻璃标题栏 · SVG 系统图标

## 下载安装

从 [GitHub Releases](https://github.com/qingchencloud/cftunnel-app/releases) 下载对应平台的安装包：

| 平台 | 文件 |
|------|------|
| macOS | `cftunnel-app-macos.zip` |
| Windows | `cftunnel-app-windows.zip` |
| Linux | `cftunnel-app-linux.tar.gz` |

## 前置条件

需要先安装 [cftunnel CLI](https://github.com/qingchencloud/cftunnel)：

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/qingchencloud/cftunnel/main/install.sh | bash

# Windows (PowerShell)
irm https://raw.githubusercontent.com/qingchencloud/cftunnel/main/install.ps1 | iex
```

## 架构

```
┌─────────────────────────────────┐
│  React + TypeScript (前端 UI)    │
├─────────────────────────────────┤
│  Wails v2 (Go ↔ JS 桥接)       │
├─────────────────────────────────┤
│  Go 后端 (exec 调用 cftunnel)   │
├─────────────────────────────────┤
│  cftunnel CLI (本机已安装)       │
└─────────────────────────────────┘
```

桌面客户端通过 `exec` 调用本机 cftunnel CLI，完全独立不耦合。

## 开发

```bash
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 检查环境
wails doctor

# 开发模式（热重载）
wails dev

# 构建
wails build

# 运行测试
go test -v ./...
```

## 构建产物

| 平台 | 产物 | 大小 |
|------|------|------|
| macOS | `build/bin/cftunnel-app.app` | ~8MB |
| Windows | `build/bin/cftunnel-app.exe` | ~8MB |
| Linux | `build/bin/cftunnel-app` | ~8MB |

## 技术栈

- **后端**: Go + Wails v2
- **前端**: React + TypeScript + Vite
- **图标**: 内联 SVG (Lucide 风格，零依赖)
- **CI**: GitHub Actions 三平台自动构建

## 关联项目

- [cftunnel](https://github.com/qingchencloud/cftunnel) — CLI 工具（本客户端的核心依赖）
- [cftunnel 官网](https://cftunnel.qt.cool) — 产品介绍与下载
- [社区讨论](https://linux.do/t/1636467) — Linux.do 讨论帖
- [QQ 交流群](https://qm.qq.com/q/qUfdR0jJVS) — OpenClaw 交流群

## License

MIT

---

由 [武汉晴辰天下网络科技有限公司](https://qingchencloud.com) 开源维护
