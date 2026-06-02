// Package adapters 把各 agent 的原始 hook payload 翻译成统一的 AgentStatus.
package adapters

import (
	"encoding/json"
	"time"

	"github.com/TarelX/AgentLamp/backend"
	"github.com/TarelX/AgentLamp/backend/aggregator"
)

// claudeHookInput Claude Code hook 通过 stdin 推送的标准字段
type claudeHookInput struct {
	SessionID      string `json:"session_id,omitempty"`
	TranscriptPath string `json:"transcript_path,omitempty"`
	CWD            string `json:"cwd,omitempty"`
	PermissionMode string `json:"permission_mode,omitempty"`
	HookEventName  string `json:"hook_event_name,omitempty"`
}

// ClaudeAdapter 处理路径 /hook/claude/<event>
type ClaudeAdapter struct {
	agg *aggregator.Aggregator
}

func NewClaudeAdapter(agg *aggregator.Aggregator) *ClaudeAdapter {
	return &ClaudeAdapter{agg: agg}
}

// HandleHook 由 webhook server 在收到请求时调用; payload 是 hook 原始 stdin JSON
func (c *ClaudeAdapter) HandleHook(event string, payload []byte) ([]byte, error) {
	var in claudeHookInput
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &in)
	}
	if in.HookEventName == "" {
		in.HookEventName = event
	}

	state := mapClaudeEventToState(in.HookEventName)
	c.agg.Push(backend.AgentStatus{
		Name:       backend.AgentClaude,
		State:      state,
		Enabled:    true,
		Message:    in.CWD,
		LastUpdate: time.Now().UnixMilli(),
	})
	return []byte(`{"continue":true}`), nil
}

// mapClaudeEventToState Claude 官方事件到主灯状态的映射;
// SessionStart 视为 idle, 真正进入 running 由 UserPromptSubmit 驱动
func mapClaudeEventToState(event string) backend.AggregatedState {
	switch event {
	case "Notification":
		return backend.StateWaiting
	case "UserPromptSubmit", "PreToolUse", "PostToolUse":
		return backend.StateRunning
	case "Stop", "SubagentStop", "SessionEnd", "SessionStart":
		return backend.StateIdle
	default:
		return backend.StateRunning
	}
}
