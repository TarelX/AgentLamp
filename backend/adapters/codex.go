package adapters

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/TarelX/AgentLamp/backend"
	"github.com/TarelX/AgentLamp/backend/aggregator"
	_ "modernc.org/sqlite"
)

// CodexWatcher 轮询 ~/.codex/state_5.sqlite 的 threads.updated_at_ms 判断 Codex 状态.
//
// 直接读 SQLite 比 fsnotify 监听 sqlite-wal 准确, 因 wal 在 checkpoint 期间会持续被
// 写入但没有用户活动. updated_at_ms 是 OpenAI 内部记账的最近一次写入时间.
type CodexWatcher struct {
	agg           *aggregator.Aggregator
	dbPath        string
	runningWindow time.Duration
	pollInterval  time.Duration
	logger        *slog.Logger
	stop          chan struct{}
}

// NewCodexWatcher runningWindow 控制 "最近一次写入多久内仍视为 running", 经验值 3 秒
func NewCodexWatcher(agg *aggregator.Aggregator, runningWindow time.Duration, logger *slog.Logger) (*CodexWatcher, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &CodexWatcher{
		agg:           agg,
		dbPath:        filepath.Join(home, ".codex", "state_5.sqlite"),
		runningWindow: runningWindow,
		pollInterval:  time.Second,
		logger:        logger,
		stop:          make(chan struct{}),
	}, nil
}

// DBPath 暴露给前端做诊断
func (w *CodexWatcher) DBPath() string {
	return w.dbPath
}

// Start 异步启动轮询; 数据库不存在则推 gray 静默退出
func (w *CodexWatcher) Start() error {
	if _, err := os.Stat(w.dbPath); errors.Is(err, os.ErrNotExist) {
		w.agg.Push(backend.AgentStatus{
			Name:    backend.AgentCodex,
			State:   backend.StateGray,
			Enabled: true,
		})
		return nil
	}
	// 只读打开避免与 Codex 写入争锁; busy_timeout 让阻塞期间能等
	dsn := fmt.Sprintf("file:%s?mode=ro&_pragma=busy_timeout(2000)", url.PathEscape(w.dbPath))
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return err
	}
	db.SetMaxOpenConns(1)

	go w.loop(db)
	return nil
}

// Stop 触发轮询 goroutine 退出
func (w *CodexWatcher) Stop() {
	select {
	case <-w.stop:
	default:
		close(w.stop)
	}
}

func (w *CodexWatcher) loop(db *sql.DB) {
	defer db.Close()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	var lastState backend.AggregatedState
	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			state, err := w.queryState(db)
			if err != nil {
				w.logger.Warn("codex sqlite query failed", "err", err)
				continue
			}
			if state == lastState {
				continue
			}
			lastState = state
			w.agg.Push(backend.AgentStatus{
				Name:       backend.AgentCodex,
				State:      state,
				Enabled:    true,
				LastUpdate: time.Now().UnixMilli(),
			})
		}
	}
}

func (w *CodexWatcher) queryState(db *sql.DB) (backend.AggregatedState, error) {
	var lastMs int64
	err := db.QueryRow("SELECT COALESCE(MAX(updated_at_ms), 0) FROM threads").Scan(&lastMs)
	if err != nil {
		return backend.StateGray, err
	}
	if lastMs == 0 {
		return backend.StateIdle, nil
	}
	if time.Since(time.UnixMilli(lastMs)) < w.runningWindow {
		return backend.StateRunning, nil
	}
	return backend.StateIdle, nil
}
