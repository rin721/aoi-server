package main

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/rei0721/go-scaffold/internal/app/buildapp"
	"github.com/rei0721/go-scaffold/pkg/cli"
)

func TestBuildCommandMetadata(t *testing.T) {
	cmd := NewBuildCommand()

	if got := cmd.Name(); got != "build" {
		t.Fatalf("Name() = %q, want build", got)
	}
	if !strings.Contains(cmd.Description(), "release packages") {
		t.Fatalf("Description() = %q, want release packages wording", cmd.Description())
	}
	if !strings.Contains(cmd.Usage(), "flags") {
		t.Fatalf("Usage() = %q, want flags wording", cmd.Usage())
	}

	flags := cmd.Flags()
	wantNames := []string{
		"yes",
		"target",
		"output",
		"cgo",
		"skip-web-generate",
		"webui-build-base-url",
		"webui-api-base-url",
		"webui-show-demo-todo",
	}
	if len(flags) != len(wantNames) {
		t.Fatalf("len(Flags()) = %d, want %d", len(flags), len(wantNames))
	}
	for i, want := range wantNames {
		if flags[i].Name != want {
			t.Fatalf("Flags()[%d].Name = %q, want %q", i, flags[i].Name, want)
		}
	}
	if flags[0].Default != false {
		t.Fatalf("yes default = %#v, want false", flags[0].Default)
	}
	if got := flags[1].Default; !reflect.DeepEqual(got, buildapp.DefaultTargetStrings()) {
		t.Fatalf("target default = %#v, want %#v", got, buildapp.DefaultTargetStrings())
	}
	if flags[2].Default != buildapp.DefaultOutputDir {
		t.Fatalf("output default = %#v, want %q", flags[2].Default, buildapp.DefaultOutputDir)
	}
	if flags[5].Default != buildapp.DefaultWebUIBuildBase {
		t.Fatalf("webui build base default = %#v, want %q", flags[5].Default, buildapp.DefaultWebUIBuildBase)
	}

	spec := cmd.Spec()
	if spec.HomeLabel != "打包 / build" || spec.HomeOrder != 40 || spec.HomeHidden {
		t.Fatalf("home metadata = label %q order %d hidden %v", spec.HomeLabel, spec.HomeOrder, spec.HomeHidden)
	}
}

