// Package server 暴露本机 webhook 入口供 agent hook 推送状态.
//
// 路由: POST /hook/<agent>/<event>
// 仅监听 127.0.0.1, 所有请求体作为原始 JSON payload 透传给 adapter.
package server

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/TarelX/AgentLamp/backend"
)

// HookHandler agent 适配器实现此接口接收 hook 推送.
// response 直接作为 HTTP 响应体返回给 agent (例如 Claude 期待 {"continue":true},
// Cursor 期待 {"permission":"allow"} 或空对象).
type HookHandler interface {
	HandleHook(event string, payload []byte) (response []byte, err error)
}

// Server 本机 HTTP webhook
type Server struct {
	addr     string
	handlers map[backend.AgentName]HookHandler
	logger   *slog.Logger
	srv      *http.Server
}

// New 创建 server 实例; addr 形如 "127.0.0.1:19840"
func New(addr string, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		addr:     addr,
		handlers: map[backend.AgentName]HookHandler{},
		logger:   logger,
	}
}

// Register 注册 agent 名到 handler 的映射 (在 Start 之前调用)
func (s *Server) Register(name backend.AgentName, h HookHandler) {
	s.handlers[name] = h
}

// Start 阻塞监听; 仅在 Stop 或致命错误时返回
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/hook/", s.handleHook)
	mux.HandleFunc("/health", s.handleHealth)

	s.srv = &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// 绑定到本机回环避免暴露到局域网
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.logger.Info("webhook listening", "addr", s.addr)
	if err := s.srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop 优雅关闭, 等待 in-flight 请求处理完
func (s *Server) Stop(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}
	return s.srv.Shutdown(ctx)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

// handleHook 解析路径 /hook/<agent>/<event> 并把 payload 透给 adapter
func (s *Server) handleHook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/hook/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		http.Error(w, "expected /hook/<agent>/<event>", http.StatusBadRequest)
		return
	}
	agent := backend.AgentName(parts[0])
	event := parts[1]

	h, ok := s.handlers[agent]
	if !ok {
		http.Error(w, "unknown agent", http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := h.HandleHook(event, body)
	if err != nil {
		s.logger.Warn("hook handler error", "agent", agent, "event", event, "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if len(resp) == 0 {
		_, _ = w.Write([]byte(`{}`))
		return
	}
	_, _ = w.Write(resp)
}
