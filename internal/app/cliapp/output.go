package cliapp

import (
	"fmt"
	"io"
)

// PrintServiceState 输出托管服务状态。
func PrintServiceState(w io.Writer, state ServiceState) {
	fmt.Fprintf(w, "服务：%s\n", state.Service)
	fmt.Fprintf(w, "状态：%s\n", state.Status)
	if state.PID > 0 {
		fmt.Fprintf(w, "PID：%d\n", state.PID)
	}
	if state.ListenAddr != "" {
		fmt.Fprintf(w, "监听：%s\n", state.ListenAddr)
	}
	if state.ExecutablePath != "" {
		fmt.Fprintf(w, "可执行文件：%s\n", state.ExecutablePath)
	}
	if state.ConfigPath != "" {
		fmt.Fprintf(w, "配置：%s\n", state.ConfigPath)
	}
	if state.StdoutLogPath != "" {
		fmt.Fprintf(w, "stdout：%s\n", state.StdoutLogPath)
	}
	if state.StderrLogPath != "" {
		fmt.Fprintf(w, "stderr：%s\n", state.StderrLogPath)
	}
	if state.LastError != "" {
		fmt.Fprintf(w, "错误：%s\n", state.LastError)
	}
}
