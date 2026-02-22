# cftunnel-app

cftunnel 桌面客户端 — 基于 [Wails](https://wails.io) 构建的跨平台 GUI。

## 功能

- 仪表盘：隧道状态一目了然，一键启停
- 免域名模式：输入端口即可生成临时公网地址
- 路由管理：可视化添加/删除路由
- 内置终端：直接执行 cftunnel 命令

## 前置条件

需要先安装 [cftunnel CLI](https://github.com/qingchencloud/cftunnel)：

```bash
curl -fsSL https://raw.githubusercontent.com/qingchencloud/cftunnel/main/install.sh | bash
```

## 开发

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
wails dev    # 开发模式（热重载）
wails build  # 构建
```

## 构建产物

| 平台 | 产物 |
|------|------|
| macOS | `build/bin/cftunnel-app.app` |
| Windows | `build/bin/cftunnel-app.exe` |
| Linux | `build/bin/cftunnel-app` |

## 技术栈

- 后端：Go + Wails v2
- 前端：React + TypeScript + Vite
- CI：GitHub Actions 三平台自动构建

## License

MIT
