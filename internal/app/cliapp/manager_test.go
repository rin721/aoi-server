package cliapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/cli"
)

// TestManagerStartServerPersistsStateAndLaunchesManagedProcess 固定后台启动写入运行态并派生 server 子进程。
func TestManagerStartServerPersistsStateAndLaunchesManagedProcess(t *testing.T) {
	runner := &fakeProcessRunner{
		startInfos:     []ProcessInfo{{PID: 321, ProcessStartTime: 12345}},
		runningResults: []bool{true, true},
	}
	manager := testManager(t, runner)
	configPath := copyExampleConfig(t)

	state, err := manager.StartServer(context.Background(), configPath)
	if err != nil {
		t.Fatalf("StartServer() error = %v", err)
	}

	if state.Status != StatusRunning {
		t.Fatalf("status = %q, want %q", state.Status, StatusRunning)
	}
	if state.PID != 321 || state.ProcessStartTime != 12345 {
		t.Fatalf("process info = pid %d start %d", state.PID, state.ProcessStartTime)
	}
	if state.ConfigPath != filepath.Clean(configPath) {
		t.Fatalf("config path = %q, want %q", state.ConfigPath, filepath.Clean(configPath))
	}
	if !strings.HasSuffix(filepath.ToSlash(state.StdoutLogPath), "/server/stdout.log") {
		t.Fatalf("stdout log path = %q", state.StdoutLogPath)
	}
	if !strings.HasSuffix(filepath.ToSlash(state.StderrLogPath), "/server/stderr.log") {
		t.Fatalf("stderr log path = %q", state.StderrLogPath)
	}

	if len(runner.starts) != 1 {
		t.Fatalf("StartProcess calls = %d, want 1", len(runner.starts))
	}
	start := runner.starts[0]
	if start.Executable != manager.Executable {
		t.Fatalf("executable = %q, want %q", start.Executable, manager.Executable)
	}
	if state.ExecutablePath != manager.Executable {
		t.Fatalf("executable path = %q, want %q", state.ExecutablePath, manager.Executable)
	}
	wantArgs := []string{"server", "--config", filepath.Clean(configPath)}
	if !reflect.DeepEqual(start.Args, wantArgs) {
		t.Fatalf("args = %#v, want %#v", start.Args, wantArgs)
	}
	for _, want := range []string{ManagedServiceEnvName + "=1", ManagedServiceNameEnvKey + "=" + ServiceServer, RuntimeDirEnvName + "="} {
		if !envContainsPrefix(start.Env, want) {
			t.Fatalf("env missing prefix %q: %#v", want, start.Env)
		}
	}

	persisted, err := manager.readState(ServiceServer)
	if err != nil {
		t.Fatalf("readState() error = %v", err)
	}
	if persisted.Status != StatusRunning || persisted.PID != 321 {
		t.Fatalf("persisted state = %#v", persisted)
	}
	if persisted.ExecutablePath != manager.Executable {
		t.Fatalf("persisted executable path = %q, want %q", persisted.ExecutablePath, manager.Executable)
	}

	refreshed, err := manager.Status(context.Background(), ServiceServer)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if refreshed.Status != StatusRunning {
		t.Fatalf("refreshed status = %q", refreshed.Status)
	}
}

func TestManagerStartServerCopiesGoRunTemporaryExecutable(t *testing.T) {
	exeName := "main"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}
	source := filepath.Join(t.TempDir(), "go-build123456789", "b001", "exe", exeName)
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatalf("create temp exe dir: %v", err)
	}
	if err := os.WriteFile(source, []byte("managed executable"), 0o755); err != nil {
		t.Fatalf("write temp exe: %v", err)
	}

	runner := &fakeProcessRunner{
		startInfos:     []ProcessInfo{{PID: 654, ProcessStartTime: 98765}},
		runningResults: []bool{true, true},
	}
	manager := testManager(t, runner)
	manager.Executable = source
	configPath := copyExampleConfig(t)

	state, err := manager.StartServer(context.Background(), configPath)
	if err != nil {
		t.Fatalf("StartServer() error = %v", err)
	}

	wantExecutable := filepath.Join(manager.RuntimeDir, "bin", managedExecutableFileName(source))
	if len(runner.starts) != 1 {
		t.Fatalf("StartProcess calls = %d, want 1", len(runner.starts))
	}
	if runner.starts[0].Executable != wantExecutable {
		t.Fatalf("managed executable = %q, want %q", runner.starts[0].Executable, wantExecutable)
	}
	if state.ExecutablePath != wantExecutable {
		t.Fatalf("state executable path = %q, want %q", state.ExecutablePath, wantExecutable)
	}
	raw, err := os.ReadFile(wantExecutable)
	if err != nil {
		t.Fatalf("read managed executable: %v", err)
	}
	if string(raw) != "managed executable" {
		t.Fatalf("managed executable content = %q", raw)
	}

	persisted, err := manager.readState(ServiceServer)
	if err != nil {
		t.Fatalf("readState() error = %v", err)
	}
	if persisted.ExecutablePath != wantExecutable {
		t.Fatalf("persisted executable path = %q, want %q", persisted.ExecutablePath, wantExecutable)
	}
}

