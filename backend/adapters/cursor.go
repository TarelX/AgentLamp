package adapters

import (
	"encoding/json"
	"time"

	"github.com/TarelX/AgentLamp/backend"
	"github.com/TarelX/AgentLamp/backend/aggregator"
)

// cursorHookInput Cursor IDE / CLI hook 通过 stdin 推送的字段, 不同事件字段不同, 取交集
type cursorHookInput struct {
	ConversationID string   `json:"conversation_id,omitempty"`
	GenerationID   string   `json:"generation_id,omitempty"`
	HookEventName  string   `json:"hook_event_name,omitempty"`
	WorkspaceRoots []string `json:"workspace_roots,omitempty"`
	FilePath       string   `json:"file_path,omitempty"`
	Command        string   `json:"command,omitempty"`
}

// CursorAdapter 处理路径 /hook/cursor/<event>
type CursorAdapter struct {
	agg *aggregator.Aggregator
}

func NewCursorAdapter(agg *aggregator.Aggregator) *CursorAdapter {
	return &CursorAdapter{agg: agg}
}

// HandleHook 接受 Cursor hook 推送, 默认放行 (返回空对象足以让 Cursor 继续)
func (c *CursorAdapter) HandleHook(event string, payload []byte) ([]byte, error) {
	var in cursorHookInput
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &in)
	}
	if in.HookEventName == "" {
		in.HookEventName = event
	}

	state := mapCursorEventToState(in.HookEventName)
	msg := ""
	if len(in.WorkspaceRoots) > 0 {
		msg = in.WorkspaceRoots[0]
	}
	c.agg.Push(backend.AgentStatus{
		Name:       backend.AgentCursor,
		State:      state,
		Enabled:    true,
		Message:    msg,
		LastUpdate: time.Now().UnixMilli(),
	})
	return []byte(`{}`), nil
}

// mapCursorEventToState Cursor 没有等待用户的事件, 仅在 running 与 idle 间切换
func mapCursorEventToState(event string) backend.AggregatedState {
	switch event {
	case "stop", "sessionEnd":
		return backend.StateIdle
	case "sessionStart":
		return backend.StateIdle
	case "beforeSubmitPrompt",
		"beforeShellExecution", "afterShellExecution",
		"beforeMCPExecution", "afterMCPExecution",
		"beforeReadFile", "afterFileEdit",
		"preToolUse", "postToolUse",
		"afterAgentResponse", "afterAgentThought",
		"beforeTabFileRead", "afterTabFileEdit":
		return backend.StateRunning
	default:
		return backend.StateRunning
	}
}
