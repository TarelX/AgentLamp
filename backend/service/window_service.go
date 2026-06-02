package service

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// WindowMode 当前界面模式: full = 普通窗口含完整 UI; floating = 透明置顶悬浮小灯
type WindowMode string

const (
	ModeFull     WindowMode = "full"
	ModeFloating WindowMode = "floating"
)

// WindowService 管理两个窗口的可见性切换, 由前端"悬浮模式"开关调用
type WindowService struct {
	main *application.WebviewWindow
	lamp *application.WebviewWindow
}

func NewWindowService() *WindowService {
	return &WindowService{}
}

// SetWindows 在 app 创建后注入两个窗口引用; service 注册顺序问题的解法
func (w *WindowService) SetWindows(main, lamp *application.WebviewWindow) {
	w.main = main
	w.lamp = lamp
}

// CurrentMode 返回当前可见的窗口模式; 两个都不可见则视为 full (启动初始)
func (w *WindowService) CurrentMode() string {
	if w.lamp != nil && w.lamp.IsVisible() {
		return string(ModeFloating)
	}
	return string(ModeFull)
}

// SwitchToFloating 隐藏主窗口, 显示悬浮灯
func (w *WindowService) SwitchToFloating() string {
	if w.main != nil {
		w.main.Hide()
	}
	if w.lamp != nil {
		w.lamp.Show().Focus()
	}
	return string(ModeFloating)
}

// SwitchToFull 隐藏悬浮灯, 显示主窗口
func (w *WindowService) SwitchToFull() string {
	if w.lamp != nil {
		w.lamp.Hide()
	}
	if w.main != nil {
		w.main.Show().Focus()
	}
	return string(ModeFull)
}
