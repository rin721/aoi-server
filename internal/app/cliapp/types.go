package cliapp

import "time"

const (
	ServiceServer = "server"

	StatusStarting   = "starting"
	StatusRunning    = "running"
	StatusStopped    = "stopped"
	StatusFailed     = "failed"
	StatusRestarting = "restarting"

	RuntimeDirEnvName        = "RIN_CLI_RUNTIME_DIR"
	ManagedServiceEnvName    = "RIN_CLI_MANAGED"
	ManagedServiceNameEnvKey = "RIN_CLI_SERVICE"

	controlActionStop = "stop"
)

// ServiceState 是 CLI 后台服务管理的持久化状态。
type ServiceState struct {
	Service          string     `json:"service"`
	Status           string     `json:"status"`
	PID              int        `json:"pid"`
	ProcessStartTime int64      `json:"processStartTime"`
	StartedAt        *time.Time `json:"startedAt,omitempty"`
	StoppedAt        *time.Time `json:"stoppedAt,omitempty"`
	ConfigPath       string     `json:"configPath"`
	ListenAddr       string     `json:"listenAddr"`
	StdoutLogPath    string     `json:"stdoutLogPath"`
	StderrLogPath    string     `json:"stderrLogPath"`
	AppLogPath       string     `json:"appLogPath"`
	LastError        string     `json:"lastError,omitempty"`
}

// ControlRequest 是 CLI 写给托管服务进程的控制消息。
type ControlRequest struct {
	Service          string    `json:"service"`
	Action           string    `json:"action"`
	PID              int       `json:"pid"`
	ProcessStartTime int64     `json:"processStartTime"`
	RequestedAt      time.Time `json:"requestedAt"`
}

// ProcessInfo 描述一个已启动或已探测进程。
type ProcessInfo struct {
	PID              int
	ProcessStartTime int64
}

// ProcessStartRequest 描述后台进程启动请求。
type ProcessStartRequest struct {
	Executable string
	Args       []string
	WorkDir    string
	Env        []string
	StdoutPath string
	StderrPath string
}

// ProcessRunner 隔离真实操作系统进程操作，便于测试服务状态机。
type ProcessRunner interface {
	StartProcess(ProcessStartRequest) (ProcessInfo, error)
	IsProcessRunning(ProcessInfo) (bool, error)
	KillProcess(ProcessInfo) error
}
