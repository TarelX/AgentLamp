package service

import (
	"github.com/TarelX/AgentLamp/backend/installer"
)

// InstallService 暴露给前端的 hook 安装能力
type InstallService struct {
	claude *installer.ClaudeInstaller
	cursor *installer.CursorInstaller
}

// NewInstallService webhookBase 形如 "http://127.0.0.1:19840"
func NewInstallService(webhookBase string) (*InstallService, error) {
	c, err := installer.NewClaudeInstaller(webhookBase)
	if err != nil {
		return nil, err
	}
	cur, err := installer.NewCursorInstaller(webhookBase)
	if err != nil {
		return nil, err
	}
	return &InstallService{claude: c, cursor: cur}, nil
}

func (s *InstallService) ClaudeStatus() (installer.InstallStatus, error) {
	return s.claude.Status()
}

func (s *InstallService) ClaudeInstall() (installer.InstallStatus, error) {
	if err := s.claude.Install(); err != nil {
		return installer.InstallStatus{}, err
	}
	return s.claude.Status()
}

func (s *InstallService) ClaudeUninstall() (installer.InstallStatus, error) {
	if err := s.claude.Uninstall(); err != nil {
		return installer.InstallStatus{}, err
	}
	return s.claude.Status()
}

func (s *InstallService) CursorStatus() (installer.InstallStatus, error) {
	return s.cursor.Status()
}

func (s *InstallService) CursorInstall() (installer.InstallStatus, error) {
	if err := s.cursor.Install(); err != nil {
		return installer.InstallStatus{}, err
	}
	return s.cursor.Status()
}

func (s *InstallService) CursorUninstall() (installer.InstallStatus, error) {
	if err := s.cursor.Uninstall(); err != nil {
		return installer.InstallStatus{}, err
	}
	return s.cursor.Status()
}
