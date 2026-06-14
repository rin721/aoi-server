package cliapp

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/rei0721/go-scaffold/pkg/cli"
)

// TestRunStartFlowUsesStdinBackedCLIUI 固定启动业务流程可以由 pkg/cli 的 stdin UI 驱动。
func TestRunStartFlowUsesStdinBackedCLIUI(t *testing.T) {
	configPath := copyExampleConfig(t)
	var stdout bytes.Buffer
	ctx := &cli.Context{
		Context: context.Background(),
		Flags: map[string]interface{}{
			"config": configPath,
		},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &stdout,
		UI:           cli.NewPromptUI(strings.NewReader("2\n"), &stdout),
	}

	if err := RunStartFlow(ctx); err != nil {
		t.Fatalf("RunStartFlow() error = %v", err)
	}
	out := stdout.String()
	for _, want := range []string{"db", "sqlite", "v1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdin-backed start flow output missing %q:\n%s", want, out)
		}
	}
}

// TestRunServiceFlowUsesStdinBackedCLIUI 固定服务管理流程可以由 pkg/cli 的 stdin UI 驱动。
func TestRunServiceFlowUsesStdinBackedCLIUI(t *testing.T) {
	manager := testManager(t, &fakeProcessRunner{})
	oldNewFlowManager := newFlowManager
	newFlowManager = func() *Manager {
		return manager
	}
	t.Cleanup(func() {
		newFlowManager = oldNewFlowManager
	})

	var stdout bytes.Buffer
	ctx := &cli.Context{
		Context: context.Background(),
		Stdout:  &stdout,
		UI:      cli.NewPromptUI(strings.NewReader("1\n7\n"), &stdout),
	}

	if err := RunServiceFlow(ctx); err != nil {
		t.Fatalf("RunServiceFlow() error = %v", err)
	}
	out := stdout.String()
	for _, want := range []string{ServiceServer, StatusStopped} {
		if !strings.Contains(out, want) {
			t.Fatalf("stdin-backed service flow output missing %q:\n%s", want, out)
		}
	}
}

// TestRunInitializationFlowUsesStdinBackedCLIUI 固定初始化业务流程可以由 pkg/cli 的 stdin UI 收集入参。
func TestRunInitializationFlowUsesStdinBackedCLIUI(t *testing.T) {
	configPath := copyExampleConfig(t)
	oldExecuteInitialization := executeInitialization
	var captured InitializationInput
	executeInitialization = func(_ context.Context, _ io.Writer, input InitializationInput) error {
		captured = input
		return nil
	}
	t.Cleanup(func() {
		executeInitialization = oldExecuteInitialization
	})

	stdin := strings.NewReader(strings.Join([]string{
		"orgx",
		"Org X",
		"root",
		"root@example.com",
		"Root User",
		"secret-password",
		"y",
		"14",
		"bootstrap token",
	}, "\n") + "\n")
	var stdout bytes.Buffer
	ctx := &cli.Context{
		Context: context.Background(),
		Flags: map[string]interface{}{
			"config": configPath,
		},
		ChangedFlags: map[string]bool{"config": true},
		Stdout:       &stdout,
		UI:           cli.NewPromptUI(stdin, &stdout),
	}

	if err := RunInitializationFlow(ctx, InitializationInput{}); err != nil {
		t.Fatalf("RunInitializationFlow() error = %v", err)
	}
	if captured.ConfigPath != configPath ||
		captured.OrgCode != "orgx" ||
		captured.OrgName != "Org X" ||
		captured.AdminUsername != "root" ||
		captured.AdminEmail != "root@example.com" ||
		captured.AdminDisplayName != "Root User" ||
		captured.AdminPassword != "secret-password" ||
		!captured.CreateServiceToken ||
		captured.ServiceTokenDays != 14 ||
		captured.ServiceTokenRemark != "bootstrap token" {
		t.Fatalf("captured initialization input = %#v", captured)
	}
}

func TestRunInitializationFlowUsesChainAnswers(t *testing.T) {
	configPath := copyExampleConfig(t)
	oldExecuteInitialization := executeInitialization
	var captured InitializationInput
	executeInitialization = func(_ context.Context, _ io.Writer, input InitializationInput) error {
		captured = input
		return nil
	}
	t.Cleanup(func() {
		executeInitialization = oldExecuteInitialization
	})

	var stdout bytes.Buffer
	ctx := &cli.Context{
		Context: context.Background(),
		Stdout:  &stdout,
		UI: cli.WithPromptAnswers(cli.NewPromptUI(strings.NewReader(""), &stdout), map[string]string{
			"config":               configPath,
			"org-code":             "chain-org",
			"org-name":             "Chain Org",
			"admin-username":       "chain-admin",
			"admin-email":          "chain@example.com",
			"admin-display-name":   "Chain Admin",
			"admin-password":       "chain-password",
			"create-service-token": "true",
			"service-token-days":   "21",
			"service-token-remark": "chain token",
		}),
	}

	if err := RunInitializationFlow(ctx, InitializationInput{}); err != nil {
		t.Fatalf("RunInitializationFlow() error = %v", err)
	}
	if captured.ConfigPath != configPath ||
		captured.OrgCode != "chain-org" ||
		captured.OrgName != "Chain Org" ||
		captured.AdminUsername != "chain-admin" ||
		captured.AdminEmail != "chain@example.com" ||
		captured.AdminDisplayName != "Chain Admin" ||
		captured.AdminPassword != "chain-password" ||
		!captured.CreateServiceToken ||
		captured.ServiceTokenDays != 21 ||
		captured.ServiceTokenRemark != "chain token" {
		t.Fatalf("captured initialization input = %#v", captured)
	}
}
