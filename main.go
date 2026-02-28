package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "cftunnel",
		Width:            960,
		Height:           640,
		MinWidth:         800,
		MinHeight:        500,
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 15, A: 255},
		AssetServer:      &assetserver.Options{Assets: assets},
		
		// --- 生命周期钩子 ---
		OnStartup:  app.startup,  // 启动时清理旧进程
		OnShutdown: app.shutdown, // 关闭时清理当前进程
		
		Bind: []interface{}{app},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarHiddenInset(),
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}