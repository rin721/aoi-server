package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/types/constants"
)

// TestRunCLIWithoutArgsUsesPkgCLIInteractiveHome 固定无参数 bin 走 pkg/cli 的 Bubble Tea 首页。
func TestRunCLIWithoutArgsUsesPkgCLIInteractiveHome(t *testing.T) {
	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := runCLI(ctx, nil, strings.NewReader("q"), &stdout, &stderr)
	var cancelled *cli.CancelledError
	if !errors.As(err, &cancelled) {
		t.Fatalf("runCLI() error = %T %v, want *cli.CancelledError", err, err)
	}
}

// TestNewCLIAppRegistersLegacyAndSystemCommands 固定旧入口兼容性和新指令中心命令注册。
func TestNewCLIAppRegistersLegacyAndSystemCommands(t *testing.T) {
	app, err := newCLIApp()
	if err != nil {
		t.Fatalf("newCLIApp() error = %v", err)
	}

	var stdout bytes.Buffer
	if err := app.RunWithIO(context.Background(), []string{"--help"}, strings.NewReader(""), &stdout, io.Discard); err != nil {
		t.Fatalf("root help error = %v", err)
	}
	rootHelp := stdout.String()
	for _, want := range []string{"server", "db", "iam", "run", "service", "init"} {
		if !strings.Contains(rootHelp, want) {
			t.Fatalf("root help missing command %q:\n%s", want, rootHelp)
		}
	}

	stdout.Reset()
	if err := app.RunWithIO(context.Background(), []string{"server", "--help"}, strings.NewReader(""), &stdout, io.Discard); err != nil {
		t.Fatalf("server help error = %v", err)
	}
	serverHelp := stdout.String()
	for _, want := range []string{"--config", constants.AppDefaultConfigPath, appconfig.EnvConfigPathName()} {
		if !strings.Contains(serverHelp, want) {
			t.Fatalf("server help missing %q:\n%s", want, serverHelp)
		}
	}

	stdout.Reset()
	if err := app.RunWithIO(context.Background(), []string{"iam", "bootstrap-admin", "--help"}, strings.NewReader(""), &stdout, io.Discard); err != nil {
		t.Fatalf("iam bootstrap-admin help error = %v", err)
	}
	iamHelp := stdout.String()
	for _, want := range []string{"--org-code", "--username", "--password-stdin"} {
		if !strings.Contains(iamHelp, want) {
			t.Fatalf("iam bootstrap-admin help missing %q:\n%s", want, iamHelp)
		}
	}

	stdout.Reset()
	if err := app.RunWithIO(context.Background(), []string{"service", "--help"}, strings.NewReader(""), &stdout, io.Discard); err != nil {
		t.Fatalf("service help error = %v", err)
	}
	serviceHelp := stdout.String()
	for _, want := range []string{"status", "info", "logs", "terminal", "restart", "stop"} {
		if !strings.Contains(serviceHelp, want) {
			t.Fatalf("service help missing %q:\n%s", want, serviceHelp)
		}
	}
}

// TestSystemCenterCommandSpecs 固定 run/service/init 的业务命令规格。
func TestSystemCenterCommandSpecs(t *testing.T) {
	specs := NewSystemCenterCommands()
	if len(specs) != 3 {
		t.Fatalf("len(NewSystemCenterCommands()) = %d, want 3", len(specs))
	}
	byName := map[string]cli.CommandSpec{}
	for _, spec := range specs {
		byName[spec.Name] = spec
	}
	for _, want := range []string{"run", "service", "init"} {
		if _, ok := byName[want]; !ok {
			t.Fatalf("missing command spec %q", want)
		}
	}
	if byName["run"].HomeLabel != "启动 / run" || byName["run"].HomeOrder != 10 {
		t.Fatalf("run home metadata = label %q order %d", byName["run"].HomeLabel, byName["run"].HomeOrder)
	}
	if byName["service"].HomeLabel != "服务 / service" || byName["service"].HomeOrder != 20 {
		t.Fatalf("service home metadata = label %q order %d", byName["service"].HomeLabel, byName["service"].HomeOrder)
	}
	if byName["init"].HomeLabel != "初始化 / init" || byName["init"].HomeOrder != 30 {
		t.Fatalf("init home metadata = label %q order %d", byName["init"].HomeLabel, byName["init"].HomeOrder)
	}

	if len(byName["run"].Flags) != 1 || byName["run"].Flags[0].Name != "config" {
		t.Fatalf("run parent config flag = %#v", byName["run"].Flags)
	}
	runServer, ok := findChildSpec(byName["run"], constants.AppServerCommandName)
	if !ok {
		t.Fatal("run command missing server child")
	}
	if len(runServer.Flags) != 1 || runServer.Flags[0].Name != "config" || runServer.Flags[0].Default != constants.AppDefaultConfigPath {
		t.Fatalf("run server config flag = %#v", runServer.Flags)
	}

	service := byName["service"]
	for _, want := range []string{"status", "info", "logs", "terminal", "restart", "stop"} {
		if _, ok := findChildSpec(service, want); !ok {
			t.Fatalf("service command missing %q child", want)
		}
	}

	initSpec := byName["init"]
	if len(initSpec.Flags) == 0 || initSpec.Flags[0].Name != "config" || initSpec.Flags[0].Default != constants.AppDefaultConfigPath {
		t.Fatalf("init config flag = %#v", initSpec.Flags)
	}
	for _, want := range []string{"admin-password", "admin-password-stdin", "create-service-token"} {
		if !hasFlag(initSpec.Flags, want) {
			t.Fatalf("init command missing flag %q", want)
		}
	}
}

func findChildSpec(parent cli.CommandSpec, name string) (cli.CommandSpec, bool) {
	for _, child := range parent.Commands {
		if child.Name == name {
			return child, true
		}
	}
	return cli.CommandSpec{}, false
}

func hasFlag(flags []cli.FlagSpec, name string) bool {
	for _, flag := range flags {
		if flag.Name == name {
			return true
		}
	}
	return false
}
