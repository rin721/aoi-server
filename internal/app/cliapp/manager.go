package cliapp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/types/constants"
)

const defaultStopTimeout = 30 * time.Second
const managedExecutableBaseName = "go-scaffold-managed"

// Manager 管理 CLI 托管的后台服务进程。
type Manager struct {
	RuntimeDir string
	Executable string
	WorkDir    string
	Runner     ProcessRunner
	Now        func() time.Time
}

// NewManager 创建默认服务管理器。
func NewManager() *Manager {
	executable, _ := os.Executable()
	workDir, _ := os.Getwd()
	runtimeDir := strings.TrimSpace(os.Getenv(RuntimeDirEnvName))
	if runtimeDir == "" {
		runtimeDir = filepath.Join("data", "cli-runtime")
	}
	return &Manager{
		RuntimeDir: runtimeDir,
		Executable: executable,
		WorkDir:    workDir,
		Runner:     newOSProcessRunner(),
		Now:        time.Now,
	}
}

// StartServer 后台启动 server 服务。
func (m *Manager) StartServer(ctx context.Context, configPath string) (ServiceState, error) {
	if err := ctx.Err(); err != nil {
		return ServiceState{}, err
	}
	if configPath == "" {
		configPath = constants.AppDefaultConfigPath
	}
	configPath = filepath.Clean(configPath)

	current, err := m.Status(ctx, ServiceServer)
	if err != nil {
		return current, err
	}
	if current.Status == StatusRunning || current.Status == StatusStarting || current.Status == StatusRestarting {
		return current, fmt.Errorf("%s service is already %s", ServiceServer, current.Status)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		state := m.baseState(ServiceServer, configPath, nil)
		state.Status = StatusFailed
		state.LastError = err.Error()
		_ = m.writeState(state)
		return state, err
	}

	runtimeDir, err := filepath.Abs(m.runtimeDir())
	if err != nil {
		runtimeDir = m.runtimeDir()
	}
	state := m.baseState(ServiceServer, configPath, cfg)
	executablePath, err := m.managedExecutable(runtimeDir)
	if err != nil {
		failedAt := m.now()
		state.Status = StatusFailed
		state.LastError = err.Error()
		state.StoppedAt = &failedAt
		_ = m.writeState(state)
		return state, err
	}
	state.ExecutablePath = executablePath
	startedAt := m.now()
	state.Status = StatusStarting
	state.StartedAt = &startedAt
	state.StoppedAt = nil
	state.LastError = ""
	if err := m.writeState(state); err != nil {
		return state, err
	}
	_ = os.Remove(m.controlPath(ServiceServer))

	info, err := m.runner().StartProcess(ProcessStartRequest{
		Executable: executablePath,
		Args:       []string{constants.AppServerCommandName, "--config", configPath},
		WorkDir:    m.workDir(),
		Env: []string{
			ManagedServiceEnvName + "=1",
			ManagedServiceNameEnvKey + "=" + ServiceServer,
			RuntimeDirEnvName + "=" + runtimeDir,
		},
		StdoutPath: state.StdoutLogPath,
		StderrPath: state.StderrLogPath,
	})
	if err != nil {
		failedAt := m.now()
		state.Status = StatusFailed
		state.LastError = err.Error()
		state.StoppedAt = &failedAt
		_ = m.writeState(state)
		return state, err
	}

	state.PID = info.PID
	state.ProcessStartTime = info.ProcessStartTime
	state.Status = StatusRunning
	if alive, _ := m.runner().IsProcessRunning(info); !alive {
		stoppedAt := m.now()
		state.Status = StatusFailed
		state.LastError = "process exited during startup"
		state.StoppedAt = &stoppedAt
	}
	if err := m.writeState(state); err != nil {
		return state, err
	}
	return state, nil
}

// Status 返回服务状态，并刷新已退出的后台进程。
func (m *Manager) Status(ctx context.Context, service string) (ServiceState, error) {
	if err := ctx.Err(); err != nil {
		return ServiceState{}, err
	}
	service = normalizeServiceName(service)
	state, err := m.readState(service)
	if err != nil {
		return ServiceState{}, err
	}
	if state.Service == "" {
		state = ServiceState{Service: service, Status: StatusStopped}
	}
	if state.PID <= 0 || !activeStatus(state.Status) {
		return state, nil
	}
	running, err := m.runner().IsProcessRunning(ProcessInfo{PID: state.PID, ProcessStartTime: state.ProcessStartTime})
	if err != nil {
		return state, err
	}
	if running {
		if state.Status == StatusStarting || state.Status == StatusRestarting {
			state.Status = StatusRunning
			_ = m.writeState(state)
		}
		return state, nil
	}
	stoppedAt := m.now()
	if state.Status != StatusStopped {
		state.Status = StatusFailed
		state.LastError = "process is not running"
	}
	state.StoppedAt = &stoppedAt
	state.PID = 0
	state.ProcessStartTime = 0
	_ = m.writeState(state)
	return state, nil
}

