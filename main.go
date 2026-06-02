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

	// 先建聚合器再构造依赖它的 service; 拿到 app 后回填以便发事件
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
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})
	agg.SetApp(app)

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

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "AgentLamp",
		Width:            360,
		Height:           640,
		BackgroundColour: application.NewRGB(10, 10, 15),
		URL:              "/",
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = hookSrv.Stop(ctx)
	agg.Stop()
}
