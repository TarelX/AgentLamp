package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// RelayScriptName hook IDE 调用的中继脚本名; cmd / sh 两种实现
func RelayScriptName() string {
	if runtime.GOOS == "windows" {
		return "hook-relay.cmd"
	}
	return "hook-relay.sh"
}

// AgentLampConfigDir AgentLamp 的用户配置目录, 存放 relay 脚本与日志
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
	content := relayScriptContent(webhookBase)
	if err := os.WriteFile(target, []byte(content), 0o755); err != nil {
		return "", err
	}
	return target, nil
}

func relayScriptContent(webhookBase string) string {
	if runtime.GOOS == "windows" {
		// %~1 = agent, %~2 = event; stdin 透传给 curl, 失败静默不打断 Cursor / Claude
		return fmt.Sprintf("@echo off\r\n"+
			"set \"AGENT=%%~1\"\r\n"+
			"set \"EVT=%%~2\"\r\n"+
			"echo [%%date%% %%time%%] %%AGENT%% %%EVT%% >>\"%%~dp0hook.log\"\r\n"+
			"\"C:\\Windows\\System32\\curl.exe\" -s -m 2 -X POST -H \"Content-Type: application/json\" --data-binary @- \"%s/hook/%%AGENT%%/%%EVT%%\" 1>>\"%%~dp0hook.log\" 2>&1\r\n"+
			"exit /b 0\r\n", webhookBase)
	}
	return fmt.Sprintf("#!/bin/sh\n"+
		"AGENT=\"$1\"\n"+
		"EVT=\"$2\"\n"+
		"DIR=$(dirname \"$0\")\n"+
		"echo \"[$(date)] $AGENT $EVT\" >>\"$DIR/hook.log\"\n"+
		"exec curl -s -m 2 -X POST -H \"Content-Type: application/json\" --data-binary @- \"%s/hook/$AGENT/$EVT\" >>\"$DIR/hook.log\" 2>&1\n", webhookBase)
}