// StopServer 优雅停止 server 服务，超时后强制结束。
func (m *Manager) StopServer(ctx context.Context) (ServiceState, error) {
	state, err := m.Status(ctx, ServiceServer)
	if err != nil {
		return state, err
	}
	if state.Status != StatusRunning && state.Status != StatusStarting && state.Status != StatusRestarting {
		state.Status = StatusStopped
		_ = m.writeState(state)
		return state, nil
	}

	info := ProcessInfo{PID: state.PID, ProcessStartTime: state.ProcessStartTime}
	if err := m.writeControl(ControlRequest{
		Service:          ServiceServer,
		Action:           controlActionStop,
		PID:              state.PID,
		ProcessStartTime: state.ProcessStartTime,
		RequestedAt:      m.now(),
	}); err != nil {
		return state, err
	}

	deadline := m.now().Add(defaultStopTimeout)
	for m.now().Before(deadline) {
		running, err := m.runner().IsProcessRunning(info)
		if err != nil {
			return state, err
		}
		if !running {
			stoppedAt := m.now()
			state.Status = StatusStopped
			state.PID = 0
			state.ProcessStartTime = 0
			state.StoppedAt = &stoppedAt
			state.LastError = ""
			_ = m.writeState(state)
			_ = os.Remove(m.controlPath(ServiceServer))
			return state, nil
		}
		select {
		case <-ctx.Done():
			return state, ctx.Err()
		case <-time.After(300 * time.Millisecond):
		}
	}

	if err := m.runner().KillProcess(info); err != nil {
		return state, err
	}
	stoppedAt := m.now()
	state.Status = StatusStopped
	state.PID = 0
	state.ProcessStartTime = 0
	state.StoppedAt = &stoppedAt
	state.LastError = "forced stop after graceful timeout"
	_ = m.writeState(state)
	_ = os.Remove(m.controlPath(ServiceServer))
	return state, nil
}

// RestartServer 重启 server 服务，沿用上次配置路径。
func (m *Manager) RestartServer(ctx context.Context) (ServiceState, error) {
	state, err := m.Status(ctx, ServiceServer)
	if err != nil {
		return state, err
	}
	configPath := state.ConfigPath
	if configPath == "" {
		configPath = constants.AppDefaultConfigPath
	}
	state.Status = StatusRestarting
	_ = m.writeState(state)
	if _, err := m.StopServer(ctx); err != nil {
		return state, err
	}
	return m.StartServer(ctx, configPath)
}

// MarkManagedServiceStopped 供托管 server 进程在优雅退出后更新状态文件。
func MarkManagedServiceStopped(service string, lastError string) {
	if os.Getenv(ManagedServiceEnvName) == "" {
		return
	}
	manager := NewManager()
	state, err := manager.readState(normalizeServiceName(service))
	if err != nil {
		return
	}
	stoppedAt := manager.now()
	state.Status = StatusStopped
	if strings.TrimSpace(lastError) != "" {
		state.Status = StatusFailed
		state.LastError = lastError
	}
	state.PID = 0
	state.ProcessStartTime = 0
	state.StoppedAt = &stoppedAt
	_ = manager.writeState(state)
}

func (m *Manager) baseState(service string, configPath string, cfg *appconfig.Config) ServiceState {
	service = normalizeServiceName(service)
	serviceDir := m.serviceDir(service)
	state := ServiceState{
		Service:       service,
		Status:        StatusStopped,
		ConfigPath:    configPath,
		StdoutLogPath: filepath.Join(serviceDir, "stdout.log"),
		StderrLogPath: filepath.Join(serviceDir, "stderr.log"),
	}
	if cfg != nil {
		state.ListenAddr = net.JoinHostPort(cfg.Server.Host, fmt.Sprint(cfg.Server.Port))
		state.AppLogPath = cfg.Logger.FilePath
	}
	return state
}

