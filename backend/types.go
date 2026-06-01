// Package backend 定义后端共享的数据类型与常量
package backend

// AggregatedState 单一 agent 或主灯聚合后的状态; 与前端 src/types.ts 一一对齐
type AggregatedState string

const (
	StateIdle    AggregatedState = "idle"
	StateRunning AggregatedState = "running"
	StateWaiting AggregatedState = "waiting"
	StateError   AggregatedState = "error"
	StateFault   AggregatedState = "fault"
	StateGray    AggregatedState = "gray"
)

// StatePriority 数值越大优先级越高, 用于多 agent 聚合
var StatePriority = map[AggregatedState]int{
	StateError:   5,
	StateFault:   4,
	StateWaiting: 3,
	StateRunning: 2,
	StateIdle:    1,
	StateGray:    0,
}

// AgentName 当前已知的 agent 标识符
type AgentName string

const (
	AgentClaude AgentName = "claude"
	AgentCursor AgentName = "cursor"
	AgentCodex  AgentName = "codex"
)

// AgentStatus 单 agent 的最新状态快照
type AgentStatus struct {
	Name       AgentName       `json:"name"`
	State      AggregatedState `json:"state"`
	Enabled    bool            `json:"enabled"`
	Message    string          `json:"message,omitempty"`
	LastUpdate int64           `json:"lastUpdate"`
}

// Snapshot 聚合后的全量状态; status service 通过此结构对前端暴露
type Snapshot struct {
	Agents    map[AgentName]AgentStatus `json:"agents"`
	MainState AggregatedState           `json:"mainState"`
	UpdatedAt int64                     `json:"updatedAt"`
}

// EventStatusUpdate 前后端事件名: aggregator 每次状态变化时通过 Wails event 推送
const EventStatusUpdate = "status:update"
