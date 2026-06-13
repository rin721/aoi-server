package buildapp

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/rei0721/go-scaffold/pkg/cli"
)

func TestParseTargetsDefaultsAndRejectsInvalid(t *testing.T) {
	targets, err := ParseTargets(nil)
	if err != nil {
		t.Fatalf("ParseTargets(nil) error = %v", err)
	}
	want := []Target{
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "windows", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "amd64"},
	}
	if !reflect.DeepEqual(targets, want) {
		t.Fatalf("targets = %#v, want %#v", targets, want)
	}

	targets, err = ParseTargets([]string{"linux/amd64,windows/amd64", "linux/amd64"})
	if err != nil {
		t.Fatalf("ParseTargets(comma) error = %v", err)
	}
	if got := targetStrings(targets); !reflect.DeepEqual(got, []string{"linux/amd64", "windows/amd64"}) {
		t.Fatalf("target strings = %#v", got)
	}

	if _, err := ParseTargets([]string{"linux"}); err == nil {
		t.Fatal("ParseTargets(linux) error = nil, want error")
	}
	if _, err := ParseTargets([]string{"linux/amd-64"}); err == nil {
		t.Fatal("ParseTargets(linux/amd-64) error = nil, want error")
	}
}

func TestBuilderRunsWebGenerateAndGoBuild(t *testing.T) {
	root := newBuildSource(t, false)
	runner := &recordingRunner{t: t, root: root, generateWeb: true}
	builder := NewBuilder(WithRoot(root), WithRunner(runner))

	var stdout strings.Builder
	artifacts, err := builder.Build(context.Background(), Options{
		Targets:           []string{"linux/amd64"},
		OutputDir:         "out",
		WebUIBuildBaseURL: "/console/",
		WebUIAPIBaseURL:   "/api",
		WebUIShowDemoTodo: true,
	}, &stdout, io.Discard)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(artifacts) != 1 || !strings.HasSuffix(artifacts[0].ArchivePath, ".tar.gz") {
		t.Fatalf("artifacts = %#v", artifacts)
	}
	if len(runner.requests) != 2 {
		t.Fatalf("len(requests) = %d, want 2: %#v", len(runner.requests), runner.requests)
	}

	webReq := runner.requests[0]
	if webReq.Name != "pnpm" || !reflect.DeepEqual(webReq.Args, []string{"generate"}) {
		t.Fatalf("web request = %#v", webReq)
	}
	if webReq.Dir != filepath.Join(root, "web", "admin") {
		t.Fatalf("web dir = %q", webReq.Dir)
	}
	for _, want := range []string{
		"NUXT_APP_BASE_URL=/console/",
		"NUXT_PUBLIC_API_BASE_URL=/api",
		"NUXT_PUBLIC_SHOW_DEMO_TODO=true",
	} {
		if !containsString(webReq.Env, want) {
			t.Fatalf("web env missing %q: %#v", want, webReq.Env)
		}
	}

	goReq := runner.requests[1]
	if goReq.Name != "go" {
		t.Fatalf("go request name = %q", goReq.Name)
	}
	for _, want := range []string{"GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0"} {
		if !containsString(goReq.Env, want) {
			t.Fatalf("go env missing %q: %#v", want, goReq.Env)
		}
	}
	for _, want := range []string{"build", "-mod=readonly", "-trimpath", "-ldflags=-s -w", "./cmd/main"} {
		if !containsString(goReq.Args, want) {
			t.Fatalf("go args missing %q: %#v", want, goReq.Args)
		}
	}
	if !strings.Contains(stdout.String(), "SQLite is not available") {
		t.Fatalf("stdout missing CGO warning:\n%s", stdout.String())
	}
}

func TestBuilderUsesCGOWhenRequested(t *testing.T) {
	root := newBuildSource(t, true)
	runner := &recordingRunner{t: t, root: root}
	builder := NewBuilder(WithRoot(root), WithRunner(runner))

	var stdout strings.Builder
	if _, err := builder.Build(context.Background(), Options{
		Targets:         []string{"linux/amd64"},
		OutputDir:       "out",
		CGOEnabled:      true,
		SkipWebGenerate: true,
	}, &stdout, io.Discard); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(runner.requests) != 1 {
		t.Fatalf("len(requests) = %d, want 1", len(runner.requests))
	}
	if !containsString(runner.requests[0].Env, "CGO_ENABLED=1") {
		t.Fatalf("go env missing CGO_ENABLED=1: %#v", runner.requests[0].Env)
	}
	if strings.Contains(stdout.String(), "SQLite is not available") {
		t.Fatalf("stdout should not include CGO warning when --cgo is set:\n%s", stdout.String())
	}
}

