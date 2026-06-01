// Package service 暴露给前端的 Wails service.
package service

import (
	"github.com/TarelX/AgentLamp/backend"
	"github.com/TarelX/AgentLamp/backend/aggregator"
)

// StatusService 由 main.go 注册到 Wails. 方法名首字母大写自动绑定到前端
type StatusService struct {
	agg *aggregator.Aggregator
}

func NewStatusService(agg *aggregator.Aggregator) *StatusService {
	return &StatusService{agg: agg}
}

// GetSnapshot 前端拉取一次完整状态, 通常在启动时调用
func (s *StatusService) GetSnapshot() backend.Snapshot {
	return s.agg.Snapshot()
}

// SetAgentEnabled 用户在设置面板切换 agent 开关
func (s *StatusService) SetAgentEnabled(name string, enabled bool) backend.Snapshot {
	s.agg.SetEnabled(backend.AgentName(name), enabled)
	return s.agg.Snapshot()
}
