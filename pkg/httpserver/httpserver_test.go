package httpserver

// 本测试文件固定 HTTP 服务器的配置校验、启动和重载边界，防止注释补全和后续重构改变外部可观察行为。

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/rei0721/go-scaffold/pkg/logger"
)

// TestNewAppliesDefaultsAndRejectsInvalidInput 固定 HTTP 服务器的配置校验、启动和重载边界，确保后续注释补全或结构调整不改变该场景。
func TestNewAppliesDefaultsAndRejectsInvalidInput(t *testing.T) {
	if _, err := New(nil, &Config{}, noopLogger{}); err == nil {
		t.Fatal("New(nil handler) error = nil, want error")
	}

	srv, err := New(http.NewServeMux(), &Config{}, noopLogger{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	got := srv.(*httpServer)
	if got.config.Host != DefaultHost {
		t.Fatalf("Host = %q, want %q", got.config.Host, DefaultHost)
	}
	if got.config.Port != DefaultPort {
		t.Fatalf("Port = %d, want %d", got.config.Port, DefaultPort)
	}
	if got.config.ReadTimeout != DefaultReadTimeout {
		t.Fatalf("ReadTimeout = %s, want %s", got.config.ReadTimeout, DefaultReadTimeout)
	}

	_, err = New(http.NewServeMux(), &Config{Port: -1}, noopLogger{})
	var serverErr *ServerError
	if !errors.As(err, &serverErr) {
		t.Fatalf("New(invalid config) error = %T, want *ServerError", err)
	}
	var configErr *ConfigError
	if !errors.As(err, &configErr) {
		t.Fatalf("New(invalid config) unwrap = %v, want *ConfigError", err)
	}
}

// TestReloadUpdatesConfigWhenStopped 固定 HTTP 服务器的配置校验、启动和重载边界，确保后续注释补全或结构调整不改变该场景。
func TestReloadUpdatesConfigWhenStopped(t *testing.T) {
	srv, err := New(http.NewServeMux(), &Config{Host: "127.0.0.1", Port: 18080}, noopLogger{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	next := &Config{
		Host:         "0.0.0.0",
		Port:         18081,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
	if err := srv.Reload(context.Background(), next); err != nil {
		t.Fatalf("Reload(stopped) error = %v", err)
	}
	got := srv.(*httpServer)
	if got.config.Host != "0.0.0.0" || got.config.Port != 18081 {
		t.Fatalf("config = %s:%d, want 0.0.0.0:18081", got.config.Host, got.config.Port)
	}
	if got.config.IdleTimeout != DefaultIdleTimeout {
		t.Fatalf("IdleTimeout = %s, want %s", got.config.IdleTimeout, DefaultIdleTimeout)
	}

	if err := srv.Reload(context.Background(), nil); err == nil {
		t.Fatal("Reload(nil) error = nil, want error")
	}
	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown(stopped) error = %v", err)
	}
}

// TestStartRejectsAlreadyRunningServer 固定 HTTP 服务器的配置校验、启动和重载边界，确保后续注释补全或结构调整不改变该场景。
func TestStartRejectsAlreadyRunningServer(t *testing.T) {
	srv, err := New(http.NewServeMux(), &Config{}, noopLogger{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	got := srv.(*httpServer)
	got.state.Store(int32(stateRunning))
	if err := srv.Start(context.Background()); err == nil {
		t.Fatal("Start(already running) error = nil, want error")
	}
}

// TestStartReturnsBindError 固定 HTTP 服务器的配置校验、启动和重载边界，确保后续注释补全或结构调整不改变该场景。
func TestStartReturnsBindError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen occupied port: %v", err)
	}
	defer listener.Close()

	_, portValue, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split listener addr: %v", err)
	}
	port, err := strconv.Atoi(portValue)
	if err != nil {
		t.Fatalf("parse listener port: %v", err)
	}

	srv, err := New(http.NewServeMux(), &Config{Host: "127.0.0.1", Port: port}, noopLogger{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := srv.Start(context.Background()); err == nil {
		t.Fatal("Start(occupied port) error = nil, want bind error")
	}
}

type noopLogger struct{}

// Debug 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (noopLogger) Debug(string, ...interface{}) {}

// Info 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (noopLogger) Info(string, ...interface{}) {}

// Warn 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (noopLogger) Warn(string, ...interface{}) {}

// Error 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (noopLogger) Error(string, ...interface{}) {}

// Fatal 实现测试日志桩的同名输出入口，当前测试只关心接口满足而不采集日志内容。
func (noopLogger) Fatal(string, ...interface{}) {}

// With 实现测试日志桩的字段绑定入口，返回自身以保持 logger.Logger 链式调用契约。
func (l noopLogger) With(...interface{}) logger.Logger {
	return l
}

// Sync 实现测试日志桩的刷新入口，测试环境不持有真实缓冲区。
func (noopLogger) Sync() error {
	return nil
}

// Reload 实现测试桩的配置重载入口，用于验证调用路径而不触发真实资源替换。
func (noopLogger) Reload(*logger.Config) error {
	return nil
}