func (m *Manager) readState(service string) (ServiceState, error) {
	service = normalizeServiceName(service)
	raw, err := os.ReadFile(m.statePath(service))
	if errors.Is(err, os.ErrNotExist) {
		return ServiceState{Service: service, Status: StatusStopped}, nil
	}
	if err != nil {
		return ServiceState{}, err
	}
	var state ServiceState
	if err := json.Unmarshal(raw, &state); err != nil {
		return ServiceState{}, err
	}
	if state.Service == "" {
		state.Service = service
	}
	if state.Status == "" {
		state.Status = StatusStopped
	}
	return state, nil
}

func (m *Manager) writeState(state ServiceState) error {
	if state.Service == "" {
		state.Service = ServiceServer
	}
	if state.Status == "" {
		state.Status = StatusStopped
	}
	path := m.statePath(state.Service)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	_ = os.Remove(path)
	return os.Rename(tmp, path)
}

func (m *Manager) writeControl(req ControlRequest) error {
	path := m.controlPath(req.Service)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func (m *Manager) statePath(service string) string {
	return filepath.Join(m.serviceDir(service), "state.json")
}

func (m *Manager) controlPath(service string) string {
	return filepath.Join(m.serviceDir(service), "control.json")
}

func (m *Manager) serviceDir(service string) string {
	return filepath.Join(m.runtimeDir(), normalizeServiceName(service))
}

func (m *Manager) runtimeDir() string {
	if strings.TrimSpace(m.RuntimeDir) != "" {
		return filepath.Clean(m.RuntimeDir)
	}
	return filepath.Join("data", "cli-runtime")
}

func (m *Manager) executable() string {
	if strings.TrimSpace(m.Executable) != "" {
		return m.Executable
	}
	executable, _ := os.Executable()
	return executable
}

func (m *Manager) managedExecutable(runtimeDir string) (string, error) {
	executable := strings.TrimSpace(m.executable())
	if executable == "" {
		return "", errors.New("executable is required")
	}
	executable = filepath.Clean(executable)
	if !isGoRunTemporaryExecutable(executable) {
		return executable, nil
	}
	target := filepath.Join(runtimeDir, "bin", managedExecutableFileName(executable))
	if err := copyExecutable(executable, target); err != nil {
		return "", fmt.Errorf("prepare managed executable: copy %s to %s: %w; stop the existing managed service or build a stable binary with go build before running it in the background", executable, target, err)
	}
	return target, nil
}

func managedExecutableFileName(source string) string {
	if strings.EqualFold(filepath.Ext(source), ".exe") {
		return managedExecutableBaseName + ".exe"
	}
	return managedExecutableBaseName
}

func isGoRunTemporaryExecutable(path string) bool {
	path = filepath.Clean(path)
	base := strings.ToLower(filepath.Base(path))
	if base != "main" && base != "main.exe" {
		return false
	}
	if strings.ToLower(filepath.Base(filepath.Dir(path))) != "exe" {
		return false
	}
	dir := filepath.Dir(filepath.Dir(path))
	for {
		name := strings.ToLower(filepath.Base(dir))
		if strings.HasPrefix(name, "go-build") {
			return true
		}
		next := filepath.Dir(dir)
		if next == dir {
			return false
		}
		dir = next
	}
}

func copyExecutable(source string, target string) error {
	sourceAbs, sourceErr := filepath.Abs(source)
	targetAbs, targetErr := filepath.Abs(target)
	if sourceErr == nil && targetErr == nil && sourceAbs == targetAbs {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := target + ".tmp"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	if err := os.Chmod(tmp, 0o755); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	_ = os.Remove(target)
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (m *Manager) workDir() string {
	if strings.TrimSpace(m.WorkDir) != "" {
		return m.WorkDir
	}
	workDir, _ := os.Getwd()
	return workDir
}

func (m *Manager) runner() ProcessRunner {
	if m.Runner != nil {
		return m.Runner
	}
	return newOSProcessRunner()
}

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now().UTC()
	}
	return time.Now().UTC()
}

func normalizeServiceName(service string) string {
	service = strings.ToLower(strings.TrimSpace(service))
	if service == "" {
		return ServiceServer
	}
	return service
}

func activeStatus(status string) bool {
	switch status {
	case StatusStarting, StatusRunning, StatusRestarting:
		return true
	default:
		return false
	}
}

func loadConfig(configPath string) (*appconfig.Config, error) {
	manager := appconfig.NewManager()
	if err := manager.Load(configPath); err != nil {
		return nil, err
	}
	return manager.Get(), nil
}
