package service

import (
	"github.com/TarelX/AgentLamp/backend/installer"
)

// InstallService 暴露给前端的 hook 安装能力, 当前仅支持 Claude Code
type InstallService struct {
	claude *installer.ClaudeInstaller
}

// NewInstallService webhookBase 形如 "http://127.0.0.1:19840"
func NewInstallService(webhookBase string) (*InstallService, error) {
	c, err := installer.NewClaudeInstaller(webhookBase)
	if err != nil {
		return nil, err
	}
	return &InstallService{claude: c}, nil
}

// ClaudeStatus 查询 Claude hook 当前是否已安装, 不修改任何文件
func (s *InstallService) ClaudeStatus() (installer.InstallStatus, error) {
	return s.claude.Status()
}

// ClaudeInstall 写入 ~/.claude/settings.json 的 hooks; 首次会备份原文件
func (s *InstallService) ClaudeInstall() (installer.InstallStatus, error) {
	if err := s.claude.Install(); err != nil {
		return installer.InstallStatus{}, err
	}
	return s.claude.Status()
}

// ClaudeUninstall 仅移除带 AgentLamp marker 的 hook 条目
func (s *InstallService) ClaudeUninstall() (installer.InstallStatus, error) {
	if err := s.claude.Uninstall(); err != nil {
		return installer.InstallStatus{}, err
	}
	return s.claude.Status()
}