// TestManagerStatusMarksDeadActiveProcessFailed 固定 PID 创建时间校验失败时不误判为运行中。
func TestManagerStatusMarksDeadActiveProcessFailed(t *testing.T) {
	runner := &fakeProcessRunner{runningResults: []bool{false}}
	manager := testManager(t, runner)
	startedAt := time.Date(2026, 6, 13, 1, 2, 3, 0, time.UTC)
	if err := manager.writeState(ServiceState{
		Service:          ServiceServer,
		Status:           StatusRunning,
		PID:              88,
		ProcessStartTime: 9900,
		StartedAt:        &startedAt,
	}); err != nil {
		t.Fatalf("writeState() error = %v", err)
	}

	state, err := manager.Status(context.Background(), ServiceServer)
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if state.Status != StatusFailed {
		t.Fatalf("status = %q, want %q", state.Status, StatusFailed)
	}
	if state.PID != 0 || state.ProcessStartTime != 0 {
		t.Fatalf("expected process info cleared, got pid=%d start=%d", state.PID, state.ProcessStartTime)
	}
	if state.LastError != "process is not running" {
		t.Fatalf("lastError = %q", state.LastError)
	}
	if len(runner.checks) != 1 || runner.checks[0].ProcessStartTime != 9900 {
		t.Fatalf("process checks = %#v", runner.checks)
	}
}

// TestManagerStopServerWritesControlAndClearsStateWhenProcessExits 固定停止流程先写 control，再等待进程退出。
func TestManagerStopServerWritesControlAndClearsStateWhenProcessExits(t *testing.T) {
	runner := &fakeProcessRunner{runningResults: []bool{true, false}}
	manager := testManager(t, runner)
	var captured ControlRequest
	runner.onCheck = func(_ ProcessInfo, call int) {
		if call != 2 {
			return
		}
		raw, err := os.ReadFile(manager.controlPath(ServiceServer))
		if err != nil {
			t.Fatalf("read control file: %v", err)
		}
		if err := json.Unmarshal(raw, &captured); err != nil {
			t.Fatalf("decode control file: %v", err)
		}
	}

	startedAt := time.Date(2026, 6, 13, 1, 2, 3, 0, time.UTC)
	if err := manager.writeState(ServiceState{
		Service:          ServiceServer,
		Status:           StatusRunning,
		PID:              98,
		ProcessStartTime: 123456,
		StartedAt:        &startedAt,
		ConfigPath:       "configs/config.yaml",
	}); err != nil {
		t.Fatalf("writeState() error = %v", err)
	}

	state, err := manager.StopServer(context.Background())
	if err != nil {
		t.Fatalf("StopServer() error = %v", err)
	}
	if state.Status != StatusStopped {
		t.Fatalf("status = %q, want %q", state.Status, StatusStopped)
	}
	if state.PID != 0 || state.ProcessStartTime != 0 {
		t.Fatalf("expected process info cleared, got pid=%d start=%d", state.PID, state.ProcessStartTime)
	}
	if len(runner.kills) != 0 {
		t.Fatalf("KillProcess calls = %#v", runner.kills)
	}
	if captured.Service != ServiceServer || captured.Action != controlActionStop || captured.PID != 98 || captured.ProcessStartTime != 123456 {
		t.Fatalf("control request = %#v", captured)
	}
	if _, err := os.Stat(manager.controlPath(ServiceServer)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("control file should be removed after stop, err=%v", err)
	}
}

// TestManagerRestartServerUsesLastConfig 固定重启沿用上次配置路径并重新启动后台进程。
func TestManagerRestartServerUsesLastConfig(t *testing.T) {
	configPath := copyExampleConfig(t)
	runner := &fakeProcessRunner{
		startInfos:     []ProcessInfo{{PID: 333, ProcessStartTime: 777}},
		runningResults: []bool{true, true, false, true},
	}
	manager := testManager(t, runner)
	if err := manager.writeState(ServiceState{
		Service:          ServiceServer,
		Status:           StatusRunning,
		PID:              98,
		ProcessStartTime: 123456,
		ConfigPath:       configPath,
	}); err != nil {
		t.Fatalf("writeState() error = %v", err)
	}

	state, err := manager.RestartServer(context.Background())
	if err != nil {
		t.Fatalf("RestartServer() error = %v", err)
	}
	if state.Status != StatusRunning || state.PID != 333 {
		t.Fatalf("restart state = %#v", state)
	}
	if len(runner.starts) != 1 {
		t.Fatalf("StartProcess calls = %d, want 1", len(runner.starts))
	}
	wantArgs := []string{"server", "--config", filepath.Clean(configPath)}
	if !reflect.DeepEqual(runner.starts[0].Args, wantArgs) {
		t.Fatalf("restart args = %#v, want %#v", runner.starts[0].Args, wantArgs)
	}
}