func TestBuilderSkipWebGenerateRequiresExistingDist(t *testing.T) {
	root := newBuildSource(t, false)
	builder := NewBuilder(WithRoot(root), WithRunner(&recordingRunner{t: t, root: root}))

	_, err := builder.Build(context.Background(), Options{
		Targets:         []string{"linux/amd64"},
		OutputDir:       "out",
		SkipWebGenerate: true,
	}, io.Discard, io.Discard)
	if !errors.Is(err, ErrWebUIDistMissing) {
		t.Fatalf("Build() error = %v, want ErrWebUIDistMissing", err)
	}
}

func TestBuilderArchivesRuntimeLayout(t *testing.T) {
	root := newBuildSource(t, true)
	runner := &recordingRunner{t: t, root: root}
	builder := NewBuilder(WithRoot(root), WithRunner(runner))

	artifacts, err := builder.Build(context.Background(), Options{
		Targets:         []string{"windows/amd64", "linux/amd64"},
		OutputDir:       "out",
		SkipWebGenerate: true,
	}, io.Discard, io.Discard)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("len(artifacts) = %d, want 2", len(artifacts))
	}

	windowsEntries := readZipEntries(t, artifacts[0].ArchivePath)
	windowsPrefix := artifactBaseName(Target{GOOS: "windows", GOARCH: "amd64"}) + "/"
	for _, want := range []string{
		windowsPrefix + "go-scaffold-server.exe",
		windowsPrefix + "configs/config.yaml",
		windowsPrefix + "configs/config.example.yaml",
		windowsPrefix + "configs/locales/zh-CN.yaml",
		windowsPrefix + "internal/migrations/20260614000100_demo.sql",
		windowsPrefix + "plugins/demo1/plugin.yaml",
		windowsPrefix + "web/admin/.output/public/index.html",
		windowsPrefix + "data/",
		windowsPrefix + "logs/",
		windowsPrefix + "README.txt",
	} {
		if !containsString(windowsEntries, want) {
			t.Fatalf("zip missing %q:\n%v", want, windowsEntries)
		}
	}

	linuxEntries := readTarGzEntries(t, artifacts[1].ArchivePath)
	linuxPrefix := artifactBaseName(Target{GOOS: "linux", GOARCH: "amd64"}) + "/"
	for _, want := range []string{
		linuxPrefix + "go-scaffold-server",
		linuxPrefix + "configs/config.yaml",
		linuxPrefix + "web/admin/.output/public/index.html",
		linuxPrefix + "data/",
		linuxPrefix + "logs/",
	} {
		if !containsString(linuxEntries, want) {
			t.Fatalf("tar.gz missing %q:\n%v", want, linuxEntries)
		}
	}
}

func TestPromptOptionsDefaultTargetsProceed(t *testing.T) {
	ui := &fakePromptUI{
		selects:  []string{"default"},
		inputs:   []string{""},
		confirms: []bool{true, false, true},
	}

	opts, proceed, err := PromptOptions(context.Background(), ui, Options{})
	if err != nil {
		t.Fatalf("PromptOptions() error = %v", err)
	}
	if !proceed {
		t.Fatal("proceed = false, want true")
	}
	if !reflect.DeepEqual(opts.Targets, DefaultTargetStrings()) {
		t.Fatalf("Targets = %#v, want %#v", opts.Targets, DefaultTargetStrings())
	}
	if opts.OutputDir != DefaultOutputDir || opts.SkipWebGenerate || opts.CGOEnabled {
		t.Fatalf("unexpected options: %#v", opts)
	}
	if len(ui.infos) != 1 || !strings.Contains(ui.infos[0], "SQLite") {
		t.Fatalf("infos = %#v, want SQLite warning", ui.infos)
	}
}

func TestPromptOptionsCurrentTarget(t *testing.T) {
	ui := &fakePromptUI{
		selects:  []string{"current"},
		inputs:   []string{"dist"},
		confirms: []bool{false, true, true},
	}

	opts, proceed, err := PromptOptions(context.Background(), ui, Options{})
	if err != nil {
		t.Fatalf("PromptOptions() error = %v", err)
	}
	if !proceed {
		t.Fatal("proceed = false, want true")
	}
	wantTargets := []string{runtime.GOOS + "/" + runtime.GOARCH}
	if !reflect.DeepEqual(opts.Targets, wantTargets) {
		t.Fatalf("Targets = %#v, want %#v", opts.Targets, wantTargets)
	}
	if opts.OutputDir != "dist" || !opts.SkipWebGenerate || !opts.CGOEnabled {
		t.Fatalf("unexpected options: %#v", opts)
	}
	if len(ui.infos) != 0 {
		t.Fatalf("infos = %#v, want none when CGO is enabled", ui.infos)
	}
}

