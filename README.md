这个项目本质上是一个基于 Go 语言和 Wails 框架开发的 Cloudflare Tunnel 桌面管理客户端。

它将复杂的命令行操作封装成了直观、简洁的 Windows 桌面应用，旨在为用户提供“一键式”的内网穿透解决方案。

通过 AI 修改官方代码，将程序做成绿色版；

编译支持win7版本的内核及主程序；

集成cftunnel.exe和cloudflared.exe和frpc.exe三个内核文件；

删除掉官方的下载内核逻辑；

修复官方一处停止临时隧道 bug；

去掉程序升级检测。