// TestPrintServiceLogsReadsHistoryAndFollowDetachesOnContext 固定日志终端只是附着输出，不参与服务停止。
func TestPrintServiceLogsReadsHistoryAndFollowDetachesOnContext(t *testing.T) {
	dir := t.TempDir()
	stdoutPath := filepath.Join(dir, "stdout.log")
	stderrPath := filepath.Join(dir, "stderr.log")
	if err := os.WriteFile(stdoutPath, []byte("out-1\nout-2\nout-3\n"), 0o644); err != nil {
		t.Fatalf("write stdout log: %v", err)
	}
	if err := os.WriteFile(stderrPath, []byte("err-1\nerr-2\n"), 0o644); err != nil {
		t.Fatalf("write stderr log: %v", err)
	}

	var out bytes.Buffer
	state := ServiceState{StdoutLogPath: stdoutPath, StderrLogPath: stderrPath}
	if err := PrintServiceLogs(context.Background(), &out, state, 2, false); err != nil {
		t.Fatalf("PrintServiceLogs(history) error = %v", err)
	}
	text := out.String()
	for _, want := range []string{"out-2", "out-3", "err-1", "err-2"} {
		if !strings.Contains(text, want) {
			t.Fatalf("history output missing %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "out-1") {
		t.Fatalf("history output should have tailed stdout lines:\n%s", text)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := PrintServiceLogs(ctx, &out, state, 2, true); err != nil {
		t.Fatalf("PrintServiceLogs(follow) error = %v", err)
	}
}

// TestRunStartFlowShowsDependencyServicesThroughCLIUI 固定启动流程通过 pkg/cli UI 抽象收集输入。
func TestRunStartFlowShowsDependencyServicesThroughCLIUI(t *testing.T) {
	configPath := copyExampleConfig(t)
	var stdout bytes.Buffer
	ctx := &cli.Context{
		Context: context.Background(),
		Flags: map[string]interface{}{
			"config": configPath,
		},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &stdout,
		UI:           &fakePromptUI{selects: []string{"db"}},
	}

	if err := RunStartFlow(ctx); err != nil {
		t.Fatalf("RunStartFlow() error = %v", err)
	}
	out := stdout.String()
	for _, want := range []string{"配置文件", "数据库", "db：driver=sqlite", "v1 仅托管 server"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRunStartFlowWithChainAnswersStartsServerNonInteractively(t *testing.T) {
	configPath := copyExampleConfig(t)
	runner := &fakeProcessRunner{
		startInfos:     []ProcessInfo{{PID: 321, ProcessStartTime: 12345}},
		runningResults: []bool{true},
	}
	restoreFlowManager(t, testManager(t, runner))
	ui := &fakePromptUI{
		selects:  []string{"db"},
		confirms: []bool{true},
	}
	var stdout bytes.Buffer
	ctx := &cli.Context{
		Context: context.Background(),
		Stdout:  &stdout,
		UI: cli.WithPromptAnswers(ui, map[string]string{
			"service": ServiceServer,
			"config":  configPath,
			"privacy": "false",
		}),
	}

	if err := RunStartFlow(ctx); err != nil {
		t.Fatalf("RunStartFlow() error = %v", err)
	}

	if len(runner.starts) != 1 {
		t.Fatalf("StartProcess calls = %d, want 1", len(runner.starts))
	}
	wantArgs := []string{"server", "--config", filepath.Clean(configPath)}
	if !reflect.DeepEqual(runner.starts[0].Args, wantArgs) {
		t.Fatalf("start args = %#v, want %#v", runner.starts[0].Args, wantArgs)
	}
	if len(ui.selects) != 1 {
		t.Fatalf("service select should not be consumed in direct input mode")
	}
	if len(ui.confirms) != 1 {
		t.Fatalf("privacy confirm should not be consumed when skipped")
	}
	out := stdout.String()
	for _, want := range []string{ServiceServer, StatusRunning} {
		if !strings.Contains(out, want) {
			t.Fatalf("direct start output missing %q:\n%s", want, out)
		}
	}
}

func TestRunStartFlowRepairsMissingCoreSecretsBeforeSummary(t *testing.T) {
	unsetCoreSecretEnvForTest(t)
	configPath := copyEnvManagedCoreSecretsConfig(t)
	runner := &fakeProcessRunner{
		startInfos:     []ProcessInfo{{PID: 321, ProcessStartTime: 12345}},
		runningResults: []bool{true},
	}
	restoreFlowManager(t, testManager(t, runner))
	ui := &fakePromptUI{
		selects: []string{ServiceServer, privacyCoreActionGenerateFile},
	}
	var stdout bytes.Buffer
	ctx := &cli.Context{
		Context:      context.Background(),
		Flags:        map[string]interface{}{"config": configPath},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &stdout,
		UI:           ui,
	}

	if err := RunStartFlow(ctx); err != nil {
		t.Fatalf("RunStartFlow() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	text := string(content)
	for _, placeholder := range []string{"${RIN_APP_AUTH_SIGNING_KEY}", "${RIN_APP_AUTH_REFRESH_TOKEN_PEPPER}", "${RIN_APP_AUTH_MFA_SECRET_KEY}"} {
		if strings.Contains(text, placeholder) {
			t.Fatalf("core secret placeholder %s should be overwritten:\n%s", placeholder, text)
		}
	}

	manager := appconfig.NewManager()
	if err := manager.Load(configPath); err != nil {
		t.Fatalf("reload repaired config: %v", err)
	}
	cfg := manager.Get()
	if len(cfg.Auth.SigningKey) < 32 || cfg.Auth.RefreshTokenPepper == "" || len(cfg.Auth.MFASecretKey) < 32 {
		t.Fatalf("generated core secrets are invalid: %#v", cfg.Auth)
	}
	for _, path := range coreSecretPaths {
		if !stringSliceContains(cfg.EnvOverride.DisabledPaths, path) {
			t.Fatalf("disabled_paths missing %q: %#v", path, cfg.EnvOverride.DisabledPaths)
		}
	}
	if len(runner.starts) != 1 {
		t.Fatalf("StartProcess calls = %d, want 1", len(runner.starts))
	}
	if !strings.Contains(stdout.String(), StatusRunning) {
		t.Fatalf("stdout missing running status:\n%s", stdout.String())
	}
}

func TestRunStartFlowRuntimeEnvOnlyMissingCoreSecretDoesNotWriteConfig(t *testing.T) {
	unsetCoreSecretEnvForTest(t)
	configPath := copyEnvManagedCoreSecretsConfig(t)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before run: %v", err)
	}
	runner := &fakeProcessRunner{}
	restoreFlowManager(t, testManager(t, runner))
	ctx := &cli.Context{
		Context:      context.Background(),
		Flags:        map[string]interface{}{"config": configPath},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &bytes.Buffer{},
		UI: &fakePromptUI{
			selects: []string{ServiceServer, privacyActionRuntimeEnvOnly},
		},
	}

	err = RunStartFlow(ctx)
	if err == nil || !strings.Contains(err.Error(), "RIN_APP_AUTH_") {
		t.Fatalf("expected missing environment error, got %v", err)
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after run: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("failed runtime env-only repair should not write config\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if len(runner.starts) != 0 {
		t.Fatalf("StartProcess calls = %d, want 0", len(runner.starts))
	}
}

func TestManagerStartServerMissingCoreSecretsReturnsActionableError(t *testing.T) {
	unsetCoreSecretEnvForTest(t)
	configPath := copyEnvManagedCoreSecretsConfig(t)
	runner := &fakeProcessRunner{}
	manager := testManager(t, runner)

	state, err := manager.StartServer(context.Background(), configPath)
	if err == nil {
		t.Fatal("StartServer() error = nil, want missing secret error")
	}
	for _, want := range []string{"RIN_APP_AUTH_SIGNING_KEY", "interactive `run`"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("StartServer() error missing %q:\n%v", want, err)
		}
	}
	if state.Status != StatusFailed || state.LastError == "" {
		t.Fatalf("state = %#v, want failed state with last error", state)
	}
	if len(runner.starts) != 0 {
		t.Fatalf("StartProcess calls = %d, want 0", len(runner.starts))
	}
}

func TestRunStartFlowForceWritesGeneratedEnvManagedPrivacy(t *testing.T) {
	envNames := appconfig.EnvNamesForPath("auth.signing_key")
	unsetEnvForTest(t, envNames...)
	if len(envNames) == 0 {
		t.Fatal("auth.signing_key should expose environment names")
	}
	t.Setenv(envNames[0], "environment-signing-secret-at-least-32-bytes")
	configPath := copyExampleConfig(t)
	runner := &fakeProcessRunner{
		startInfos:     []ProcessInfo{{PID: 321, ProcessStartTime: 12345}},
		runningResults: []bool{true},
	}
	restoreFlowManager(t, testManager(t, runner))
	ui := &fakePromptUI{
		selects:  []string{ServiceServer, privacyActionForceFile, privacyActionSkip, privacyActionSkip},
		confirms: []bool{true},
		inputs:   []string{"generate"},
	}
	ctx := &cli.Context{
		Context:      context.Background(),
		Flags:        map[string]interface{}{"config": configPath},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &bytes.Buffer{},
		UI:           ui,
	}

	if err := RunStartFlow(ctx); err != nil {
		t.Fatalf("RunStartFlow() error = %v", err)
	}
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	text := string(content)
	if strings.Contains(text, "${AUTH_SIGNING_KEY") || strings.Contains(text, "dev-signing-key-change-me-32-bytes") {
		t.Fatalf("force generated signing key should replace placeholder:\n%s", text)
	}
	if !strings.Contains(text, `signing_key: "`) {
		t.Fatalf("force generated signing key should be persisted as a quoted scalar:\n%s", text)
	}
	if !strings.Contains(text, `- "auth.signing_key"`) {
		t.Fatalf("force generated signing key should disable env override:\n%s", text)
	}
	manager := appconfig.NewManager()
	if err := manager.Load(configPath); err != nil {
		t.Fatalf("reload forced config: %v", err)
	}
	if got := manager.Get().Auth.SigningKey; got == "environment-signing-secret-at-least-32-bytes" {
		t.Fatalf("forced config should use file value instead of environment value")
	}
	if len(runner.starts) != 1 {
		t.Fatalf("StartProcess calls = %d, want 1", len(runner.starts))
	}
	if len(ui.infos) != 1 {
		t.Fatalf("Info calls = %#v, want one privacy completion message", ui.infos)
	}
}

func TestRunStartFlowRuntimeEnvOnlyRestoresEnvOverride(t *testing.T) {
	envNames := appconfig.EnvNamesForPath("auth.signing_key")
	unsetEnvForTest(t, envNames...)
	if len(envNames) == 0 {
		t.Fatal("auth.signing_key should expose environment names")
	}
	t.Setenv(envNames[0], "runtime-env-signing-secret-at-least-32-bytes")
	configPath := copyExampleConfigWithEnvOverride(t, []string{"auth.signing_key"})
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before run: %v", err)
	}
	runner := &fakeProcessRunner{
		startInfos:     []ProcessInfo{{PID: 321, ProcessStartTime: 12345}},
		runningResults: []bool{true},
	}
	restoreFlowManager(t, testManager(t, runner))
	ctx := &cli.Context{
		Context: context.Background(),
		Flags: map[string]interface{}{
			"config": configPath,
		},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &bytes.Buffer{},
		UI: &fakePromptUI{
			selects:  []string{ServiceServer, privacyActionRuntimeEnvOnly, privacyActionSkip, privacyActionSkip},
			confirms: []bool{true},
		},
	}

	if err := RunStartFlow(ctx); err != nil {
		t.Fatalf("RunStartFlow() error = %v", err)
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after run: %v", err)
	}
	if string(after) == string(before) {
		t.Fatalf("runtime env-only flow should remove disabled env override metadata")
	}
	if strings.Contains(string(after), `- "auth.signing_key"`) {
		t.Fatalf("runtime env-only flow should restore env override for signing key:\n%s", after)
	}
	if len(runner.starts) != 1 {
		t.Fatalf("StartProcess calls = %d, want 1", len(runner.starts))
	}
}

func TestRunStartFlowRuntimeEnvOnlyRejectsMissingEnv(t *testing.T) {
	unsetEnvForTest(t, appconfig.EnvNamesForPath("auth.signing_key")...)
	configPath := copyExampleConfig(t)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before run: %v", err)
	}
	runner := &fakeProcessRunner{}
	restoreFlowManager(t, testManager(t, runner))
	ctx := &cli.Context{
		Context:      context.Background(),
		Flags:        map[string]interface{}{"config": configPath},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &bytes.Buffer{},
		UI: &fakePromptUI{
			selects:  []string{ServiceServer, privacyActionRuntimeEnvOnly, privacyActionSkip, privacyActionSkip},
			confirms: []bool{true},
		},
	}

	err = RunStartFlow(ctx)
	if err == nil || !strings.Contains(err.Error(), "set one of") {
		t.Fatalf("expected missing environment error, got %v", err)
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after run: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("failed runtime env-only flow should not write config\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if len(runner.starts) != 0 {
		t.Fatalf("StartProcess calls = %d, want 0", len(runner.starts))
	}
}

func TestRunStartFlowPreflightRepairsProductionConfigBeforeStart(t *testing.T) {
	unsetPreflightEnvForTest(t)
	configPath := copyProductionConfig(t)
	runner := &fakeProcessRunner{
		startInfos:     []ProcessInfo{{PID: 321, ProcessStartTime: 12345}},
		runningResults: []bool{true},
	}
	restoreFlowManager(t, testManager(t, runner))
	ctx := &cli.Context{
		Context:      context.Background(),
		Flags:        map[string]interface{}{"config": configPath},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &bytes.Buffer{},
		UI: &fakePromptUI{
			selects: []string{
				ServiceServer,
				preflightDatabaseActionSQLite,
				privacyCoreActionGenerateFile,
				preflightSMTPActionDebug,
			},
		},
	}

	if err := RunStartFlow(ctx); err != nil {
		t.Fatalf("RunStartFlow() error = %v", err)
	}

	manager := appconfig.NewManager()
	if err := manager.Load(configPath); err != nil {
		t.Fatalf("reload repaired config: %v", err)
	}
	cfg := manager.Get()
	if cfg.Database.Driver != "sqlite" {
		t.Fatalf("database driver = %q, want sqlite", cfg.Database.Driver)
	}
	if cfg.Auth.NotificationDriver != "debug" {
		t.Fatalf("notification driver = %q, want debug", cfg.Auth.NotificationDriver)
	}
	if len(cfg.Auth.SigningKey) < 32 || cfg.Auth.RefreshTokenPepper == "" || len(cfg.Auth.MFASecretKey) < 32 {
		t.Fatalf("generated core secrets are invalid: %#v", cfg.Auth)
	}
	for _, path := range []string{"database.driver", "auth.notification_driver"} {
		if !stringSliceContains(cfg.EnvOverride.DisabledPaths, path) {
			t.Fatalf("disabled_paths missing %q: %#v", path, cfg.EnvOverride.DisabledPaths)
		}
	}
	for _, path := range coreSecretPaths {
		if !stringSliceContains(cfg.EnvOverride.DisabledPaths, path) {
			t.Fatalf("disabled_paths missing core secret %q: %#v", path, cfg.EnvOverride.DisabledPaths)
		}
	}
	if len(runner.starts) != 1 {
		t.Fatalf("StartProcess calls = %d, want 1", len(runner.starts))
	}
}

func TestRunStartFlowPreflightRuntimeEnvOnlyMissingDatabaseEnvDoesNotWriteConfig(t *testing.T) {
	unsetPreflightEnvForTest(t)
	configPath := copyProductionConfig(t)
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before run: %v", err)
	}
	runner := &fakeProcessRunner{}
	restoreFlowManager(t, testManager(t, runner))
	ctx := &cli.Context{
		Context:      context.Background(),
		Flags:        map[string]interface{}{"config": configPath},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &bytes.Buffer{},
		UI: &fakePromptUI{
			selects: []string{ServiceServer, preflightActionRuntimeEnvOnly},
		},
	}

	err = RunStartFlow(ctx)
	if err == nil || !strings.Contains(err.Error(), "RIN_APP_DB_HOST") {
		t.Fatalf("expected missing database environment error, got %v", err)
	}
	after, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config after run: %v", err)
	}
	if string(after) != string(before) {
		t.Fatalf("failed runtime env-only preflight should not write config\nbefore:\n%s\nafter:\n%s", before, after)
	}
	if len(runner.starts) != 0 {
		t.Fatalf("StartProcess calls = %d, want 0", len(runner.starts))
	}
}

func TestManagerStartServerPreflightReturnsAllBlockingDiagnostics(t *testing.T) {
	unsetPreflightEnvForTest(t)
	configPath := copyProductionConfig(t)
	runner := &fakeProcessRunner{}
	manager := testManager(t, runner)

	state, err := manager.StartServer(context.Background(), configPath)
	if err == nil {
		t.Fatal("StartServer() error = nil, want preflight diagnostics")
	}
	for _, want := range []string{"database.host", "auth.signing_key", "auth.smtp.host", "interactive `run`"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("StartServer() error missing %q:\n%v", want, err)
		}
	}
	if state.Status != StatusFailed || state.LastError == "" {
		t.Fatalf("state = %#v, want failed state with last error", state)
	}
	if len(runner.starts) != 0 {
		t.Fatalf("StartProcess calls = %d, want 0", len(runner.starts))
	}
}

// TestControlRequestMatchingRequiresServicePIDAndCreateTime 固定控制文件只作用于匹配的托管进程。
func TestControlRequestMatchingRequiresServicePIDAndCreateTime(t *testing.T) {
	self := ProcessInfo{PID: 10, ProcessStartTime: 20}
	valid := ControlRequest{Service: ServiceServer, Action: controlActionStop, PID: 10, ProcessStartTime: 20}
	if !matchesCurrentProcess(valid, ServiceServer, self) {
		t.Fatal("expected matching control request")
	}

	cases := []ControlRequest{
		{Service: "db", Action: controlActionStop, PID: 10, ProcessStartTime: 20},
		{Service: ServiceServer, Action: "restart", PID: 10, ProcessStartTime: 20},
		{Service: ServiceServer, Action: controlActionStop, PID: 11, ProcessStartTime: 20},
		{Service: ServiceServer, Action: controlActionStop, PID: 10, ProcessStartTime: 21},
	}
	for _, tc := range cases {
		if matchesCurrentProcess(tc, ServiceServer, self) {
			t.Fatalf("unexpected match for %#v", tc)
		}
	}
}

func TestApplyPrivacyUpdatesPersistsSecretsAndRejectsEnvManagedFields(t *testing.T) {
	configPath := copyWritablePrivacyConfig(t)
	if err := ApplyPrivacyUpdates(configPath, map[string]string{
		"auth.signing_key":          "updated-signing-secret-at-least-32-bytes",
		"auth.refresh_token_pepper": "",
		"unsupported.path":          "ignored",
	}); err != nil {
		t.Fatalf("ApplyPrivacyUpdates() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read updated config: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `signing_key: "updated-signing-secret-at-least-32-bytes"`) {
		t.Fatalf("updated config missing persisted signing key:\n%s", text)
	}
	if strings.Contains(text, "unsupported.path") {
		t.Fatalf("unsupported path should not be persisted:\n%s", text)
	}

	t.Setenv("AUTH_SIGNING_KEY", "managed-by-env-signing-secret-at-least-32-bytes")
	err = ApplyPrivacyUpdates(configPath, map[string]string{
		"auth.signing_key": "another-secret-at-least-32-bytes",
	})
	if err == nil || !strings.Contains(err.Error(), "environment variable AUTH_SIGNING_KEY") {
		t.Fatalf("expected env-managed field error, got %v", err)
	}
}

func testManager(t *testing.T, runner *fakeProcessRunner) *Manager {
	t.Helper()
	return &Manager{
		RuntimeDir: filepath.Join(t.TempDir(), "runtime"),
		Executable: filepath.Join(t.TempDir(), "bin-test"),
		WorkDir:    t.TempDir(),
		Runner:     runner,
		Now: func() time.Time {
			return time.Date(2026, 6, 13, 1, 2, 3, 0, time.UTC)
		},
	}
}

func copyExampleConfig(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	raw, err := os.ReadFile(filepath.Join(root, "configs", "config.example.yaml"))
	if err != nil {
		t.Fatalf("read config example: %v", err)
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func copyProductionConfig(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	raw, err := os.ReadFile(filepath.Join(root, "deploy", "config.production.example.yaml"))
	if err != nil {
		t.Fatalf("read production config example: %v", err)
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write temp production config: %v", err)
	}
	return path
}

func copyEnvManagedCoreSecretsConfig(t *testing.T) string {
	t.Helper()
	path := copyExampleConfig(t)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp config: %v", err)
	}
	replacements := map[string]string{
		"signing_key: ${AUTH_SIGNING_KEY:dev-signing-key-change-me-32-bytes}":                "signing_key: ${RIN_APP_AUTH_SIGNING_KEY}",
		"refresh_token_pepper: ${AUTH_REFRESH_TOKEN_PEPPER:dev-refresh-pepper-change-me-32}": "refresh_token_pepper: ${RIN_APP_AUTH_REFRESH_TOKEN_PEPPER}",
		"mfa_secret_key: ${AUTH_MFA_SECRET_KEY:dev-mfa-secret-key-change-me-32-bytes}":       "mfa_secret_key: ${RIN_APP_AUTH_MFA_SECRET_KEY}",
	}
	text := string(raw)
	for oldValue, newValue := range replacements {
		next := strings.Replace(text, oldValue, newValue, 1)
		if next == text {
			t.Fatalf("config copy did not contain %q", oldValue)
		}
		text = next
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("write env managed core secrets config: %v", err)
	}
	return path
}

func copyWritablePrivacyConfig(t *testing.T) string {
	t.Helper()
	path := copyExampleConfig(t)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp config: %v", err)
	}
	text := strings.Replace(
		string(raw),
		"signing_key: ${AUTH_SIGNING_KEY:dev-signing-key-change-me-32-bytes}",
		"signing_key: writable-signing-secret-at-least-32-bytes",
		1,
	)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("write writable temp config: %v", err)
	}
	return path
}

func copyExampleConfigWithEnvOverride(t *testing.T, disabledPaths []string) string {
	t.Helper()
	path := copyExampleConfig(t)
	var builder strings.Builder
	builder.WriteString("disabled_paths:\n")
	for _, disabledPath := range disabledPaths {
		builder.WriteString("    - ")
		builder.WriteString(disabledPath)
		builder.WriteByte('\n')
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config copy: %v", err)
	}
	next := strings.Replace(string(raw), "disabled_paths: []", builder.String(), 1)
	if next == string(raw) {
		t.Fatalf("config copy did not contain env_override disabled_paths placeholder")
	}
	if err := os.WriteFile(path, []byte(next), 0o644); err != nil {
		t.Fatalf("write env override config copy: %v", err)
	}
	return path
}

func restoreFlowManager(t *testing.T, manager *Manager) {
	t.Helper()
	previous := newFlowManager
	newFlowManager = func() *Manager {
		return manager
	}
	t.Cleanup(func() {
		newFlowManager = previous
	})
}

func unsetEnvForTest(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		key := key
		oldValue, existed := os.LookupEnv(key)
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
		t.Cleanup(func() {
			if existed {
				if err := os.Setenv(key, oldValue); err != nil {
					t.Errorf("restore %s: %v", key, err)
				}
				return
			}
			if err := os.Unsetenv(key); err != nil {
				t.Errorf("restore unset %s: %v", key, err)
			}
		})
	}
}

func unsetCoreSecretEnvForTest(t *testing.T) {
	t.Helper()
	for _, path := range coreSecretPaths {
		unsetEnvForTest(t, appconfig.EnvNamesForPath(path)...)
	}
}

func unsetPreflightEnvForTest(t *testing.T) {
	t.Helper()
	for _, path := range []string{
		"database.driver",
		"database.host",
		"database.port",
		"database.user",
		"database.dbname",
		"auth.signing_key",
		"auth.refresh_token_pepper",
		"auth.mfa_secret_key",
		"auth.notification_driver",
		"auth.smtp.host",
		"auth.smtp.port",
		"auth.smtp.from",
	} {
		unsetEnvForTest(t, appconfig.EnvNamesForPath(path)...)
	}
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func envContainsPrefix(values []string, prefix string) bool {
	for _, value := range values {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

type fakeProcessRunner struct {
	startInfos     []ProcessInfo
	runningResults []bool
	starts         []ProcessStartRequest
	checks         []ProcessInfo
	kills          []ProcessInfo
	onCheck        func(ProcessInfo, int)
}

func (f *fakeProcessRunner) StartProcess(req ProcessStartRequest) (ProcessInfo, error) {
	f.starts = append(f.starts, req)
	if len(f.startInfos) == 0 {
		return ProcessInfo{PID: 100 + len(f.starts), ProcessStartTime: int64(1000 + len(f.starts))}, nil
	}
	info := f.startInfos[0]
	f.startInfos = f.startInfos[1:]
	return info, nil
}

func (f *fakeProcessRunner) IsProcessRunning(info ProcessInfo) (bool, error) {
	f.checks = append(f.checks, info)
	if f.onCheck != nil {
		f.onCheck(info, len(f.checks))
	}
	if len(f.runningResults) == 0 {
		return true, nil
	}
	running := f.runningResults[0]
	f.runningResults = f.runningResults[1:]
	return running, nil
}

func (f *fakeProcessRunner) KillProcess(info ProcessInfo) error {
	f.kills = append(f.kills, info)
	return nil
}

type fakePromptUI struct {
	selects  []string
	confirms []bool
	inputs   []string
	infos    []string
}

func (f *fakePromptUI) Select(context.Context, string, []cli.SelectOption) (string, error) {
	if len(f.selects) == 0 {
		return "", nil
	}
	value := f.selects[0]
	f.selects = f.selects[1:]
	return value, nil
}

func (f *fakePromptUI) Confirm(context.Context, string, bool) (bool, error) {
	if len(f.confirms) == 0 {
		return false, nil
	}
	value := f.confirms[0]
	f.confirms = f.confirms[1:]
	return value, nil
}

func (f *fakePromptUI) Input(context.Context, string, string) (string, error) {
	if len(f.inputs) == 0 {
		return "", nil
	}
	value := f.inputs[0]
	f.inputs = f.inputs[1:]
	return value, nil
}

func (f *fakePromptUI) Password(context.Context, string) (string, error) {
	return "", nil
}

func (f *fakePromptUI) Info(message string) error {
	f.infos = append(f.infos, message)
	return nil
}
