package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// RelayScriptName Windows 用 .ps1 (spawn 不能直接跑 .cmd, 必须 PowerShell)
func RelayScriptName() string {
	if runtime.GOOS == "windows" {
		return "hook-relay.ps1"
	}
	return "hook-relay.sh"
}

// AgentLampConfigDir 用户配置目录, 存放 relay 脚本与日志
func AgentLampConfigDir() (string, error) {
	if runtime.GOOS == "windows" {
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appdata = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appdata, "AgentLamp"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agentlamp"), nil
}

// EnsureRelayScript 写出/更新 relay 脚本; 返回脚本绝对路径
func EnsureRelayScript(webhookBase string) (string, error) {
	dir, err := AgentLampConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	target := filepath.Join(dir, RelayScriptName())
	cleanupOldScripts(dir)
	content := relayScriptContent(webhookBase)
	data := []byte(content)
	if runtime.GOOS == "windows" {
		// Windows PowerShell 5.x 在中文系统默认按 ANSI 读 .ps1, 加 UTF-8 BOM 强制 UTF-8 解析
		data = append([]byte{0xEF, 0xBB, 0xBF}, data...)
	}
	if err := os.WriteFile(target, data, 0o755); err != nil {
		return "", err
	}
	return target, nil
}

// cleanupOldScripts 删掉早期版本遗留的 hook-relay.cmd, 避免 Cursor 仍解析旧路径
func cleanupOldScripts(dir string) {
	if runtime.GOOS == "windows" {
		_ = os.Remove(filepath.Join(dir, "hook-relay.cmd"))
	}
}

func relayScriptContent(webhookBase string) string {
	if runtime.GOOS == "windows" {
		// ValueFromRemainingArguments 吞掉残留 marker / 多余位置参数, 防 IDE 升级老 hooks.json 报错
		return fmt.Sprintf(`# AgentLamp hook relay (Windows / PowerShell)
# 由 Cursor / Claude IDE 通过 hooks 配置调用
# 用法: powershell -ExecutionPolicy Bypass -File hook-relay.ps1 <agent> <event>
param(
    [Parameter(Position=0)] [string] $Agent,
    [Parameter(Position=1)] [string] $Event,
    [Parameter(ValueFromRemainingArguments=$true)] $Rest
)

$WebhookBase = '%s'
$Url = "$WebhookBase/hook/$Agent/$Event"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$LogFile = Join-Path $ScriptDir 'hook.log'

try {
    $body = [Console]::In.ReadToEnd()
    "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] $Agent $Event" | Out-File -FilePath $LogFile -Encoding UTF8 -Append
    if ([string]::IsNullOrEmpty($body)) {
        Invoke-RestMethod -Uri $Url -Method Post -ContentType 'application/json' -TimeoutSec 2 -ErrorAction Stop | Out-Null
    } else {
        Invoke-RestMethod -Uri $Url -Method Post -Body $body -ContentType 'application/json' -TimeoutSec 2 -ErrorAction Stop | Out-Null
    }
} catch {
    "[$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')] $Agent $Event ERROR: $_" | Out-File -FilePath $LogFile -Encoding UTF8 -Append
}

Write-Output '{}'
exit 0
`, webhookBase)
	}

	return fmt.Sprintf(`#!/bin/sh
# AgentLamp hook relay (POSIX)
AGENT="$1"
EVT="$2"
DIR=$(dirname "$0")
LOG="$DIR/hook.log"
echo "[$(date '+%%F %%T')] $AGENT $EVT" >>"$LOG"
exec curl -s -m 2 -X POST -H "Content-Type: application/json" --data-binary @- "%s/hook/$AGENT/$EVT" >>"$LOG" 2>&1
`, webhookBase)
}

// HookCommand 拼出可直接写入 hooks.json 的命令字符串.
// 不附带任何 # 注释 marker, 避免 Cursor 在 Windows 用 PowerShell scriptblock
// 包裹命令时把 # 当注释起始符吞掉后续大括号导致解析失败.
// 我们装的 hook 通过命令里包含 relay 脚本路径 (天然唯一) 来识别.
func HookCommand(relayPath, agent, event string) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf(`powershell -ExecutionPolicy Bypass -File "%s" %s %s`, relayPath, agent, event)
	}
	return fmt.Sprintf(`"%s" %s %s`, relayPath, agent, event)
}

// IsAgentLampCommand 判断一条 hook 命令是不是本应用安装的, 既识别新路径也兼容老 marker
func IsAgentLampCommand(cmd string) bool {
	return strings.Contains(cmd, RelayScriptName()) ||
		strings.Contains(cmd, AgentLampMarker)
}