func TestPromptOptionsCustomTargetsCancel(t *testing.T) {
	ui := &fakePromptUI{
		selects:  []string{"custom"},
		inputs:   []string{"linux/amd64,windows/amd64", "out"},
		confirms: []bool{true, false, false},
	}

	opts, proceed, err := PromptOptions(context.Background(), ui, Options{})
	if err != nil {
		t.Fatalf("PromptOptions() error = %v", err)
	}
	if proceed {
		t.Fatal("proceed = true, want false")
	}
	if !reflect.DeepEqual(opts.Targets, []string{"linux/amd64,windows/amd64"}) || opts.OutputDir != "out" {
		t.Fatalf("unexpected options: %#v", opts)
	}
	if len(ui.infos) != 2 || !strings.Contains(ui.infos[1], "取消") {
		t.Fatalf("infos = %#v, want warning and cancel message", ui.infos)
	}
}

func TestPromptOptionsRejectsInvalidCustomTarget(t *testing.T) {
	ui := &fakePromptUI{
		selects: []string{"custom"},
		inputs:  []string{"linux"},
	}

	if _, _, err := PromptOptions(context.Background(), ui, Options{}); err == nil {
		t.Fatal("PromptOptions() error = nil, want invalid target error")
	}
}

type recordingRunner struct {
	t           *testing.T
	root        string
	generateWeb bool
	requests    []CommandRequest
}

type fakePromptUI struct {
	selects  []string
	inputs   []string
	confirms []bool
	infos    []string
}

func (ui *fakePromptUI) Select(_ context.Context, _ string, _ []cli.SelectOption) (string, error) {
	if len(ui.selects) == 0 {
		return "", errors.New("missing select answer")
	}
	answer := ui.selects[0]
	ui.selects = ui.selects[1:]
	return answer, nil
}

func (ui *fakePromptUI) Confirm(_ context.Context, _ string, _ bool) (bool, error) {
	if len(ui.confirms) == 0 {
		return false, errors.New("missing confirm answer")
	}
	answer := ui.confirms[0]
	ui.confirms = ui.confirms[1:]
	return answer, nil
}

func (ui *fakePromptUI) Input(_ context.Context, _ string, _ string) (string, error) {
	if len(ui.inputs) == 0 {
		return "", errors.New("missing input answer")
	}
	answer := ui.inputs[0]
	ui.inputs = ui.inputs[1:]
	return answer, nil
}

func (ui *fakePromptUI) Password(context.Context, string) (string, error) {
	return "", errors.New("password prompt is not supported")
}

func (ui *fakePromptUI) Info(message string) error {
	ui.infos = append(ui.infos, message)
	return nil
}

func (r *recordingRunner) Run(_ context.Context, req CommandRequest) error {
	copied := req
	copied.Args = append([]string(nil), req.Args...)
	copied.Env = append([]string(nil), req.Env...)
	r.requests = append(r.requests, copied)

	switch req.Name {
	case "pnpm":
		if r.generateWeb {
			writeTestFile(r.t, filepath.Join(req.Dir, ".output", "public", "index.html"), "<html>admin</html>")
		}
	case "go":
		out := outputArg(r.t, req.Args)
		writeTestFile(r.t, out, "binary")
	}
	return nil
}

func newBuildSource(t *testing.T, withWebDist bool) string {
	t.Helper()
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "deploy", "config.production.example.yaml"), "server:\n  mode: release\n")
	writeTestFile(t, filepath.Join(root, "configs", "config.example.yaml"), "server:\n  mode: debug\n")
	writeTestFile(t, filepath.Join(root, "configs", "locales", "zh-CN.yaml"), "hello: nihao\n")
	writeTestFile(t, filepath.Join(root, "internal", "migrations", "20260614000100_demo.sql"), "-- migrate\n")
	writeTestFile(t, filepath.Join(root, "plugins", "demo1", "plugin.yaml"), "id: demo1\n")
	if withWebDist {
		writeTestFile(t, filepath.Join(root, "web", "admin", ".output", "public", "index.html"), "<html>admin</html>")
	}
	return root
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}

func outputArg(t *testing.T, args []string) string {
	t.Helper()
	for i, arg := range args {
		if arg == "-o" && i+1 < len(args) {
			return args[i+1]
		}
	}
	t.Fatalf("args missing -o: %#v", args)
	return ""
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func targetStrings(targets []Target) []string {
	out := make([]string, 0, len(targets))
	for _, target := range targets {
		out = append(out, target.String())
	}
	return out
}

func readZipEntries(t *testing.T, path string) []string {
	t.Helper()
	reader, err := zip.OpenReader(path)
	if err != nil {
		t.Fatalf("OpenReader(%s) error = %v", path, err)
	}
	defer reader.Close()

	entries := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		entries = append(entries, file.Name)
	}
	return entries
}

func readTarGzEntries(t *testing.T, path string) []string {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open(%s) error = %v", path, err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("NewReader(%s) error = %v", path, err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	var entries []string
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("tar Next(%s) error = %v", path, err)
		}
		entries = append(entries, header.Name)
	}
	return entries
}
