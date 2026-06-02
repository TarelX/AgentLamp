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
	"github.com/wailsapp/wails/v3/pkg/events"
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
	winSvc := service.NewWindowService()

	app := application.New(application.Options{
		Name:        "AgentLamp",
		Description: "跨平台 AI Agent 状态灯 - 程序员的过街信号",
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "io.github.tarelx.agentlamp",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				logger.Info("second instance launched, focusing existing main window", "args", data.Args)
			},
		},
		Services: []application.Service{
			application.NewService(service.NewStatusService(agg)),
			application.NewService(installSvc),
			application.NewService(winSvc),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})
	agg.SetApp(app)

	mainWin := buildMainWindow(app)
	lampWin := buildLampWindow(app)
	winSvc.SetWindows(mainWin, lampWin)

	// 主窗口的关闭按钮 = 退出整个应用 (与 Windows 用户预期一致);
	// 切换悬浮模式时主窗口走 Hide() 路径不会触发 WindowClosing
	mainWin.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		app.Quit()
	})

	tray := buildSystemTray(app, agg, winSvc, mainWin, lampWin)
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

// buildMainWindow 默认显示的完整 UI 窗口, 普通窗口装饰
func buildMainWindow(app *application.App) *application.WebviewWindow {
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "AgentLamp",
		Width:            420,
		Height:           720,
		MinWidth:         360,
		MinHeight:        560,
		URL:              "/",
		BackgroundColour: application.NewRGB(10, 10, 15),
	})
}

// buildLampWindow 透明置顶悬浮小灯, 锁死尺寸防被拖拽 resize.
//
// 不启用 Windows.WindowMask: Wails v3 当前版本的 setWindowMask 在
// UpdateLayeredWindow 调用时把 size 当 position 传, 导致 mask 与 webview 错位.
func buildLampWindow(app *application.App) *application.WebviewWindow {
	const w, h = 220, 380
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:               "AgentLamp · Floating",
		Width:               w,
		Height:              h,
		MinWidth:            w,
		MinHeight:           h,
		MaxWidth:            w,
		MaxHeight:           h,
		URL:                 "/?mode=lamp",
		Frameless:           true,
		AlwaysOnTop:         true,
		Hidden:              true,
		BackgroundType:      application.BackgroundTypeTransparent,
		BackgroundColour:    application.NewRGBA(0, 0, 0, 0),
		MinimiseButtonState: application.ButtonHidden,
		MaximiseButtonState: application.ButtonHidden,
		CloseButtonState:    application.ButtonHidden,
		Windows: application.WindowsWindow{
			DisableFramelessWindowDecorations: true,
			HiddenOnTaskbar:                   true,
		},
		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTransparent,
			TitleBar: application.MacTitleBarHidden,
		},
	})
}

// buildSystemTray 托盘图标 + 右键菜单, 订阅状态变化更新颜色
func buildSystemTray(
	app *application.App,
	agg *aggregator.Aggregator,
	winSvc *service.WindowService,
	mainWin *application.WebviewWindow,
	lampWin *application.WebviewWindow,
) *application.SystemTray {
	tray := app.SystemTray.New()
	tray.SetTooltip("AgentLamp")
	tray.SetIcon(icon.PNGForState(agg.Snapshot().MainState))

	menu := application.NewMenu()
	menu.Add("打开主窗口").OnClick(func(*application.Context) {
		winSvc.SwitchToFull()
	})
	menu.Add("切换悬浮模式").OnClick(func(*application.Context) {
		if winSvc.CurrentMode() == string(service.ModeFloating) {
			winSvc.SwitchToFull()
		} else {
			winSvc.SwitchToFloating()
		}
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(*application.Context) {
		app.Quit()
	})
	tray.SetMenu(menu)

	tray.OnClick(func() {
		if mainWin.IsVisible() {
			mainWin.Hide()
		} else if lampWin.IsVisible() {
			lampWin.Hide()
		} else {
			winSvc.SwitchToFull()
		}
	})

	agg.Subscribe(func(snap backend.Snapshot) {
		tray.SetIcon(icon.PNGForState(snap.MainState))
	})
	return tray
}
