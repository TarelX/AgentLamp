package installer

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CursorInstaller 操作 ~/.cursor/hooks.json (全局 hooks)
type CursorInstaller struct {
	hooksPath   string
	webhookBase string
}

func NewCursorInstaller(webhookBase string) (*CursorInstaller, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	return &CursorInstaller{
		hooksPath:   filepath.Join(home, ".cursor", "hooks.json"),
		webhookBase: strings.TrimRight(webhookBase, "/"),
	}, nil
}

func (c *CursorInstaller) HooksPath() string {
	return c.hooksPath
}

// cursorEvents 仅挂信号最强的事件, 排除 afterAgentThought 等高频事件以减少抖动
var cursorEvents = []string{
	"sessionStart",
	"beforeSubmitPrompt",
	"beforeShellExecution",
	"afterShellExecution",
	"afterFileEdit",
	"postToolUse",
	"stop",
	"sessionEnd",
}

type cursorHookEntry struct {
	Command string `json:"command"`
}

type cursorHooksFile struct {
	Schema  string                       `json:"$schema,omitempty"`
	Version int                          `json:"version"`
	Hooks   map[string][]cursorHookEntry `json:"hooks"`
}

func (c *CursorInstaller) IsInstalled() (bool, error) {
	data, err := os.ReadFile(c.hooksPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.Contains(string(data), AgentLampMarker), nil
}

func (c *CursorInstaller) Install() error {
	if err := os.MkdirAll(filepath.Dir(c.hooksPath), 0o755); err != nil {
		return err
	}
	if err := c.backupOnce(); err != nil {
		return err
	}

	cfg, err := c.read()
	if err != nil {
		return err
	}
	if cfg.Hooks == nil {
		cfg.Hooks = map[string][]cursorHookEntry{}
	}
	cfg.Version = 1

	for _, ev := range cursorEvents {
		cfg.Hooks[ev] = removeAgentLampCursor(cfg.Hooks[ev])
		cfg.Hooks[ev] = append(cfg.Hooks[ev], cursorHookEntry{
			Command: c.buildCommand(ev),
		})
	}

	return c.write(cfg)
}

func (c *CursorInstaller) Uninstall() error {
	cfg, err := c.read()
	if err != nil {
		return err
	}
	if cfg.Hooks == nil {
		return nil
	}
	for k := range cfg.Hooks {
		filtered := removeAgentLampCursor(cfg.Hooks[k])
		if len(filtered) == 0 {
			delete(cfg.Hooks, k)
		} else {
			cfg.Hooks[k] = filtered
		}
	}
	if len(cfg.Hooks) == 0 {
		// 完全恢复成空文件可能误导用户, 留 version+空 hooks 兼容 Cursor
		cfg.Hooks = nil
	}
	return c.write(cfg)
}

func (c *CursorInstaller) Status() (InstallStatus, error) {
	installed, err := c.IsInstalled()
	if err != nil {
		return InstallStatus{}, err
	}
	st := InstallStatus{
		Installed:    installed,
		SettingsPath: c.hooksPath,
		CheckedAt:    time.Now().UnixMilli(),
	}
	backup := c.hooksPath + ".agentlamp.bak"
	if _, err := os.Stat(backup); err == nil {
		st.BackupPath = backup
	}
	return st, nil
}

func (c *CursorInstaller) buildCommand(event string) string {
	url := fmt.Sprintf("%s/hook/cursor/%s", c.webhookBase, event)
	return fmt.Sprintf(
		`"%s" -s -m 2 -X POST -H "Content-Type: application/json" --data-binary @- %s # %s`,
		curlExecutable(), url, AgentLampMarker,
	)
}

func removeAgentLampCursor(entries []cursorHookEntry) []cursorHookEntry {
	out := make([]cursorHookEntry, 0, len(entries))
	for _, e := range entries {
		if !strings.Contains(e.Command, AgentLampMarker) {
			out = append(out, e)
		}
	}
	return out
}

func (c *CursorInstaller) backupOnce() error {
	if _, err := os.Stat(c.hooksPath); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	backup := c.hooksPath + ".agentlamp.bak"
	if _, err := os.Stat(backup); err == nil {
		return nil
	}
	data, err := os.ReadFile(c.hooksPath)
	if err != nil {
		return err
	}
	return os.WriteFile(backup, data, 0o644)
}

func (c *CursorInstaller) read() (*cursorHooksFile, error) {
	cfg := &cursorHooksFile{Version: 1, Hooks: map[string][]cursorHookEntry{}}
	data, err := os.ReadFile(c.hooksPath)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", c.hooksPath, err)
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if cfg.Hooks == nil {
		cfg.Hooks = map[string][]cursorHookEntry{}
	}
	return cfg, nil
}

func (c *CursorInstaller) write(cfg *cursorHooksFile) error {
	enc, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := c.hooksPath + ".tmp"
	if err := os.WriteFile(tmp, enc, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.hooksPath)
}
