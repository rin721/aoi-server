package cliapp

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"
)

// WatchManagedServiceControl 监听 CLI 写入的控制文件。
//
// 该函数只在由 CLI 托管的后台进程中生效；前台 server 命令不会额外监听控制文件。
func WatchManagedServiceControl(ctx context.Context, service string) <-chan ControlRequest {
	if os.Getenv(ManagedServiceEnvName) == "" {
		return nil
	}
	out := make(chan ControlRequest, 1)
	service = normalizeServiceName(service)
	manager := NewManager()
	self := currentProcessInfo()
	go func() {
		defer close(out)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				req, ok := readControlRequest(manager.controlPath(service))
				if !ok || !matchesCurrentProcess(req, service, self) {
					continue
				}
				_ = os.Remove(manager.controlPath(service))
				out <- req
				return
			}
		}
	}()
	return out
}

func readControlRequest(path string) (ControlRequest, bool) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) || err != nil {
		return ControlRequest{}, false
	}
	var req ControlRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return ControlRequest{}, false
	}
	return req, true
}

func matchesCurrentProcess(req ControlRequest, service string, self ProcessInfo) bool {
	if normalizeServiceName(req.Service) != service {
		return false
	}
	if req.Action != controlActionStop {
		return false
	}
	if req.PID != self.PID {
		return false
	}
	if req.ProcessStartTime > 0 && self.ProcessStartTime > 0 && req.ProcessStartTime != self.ProcessStartTime {
		return false
	}
	return true
}