func TestBuildCommandExecuteYesPassesOptionsToRunner(t *testing.T) {
	var got buildapp.Options
	var gotStdout, gotStderr io.Writer
	cmd := &BuildCommand{
		runner: func(_ context.Context, opts buildapp.Options, stdout, stderr io.Writer) error {
			got = opts
			gotStdout = stdout
			gotStderr = stderr
			return nil
		},
	}

	var stdout, stderr bytes.Buffer
	err := cmd.Execute(&cli.Context{
		Context: context.Background(),
		Flags: map[string]interface{}{
			"yes":                  true,
			"target":               []string{"linux/amd64", "windows/amd64"},
			"output":               "build/releases",
			"cgo":                  true,
			"skip-web-generate":    true,
			"webui-build-base-url": "/console/",
			"webui-api-base-url":   "/api",
			"webui-show-demo-todo": true,
		},
		ChangedFlags: map[string]bool{"yes": true},
		Stdout:       &stdout,
		Stderr:       &stderr,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := buildapp.Options{
		Targets:           []string{"linux/amd64", "windows/amd64"},
		OutputDir:         "build/releases",
		CGOEnabled:        true,
		SkipWebGenerate:   true,
		WebUIBuildBaseURL: "/console/",
		WebUIAPIBaseURL:   "/api",
		WebUIShowDemoTodo: true,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("options = %#v, want %#v", got, want)
	}
	if gotStdout != &stdout || gotStderr != &stderr {
		t.Fatalf("runner writers = %#v %#v, want command stdout/stderr", gotStdout, gotStderr)
	}
}

func TestBuildCommandExecuteExplicitFlagSkipsInteractivePrompt(t *testing.T) {
	called := false
	cmd := &BuildCommand{
		runner: func(_ context.Context, opts buildapp.Options, _ io.Writer, _ io.Writer) error {
			called = true
			if !reflect.DeepEqual(opts.Targets, []string{"linux/amd64"}) {
				t.Fatalf("Targets = %#v, want linux/amd64", opts.Targets)
			}
			return nil
		},
	}

	err := cmd.Execute(&cli.Context{
		Context: context.Background(),
		Flags: map[string]interface{}{
			"yes":                  false,
			"target":               []string{"linux/amd64"},
			"output":               buildapp.DefaultOutputDir,
			"cgo":                  false,
			"skip-web-generate":    true,
			"webui-build-base-url": buildapp.DefaultWebUIBuildBase,
			"webui-api-base-url":   buildapp.DefaultWebUIAPIBase,
			"webui-show-demo-todo": buildapp.DefaultWebUIShowDemo,
		},
		ChangedFlags: map[string]bool{"target": true, "skip-web-generate": true},
		Stdout:       io.Discard,
		Stderr:       io.Discard,
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("runner was not called")
	}
}

func TestBuildCommandExecuteWithoutFlagsUsesInteractivePrompt(t *testing.T) {
	var got buildapp.Options
	called := false
	var stdout bytes.Buffer
	cmd := &BuildCommand{
		runner: func(_ context.Context, opts buildapp.Options, _ io.Writer, _ io.Writer) error {
			called = true
			got = opts
			return nil
		},
	}

	err := cmd.Execute(&cli.Context{
		Context: context.Background(),
		Flags: map[string]interface{}{
			"yes":                  false,
			"target":               buildapp.DefaultTargetStrings(),
			"output":               buildapp.DefaultOutputDir,
			"cgo":                  false,
			"skip-web-generate":    false,
			"webui-build-base-url": buildapp.DefaultWebUIBuildBase,
			"webui-api-base-url":   buildapp.DefaultWebUIAPIBase,
			"webui-show-demo-todo": buildapp.DefaultWebUIShowDemo,
		},
		ChangedFlags: map[string]bool{},
		Stdin:        strings.NewReader("\n\n\n\ny\n"),
		Stdout:       &stdout,
		Stderr:       io.Discard,
		UI:           cli.NewPromptUI(strings.NewReader("\n\n\n\ny\n"), &stdout),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !called {
		t.Fatal("runner was not called")
	}
	if !reflect.DeepEqual(got.Targets, buildapp.DefaultTargetStrings()) || got.OutputDir != buildapp.DefaultOutputDir || got.SkipWebGenerate || got.CGOEnabled {
		t.Fatalf("options = %#v", got)
	}
	if !strings.Contains(stdout.String(), "开始构建") {
		t.Fatalf("stdout missing confirmation prompt:\n%s", stdout.String())
	}
}

func TestBuildCommandExecuteWithoutFlagsCanCancelInteractivePrompt(t *testing.T) {
	called := false
	var stdout bytes.Buffer
	cmd := &BuildCommand{
		runner: func(context.Context, buildapp.Options, io.Writer, io.Writer) error {
			called = true
			return nil
		},
	}

	err := cmd.Execute(&cli.Context{
		Context: context.Background(),
		Flags: map[string]interface{}{
			"yes":                  false,
			"target":               buildapp.DefaultTargetStrings(),
			"output":               buildapp.DefaultOutputDir,
			"cgo":                  false,
			"skip-web-generate":    false,
			"webui-build-base-url": buildapp.DefaultWebUIBuildBase,
			"webui-api-base-url":   buildapp.DefaultWebUIAPIBase,
			"webui-show-demo-todo": buildapp.DefaultWebUIShowDemo,
		},
		ChangedFlags: map[string]bool{},
		Stdout:       &stdout,
		Stderr:       io.Discard,
		UI:           cli.NewPromptUI(strings.NewReader("\n\n\n\n\n"), &stdout),
	})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if called {
		t.Fatal("runner was called after interactive cancellation")
	}
	if !strings.Contains(stdout.String(), "已取消") {
		t.Fatalf("stdout missing cancellation message:\n%s", stdout.String())
	}
}
