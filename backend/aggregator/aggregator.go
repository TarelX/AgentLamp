// Package aggregator 维护各 agent 的最新状态并按优先级合成主灯状态.
// 通过 Wails event 把变化推给前端, 同时支持 status service 拉取快照.
package aggregator

import (
	"sync"
	"time"

	"github.com/TarelX/AgentLamp/backend"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// Aggregator 状态聚合器, 单实例由 main.go 持有并注入到各 adapter
type Aggregator struct {
	mu          sync.RWMutex
	agents      map[backend.AgentName]backend.AgentStatus
	main        backend.AggregatedState
	updates     chan backend.AgentStatus
	app         *application.App
	stop        chan struct{}
	subscribers []func(backend.Snapshot)
}

// Subscribe 注册一个回调, 主灯状态变化时会被调用 (在 emit 同一线程, 不要长阻塞)
func (a *Aggregator) Subscribe(fn func(backend.Snapshot)) {
	a.mu.Lock()
	a.subscribers = append(a.subscribers, fn)
	a.mu.Unlock()
}

// SetApp 在 app 创建后回填引用, 用于后续 emit event
func (a *Aggregator) SetApp(app *application.App) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.app = app
}

// New 创建一个聚合器并把 enabledAgents 列表初始化为 idle, 其余为 gray
func New(app *application.App, enabledAgents []backend.AgentName) *Aggregator {
	now := time.Now().UnixMilli()
	agents := map[backend.AgentName]backend.AgentStatus{}
	enabled := map[backend.AgentName]bool{}
	for _, n := range enabledAgents {
		enabled[n] = true
	}
	for _, n := range []backend.AgentName{backend.AgentClaude, backend.AgentCursor, backend.AgentCodex} {
		agents[n] = backend.AgentStatus{
			Name:       n,
			State:      stateForInit(enabled[n]),
			Enabled:    enabled[n],
			LastUpdate: now,
		}
	}
	a := &Aggregator{
		agents:  agents,
		updates: make(chan backend.AgentStatus, 64),
		app:     app,
		stop:    make(chan struct{}),
	}
	a.main = a.computeMainLocked()
	return a
}

func stateForInit(enabled bool) backend.AggregatedState {
	if enabled {
		return backend.StateIdle
	}
	return backend.StateGray
}

// Run 启动事件循环, 阻塞直到 Stop 调用
func (a *Aggregator) Run() {
	for {
		select {
		case <-a.stop:
			return
		case s := <-a.updates:
			a.apply(s)
		}
	}
}

// Stop 触发 Run 退出
func (a *Aggregator) Stop() {
	close(a.stop)
}

// Push 提交一次状态更新 (非阻塞, 队列满时丢弃以避免锁住 adapter)
func (a *Aggregator) Push(s backend.AgentStatus) {
	if s.LastUpdate == 0 {
		s.LastUpdate = time.Now().UnixMilli()
	}
	select {
	case a.updates <- s:
	default:
	}
}

// SetEnabled 由 status service / 前端调用, 切换 agent 启用状态
func (a *Aggregator) SetEnabled(name backend.AgentName, enabled bool) {
	a.mu.Lock()
	cur, ok := a.agents[name]
	if !ok {
		a.mu.Unlock()
		return
	}
	cur.Enabled = enabled
	if !enabled {
		cur.State = backend.StateGray
	} else if cur.State == backend.StateGray {
		cur.State = backend.StateIdle
	}
	cur.LastUpdate = time.Now().UnixMilli()
	a.agents[name] = cur
	a.main = a.computeMainLocked()
	snap := a.snapshotLocked()
	a.mu.Unlock()
	a.emit(snap)
}

// Snapshot 返回当前完整状态副本, 安全用于跨 goroutine 读取
func (a *Aggregator) Snapshot() backend.Snapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.snapshotLocked()
}

func (a *Aggregator) apply(s backend.AgentStatus) {
	a.mu.Lock()
	cur, ok := a.agents[s.Name]
	if !ok {
		// 未知 agent: 接受并标记为 enabled, 与未来扩展兼容
		cur = backend.AgentStatus{Name: s.Name, Enabled: true}
	}
	if !cur.Enabled {
		// agent 被前端关闭则忽略外部推送
		a.mu.Unlock()
		return
	}
	cur.State = s.State
	cur.Message = s.Message
	cur.LastUpdate = s.LastUpdate
	a.agents[s.Name] = cur
	prevMain := a.main
	a.main = a.computeMainLocked()
	mainChanged := prevMain != a.main
	snap := a.snapshotLocked()
	a.mu.Unlock()
	if mainChanged {
		a.emit(snap)
	}
}

func (a *Aggregator) computeMainLocked() backend.AggregatedState {
	bestPriority := -1
	best := backend.StateGray
	for _, ag := range a.agents {
		if !ag.Enabled {
			continue
		}
		if p := backend.StatePriority[ag.State]; p > bestPriority {
			bestPriority = p
			best = ag.State
		}
	}
	return best
}

func (a *Aggregator) snapshotLocked() backend.Snapshot {
	cp := make(map[backend.AgentName]backend.AgentStatus, len(a.agents))
	for k, v := range a.agents {
		cp[k] = v
	}
	return backend.Snapshot{
		Agents:    cp,
		MainState: a.main,
		UpdatedAt: time.Now().UnixMilli(),
	}
}

func (a *Aggregator) emit(snap backend.Snapshot) {
	a.mu.RLock()
	subs := make([]func(backend.Snapshot), len(a.subscribers))
	copy(subs, a.subscribers)
	a.mu.RUnlock()
	for _, fn := range subs {
		fn(snap)
	}
	if a.app != nil {
		a.app.Event.Emit(backend.EventStatusUpdate, snap)
	}
}
