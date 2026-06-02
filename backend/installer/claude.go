// Package installer 负责把 AgentLamp 的 hook 写入各 agent 的配置文件,
// 必须先备份原文件再修改, 卸载时按 marker 精确回滚.
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

// AgentLampMarker 嵌在 hook 命令中, 用于识别哪些 hook 是本应用安装的
const AgentLampMarker = "AGENTLAMP_V1"

// ClaudeInstaller 操作 ~/.claude/settings.json
type ClaudeInstaller struct {
	settingsPath string
	webhookBase  string
	relayScript  string
}

// NewClaudeInstaller webhookBase 例如 "http://127.0.0.1:19840"
func NewClaudeInstaller(webhookBase string) (*ClaudeInstaller, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}
	relay, err := EnsureRelayScript(webhookBase)
	if err != nil {
		return nil, fmt.Errorf("write relay script: %w", err)
	}
	return &ClaudeInstaller{
		settingsPath: filepath.Join(home, ".claude", "settings.json"),
		webhookBase:  strings.TrimRight(webhookBase, "/"),
		relayScript:  relay,
	}, nil
}

// SettingsPath 暴露给前端做诊断展示
func (c *ClaudeInstaller) SettingsPath() string {
	return c.settingsPath
}

// claudeEvents 监听的 hook 事件名; SubagentStop / PreCompact 信号较弱不挂
var claudeEvents = []string{
	"SessionStart",
	"UserPromptSubmit",
	"PreToolUse",
	"PostToolUse",
	"Notification",
	"Stop",
	"SessionEnd",
}

type hookCmd struct {
	Type    string `json:"type"`
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

type hookGroup struct {
	Matcher string    `json:"matcher"`
	Hooks   []hookCmd `json:"hooks"`
}

// IsInstalled 检查 settings.json 是否已包含本应用的 hooks
func (c *ClaudeInstaller) IsInstalled() (bool, error) {
	raw, err := c.readRaw()
	if err != nil {
		return false, err
	}
	hooksRaw, ok := raw["hooks"]
	if !ok {
		return false, nil
	}
	body := string(hooksRaw)
	return strings.Contains(body, RelayScriptName()) || strings.Contains(body, AgentLampMarker), nil
}

// Install 安装/更新 hooks; 第一次会备份原 settings.json
func (c *ClaudeInstaller) Install() error {
	if err := c.ensureDir(); err != nil {
		return err
	}
	if err := c.backupOnce(); err != nil {
		return err
	}

	raw, err := c.readRaw()
	if err != nil {
		return err
	}

	existing := map[string][]hookGroup{}
	if hooksRaw, ok := raw["hooks"]; ok && len(hooksRaw) > 0 {
		_ = json.Unmarshal(hooksRaw, &existing)
	}

	for _, event := range claudeEvents {
		existing[event] = removeAgentLamp(existing[event])
		existing[event] = append(existing[event], hookGroup{
			Matcher: "*",
			Hooks: []hookCmd{
				{
					Type:    "command",
					Command: c.buildCommand(event),
					Timeout: 5,
				},
			},
		})
	}

	newHooks, err := json.Marshal(existing)
	if err != nil {
		return err
	}
	raw["hooks"] = newHooks

	return c.writeRaw(raw)
}

// Uninstall 移除 hooks 中所有带 AgentLampMarker 的条目, 不影响用户原有 hooks
func (c *ClaudeInstaller) Uninstall() error {
	raw, err := c.readRaw()
	if err != nil {
		return err
	}
	hooksRaw, ok := raw["hooks"]
	if !ok {
		return nil
	}

	existing := map[string][]hookGroup{}
	if err := json.Unmarshal(hooksRaw, &existing); err != nil {
		return err
	}

	for k := range existing {
		filtered := removeAgentLamp(existing[k])
		if len(filtered) == 0 {
			delete(existing, k)
		} else {
			existing[k] = filtered
		}
	}

	if len(existing) == 0 {
		delete(raw, "hooks")
	} else {
		newHooks, err := json.Marshal(existing)
		if err != nil {
			return err
		}
		raw["hooks"] = newHooks
	}

	return c.writeRaw(raw)
}

func (c *ClaudeInstaller) buildCommand(event string) string {
	return HookCommand(c.relayScript, "claude", event)
}

func removeAgentLamp(groups []hookGroup) []hookGroup {
	out := make([]hookGroup, 0, len(groups))
	for _, g := range groups {
		filteredHooks := make([]hookCmd, 0, len(g.Hooks))
		for _, h := range g.Hooks {
			if !IsAgentLampCommand(h.Command) {
				filteredHooks = append(filteredHooks, h)
			}
		}
		if len(filteredHooks) > 0 {
			g.Hooks = filteredHooks
			out = append(out, g)
		}
	}
	return out
}

func (c *ClaudeInstaller) ensureDir() error {
	dir := filepath.Dir(c.settingsPath)
	return os.MkdirAll(dir, 0o755)
}

// backupOnce 仅在还没备份过时复制一份 .agentlamp.bak; 反复 install 不会覆盖原始备份
func (c *ClaudeInstaller) backupOnce() error {
	if _, err := os.Stat(c.settingsPath); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	backup := c.settingsPath + ".agentlamp.bak"
	if _, err := os.Stat(backup); err == nil {
		return nil
	}
	data, err := os.ReadFile(c.settingsPath)
	if err != nil {
		return err
	}
	return os.WriteFile(backup, data, 0o644)
}

func (c *ClaudeInstaller) readRaw() (map[string]json.RawMessage, error) {
	out := map[string]json.RawMessage{}
	data, err := os.ReadFile(c.settingsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return out, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return out, nil
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parse %s: %w", c.settingsPath, err)
	}
	return out, nil
}

func (c *ClaudeInstaller) writeRaw(raw map[string]json.RawMessage) error {
	// 排序 key 让结果稳定可读
	enc, err := marshalSorted(raw)
	if err != nil {
		return err
	}
	tmp := c.settingsPath + ".tmp"
	if err := os.WriteFile(tmp, enc, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, c.settingsPath)
}

// InstallStatus 返回值给前端展示用
type InstallStatus struct {
	Installed    bool   `json:"installed"`
	SettingsPath string `json:"settingsPath"`
	BackupPath   string `json:"backupPath,omitempty"`
	CheckedAt    int64  `json:"checkedAt"`
}

// Status 综合检查
func (c *ClaudeInstaller) Status() (InstallStatus, error) {
	installed, err := c.IsInstalled()
	if err != nil {
		return InstallStatus{}, err
	}
	st := InstallStatus{
		Installed:    installed,
		SettingsPath: c.settingsPath,
		CheckedAt:    time.Now().UnixMilli(),
	}
	backup := c.settingsPath + ".agentlamp.bak"
	if _, err := os.Stat(backup); err == nil {
		st.BackupPath = backup
	}
	return st, nil
}
