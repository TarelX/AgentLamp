// AgentLamp 主入口
package main

import (
	"context"
	"embed"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/TarelX/AgentLamp/backend"
	"github.com/TarelX/AgentLamp/backend/adapters"
	"github.com/TarelX/AgentLamp/backend/aggregator"
	"github.com/TarelX/AgentLamp/backend/icon"
	"github.com/TarelX/AgentLamp/backend/server"
	"github.com/TarelX/AgentLamp/backend/service"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

const webhookAddr = "127.0.0.1:19840"

func init() {
	application.RegisterEvent[backend.Snapshot](backend.EventStatusUpdate)
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	agg := aggregator.New(nil, []backend.AgentName{
		backend.AgentClaude,
		backend.AgentCursor,
		backend.AgentCodex,
	})

	installSvc, err := service.NewInstallService("http://" + webhookAddr)
	if err != nil {
		log.Fatalf("init install service: %v", err)
	}

	app := application.New(application.Options{
		Name:        "AgentLamp",
		Description: "跨平台 AI Agent 状态灯 - 程序员的过街信号",
		Services: []application.Service{
			application.NewService(service.NewStatusService(agg)),
			application.NewService(installSvc),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})
	agg.SetApp(app)

	lamp := buildLampWindow(app)
	settings := buildSettingsWindow(app)
	tray := buildSystemTray(app, agg, lamp, settings)
	_ = tray

	hookSrv := server.New(webhookAddr, logger)
	hookSrv.Register(backend.AgentClaude, adapters.NewClaudeAdapter(agg))
	hookSrv.Register(backend.AgentCursor, adapters.NewCursorAdapter(agg))

	codex, err := adapters.NewCodexWatcher(agg, 3*time.Second, logger)
	if err != nil {
		logger.Warn("codex watcher init failed", "err", err)
	} else if err := codex.Start(); err != nil {
		logger.Warn("codex watcher start failed", "err", err)
	}

	go agg.Run()
	go func() {
		if err := hookSrv.Start(); err != nil {
			logger.Error("webhook server stopped", "err", err)
		}
	}()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = hookSrv.Stop(ctx)
	agg.Stop()
}

// buildLampWindow 构造主窗口: 透明置顶可拖动小窗, 仅渲染物理灯
func buildLampWindow(app *application.App) *application.WebviewWindow {
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:               "AgentLamp",
		Width:               240,
		Height:              360,
		MinWidth:            200,
		MinHeight:           300,
		URL:                 "/",
		Frameless:           true,
		AlwaysOnTop:         true,
		BackgroundType:      application.BackgroundTypeTransparent,
		BackgroundColour:    application.NewRGBA(0, 0, 0, 0),
		MinimiseButtonState: application.ButtonHidden,
		MaximiseButtonState: application.ButtonHidden,
		CloseButtonState:    application.ButtonHidden,
		InitialPosition:     application.WindowCentered,
		Windows: application.WindowsWindow{
			DisableFramelessWindowDecorations: true,
		},
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 0,
			Backdrop:                application.MacBackdropTransparent,
			TitleBar:                application.MacTitleBarHidden,
		},
	})
}

// buildSettingsWindow 构造独立设置窗口, 启动时隐藏, 由托盘菜单触发显示
func buildSettingsWindow(app *application.App) *application.WebviewWindow {
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "AgentLamp · 设置",
		Width:            860,
		Height:           640,
		MinWidth:         640,
		MinHeight:        480,
		URL:              "/settings",
		Hidden:           true,
		BackgroundColour: application.NewRGB(10, 10, 15),
	})
}

// buildSystemTray 创建托盘图标 + 右键菜单, 并订阅主灯状态变化更新图标
func buildSystemTray(
	app *application.App,
	agg *aggregator.Aggregator,
	lamp *application.WebviewWindow,
	settings *application.WebviewWindow,
) *application.SystemTray {
	tray := app.SystemTray.New()
	tray.SetTooltip("AgentLamp")
	tray.SetIcon(icon.PNGForState(agg.Snapshot().MainState))
	tray.AttachWindow(lamp)

	menu := application.NewMenu()
	menu.Add("显示 / 隐藏主灯").OnClick(func(*application.Context) {
		if lamp.IsVisible() {
			lamp.Hide()
		} else {
			lamp.Show().Focus()
		}
	})
	menu.Add("打开设置…").OnClick(func(*application.Context) {
		settings.Show().Focus()
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(*application.Context) {
		app.Quit()
	})
	tray.SetMenu(menu)

	agg.Subscribe(func(snap backend.Snapshot) {
		tray.SetIcon(icon.PNGForState(snap.MainState))
	})
	return tray
}
