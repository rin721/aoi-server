package buildapp

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/types/constants"
)

const (
	DefaultOutputDir       = "build/releases"
	DefaultWebUIBuildBase  = "/admin/"
	DefaultWebUIAPIBase    = ""
	DefaultWebUIShowDemo   = false
	defaultBinaryName      = "go-scaffold-server"
	defaultWebUIDistDir    = "web/admin/.output/public"
	defaultWebUIPackageDir = "web/admin/.output/public"
)

var (
	defaultTargets = []Target{
		{GOOS: "linux", GOARCH: "amd64"},
		{GOOS: "windows", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "amd64"},
	}

	ErrWebUIDistMissing = errors.New("admin webui static dist missing")
)

type Options struct {
	Targets           []string
	OutputDir         string
	CGOEnabled        bool
	SkipWebGenerate   bool
	WebUIBuildBaseURL string
	WebUIAPIBaseURL   string
	WebUIShowDemoTodo bool
}

type Target struct {
	GOOS   string
	GOARCH string
}

type Artifact struct {
	Target      Target
	ArchivePath string
}

type CommandRequest struct {
	Name   string
	Args   []string
	Dir    string
	Env    []string
	Stdout io.Writer
	Stderr io.Writer
}

type CommandRunner interface {
	Run(context.Context, CommandRequest) error
}

type Builder struct {
	runner CommandRunner
	root   string
}

type BuilderOption func(*Builder)

func WithRunner(runner CommandRunner) BuilderOption {
	return func(builder *Builder) {
		builder.runner = runner
	}
}

func WithRoot(root string) BuilderOption {
	return func(builder *Builder) {
		builder.root = root
	}
}

func NewBuilder(opts ...BuilderOption) *Builder {
	builder := &Builder{
		runner: execRunner{},
		root:   ".",
	}
	for _, opt := range opts {
		opt(builder)
	}
	if builder.runner == nil {
		builder.runner = execRunner{}
	}
	if strings.TrimSpace(builder.root) == "" {
		builder.root = "."
	}
	return builder
}

func Build(ctx context.Context, opts Options, stdout, stderr io.Writer) error {
	_, err := NewBuilder().Build(ctx, opts, stdout, stderr)
	return err
}

func PromptOptions(ctx context.Context, ui cli.PromptUI, opts Options) (Options, bool, error) {
	if ui == nil {
		return Options{}, false, fmt.Errorf("interactive UI is not available")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	opts = applyDefaults(opts)
	currentTarget := Target{GOOS: runtime.GOOS, GOARCH: runtime.GOARCH}
	targetChoice, err := ui.Select(ctx, "选择构建目标", []cli.SelectOption{
		{Value: "default", Label: "默认三平台 amd64", Description: strings.Join(DefaultTargetStrings(), ", ")},
		{Value: "current", Label: "当前平台", Description: currentTarget.String()},
		{Value: "custom", Label: "自定义目标", Description: "例如 linux/amd64,windows/amd64"},
	})
	if err != nil {
		return Options{}, false, err
	}
	switch targetChoice {
	case "default":
		opts.Targets = DefaultTargetStrings()
	case "current":
		opts.Targets = []string{currentTarget.String()}
	case "custom":
		custom, err := ui.Input(ctx, "目标平台，逗号分隔", strings.Join(DefaultTargetStrings(), ","))
		if err != nil {
			return Options{}, false, err
		}
		if _, err := ParseTargets([]string{custom}); err != nil {
			return Options{}, false, err
		}
		opts.Targets = []string{custom}
	}

	output, err := ui.Input(ctx, "输出目录", opts.OutputDir)
	if err != nil {
		return Options{}, false, err
	}
	if strings.TrimSpace(output) != "" {
		opts.OutputDir = strings.TrimSpace(output)
	}

	generateWeb, err := ui.Confirm(ctx, "是否执行 pnpm generate 生成并打包 Admin WebUI？选择否时若无既有产物将构建后端-only 包。", true)
	if err != nil {
		return Options{}, false, err
	}
	opts.SkipWebGenerate = !generateWeb

	cgoEnabled, err := ui.Confirm(ctx, "是否启用 CGO？", false)
	if err != nil {
		return Options{}, false, err
	}
	opts.CGOEnabled = cgoEnabled
	if !opts.CGOEnabled {
		if err := ui.Info("CGO_ENABLED=0：SQLite 运行时不可用；请使用 MySQL/Postgres，或在兼容工具链上使用 --cgo 重建。"); err != nil {
			return Options{}, false, err
		}
	}

	proceed, err := ui.Confirm(ctx, "开始构建发布包？", false)
	if err != nil {
		return Options{}, false, err
	}
	if !proceed {
		if err := ui.Info("已取消构建。"); err != nil {
			return Options{}, false, err
		}
		return opts, false, nil
	}
	return opts, true, nil
}

func DefaultTargetStrings() []string {
	out := make([]string, 0, len(defaultTargets))
	for _, target := range defaultTargets {
		out = append(out, target.String())
	}
	return out
}

func ParseTargets(values []string) ([]Target, error) {
	if len(values) == 0 {
		return append([]Target(nil), defaultTargets...), nil
	}

	targets := make([]Target, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		for _, raw := range strings.Split(value, ",") {
			target, err := parseTarget(raw)
			if err != nil {
				return nil, err
			}
			key := target.String()
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			targets = append(targets, target)
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("at least one target is required")
	}
	return targets, nil
}

func (t Target) String() string {
	return t.GOOS + "/" + t.GOARCH
}

func (b *Builder) Build(ctx context.Context, opts Options, stdout, stderr io.Writer) ([]Artifact, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	opts = applyDefaults(opts)
	targets, err := ParseTargets(opts.Targets)
	if err != nil {
		return nil, err
	}

	root, err := filepath.Abs(b.root)
	if err != nil {
		return nil, err
	}
	outputDir := filepath.Join(root, filepath.Clean(opts.OutputDir))
	if filepath.IsAbs(opts.OutputDir) {
		outputDir = filepath.Clean(opts.OutputDir)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	includeWebUI, err := webUIDistExists(root)
	if err != nil {
		return nil, err
	}
	if !opts.SkipWebGenerate {
		fmt.Fprintln(stdout, "Generating Admin WebUI static files...")
		if err := b.generateWebUI(ctx, root, opts, stdout, stderr); err != nil {
			return nil, err
		}
		if err := requireWebUIDist(root); err != nil {
			return nil, err
		}
		includeWebUI = true
	} else if !includeWebUI {
		fmt.Fprintln(stdout, "Admin WebUI static dist not found; packaging backend-only release.")
	}

	if !opts.CGOEnabled {
		fmt.Fprintln(stdout, "CGO_ENABLED=0: SQLite is not available in these packages; use MySQL/Postgres or rebuild with --cgo.")
	}

	stagingRoot := filepath.Join(outputDir, ".staging")
	if err := os.RemoveAll(stagingRoot); err != nil {
		return nil, fmt.Errorf("clean staging dir: %w", err)
	}
	defer os.RemoveAll(stagingRoot)

	artifacts := make([]Artifact, 0, len(targets))
	for _, target := range targets {
		artifact, err := b.buildTarget(ctx, root, outputDir, stagingRoot, opts, target, includeWebUI, stdout, stderr)
		if err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
		fmt.Fprintf(stdout, "Built %s -> %s\n", target.String(), artifact.ArchivePath)
	}

	return artifacts, nil
}

func (b *Builder) generateWebUI(ctx context.Context, root string, opts Options, stdout, stderr io.Writer) error {
	return b.runner.Run(ctx, CommandRequest{
		Name: "pnpm",
		Args: []string{"generate"},
		Dir:  filepath.Join(root, "web", "admin"),
		Env: []string{
			"NUXT_APP_BASE_URL=" + opts.WebUIBuildBaseURL,
			"NUXT_PUBLIC_API_BASE_URL=" + opts.WebUIAPIBaseURL,
			"NUXT_PUBLIC_SHOW_DEMO_TODO=" + boolString(opts.WebUIShowDemoTodo),
		},
		Stdout: stdout,
		Stderr: stderr,
	})
}

func (b *Builder) buildTarget(ctx context.Context, root, outputDir, stagingRoot string, opts Options, target Target, includeWebUI bool, stdout, stderr io.Writer) (Artifact, error) {
	artifactBase := artifactBaseName(target)
	packageRoot := filepath.Join(stagingRoot, artifactBase)
	if err := os.MkdirAll(packageRoot, 0o755); err != nil {
		return Artifact{}, fmt.Errorf("create package root: %w", err)
	}

	binaryName := defaultBinaryName
	if target.GOOS == "windows" {
		binaryName += ".exe"
	}
	binaryOutput := filepath.Join(stagingRoot, artifactBase+"-"+binaryName)
	if err := b.runner.Run(ctx, CommandRequest{
		Name: "go",
		Args: []string{"build", "-mod=readonly", "-trimpath", "-ldflags=-s -w", "-o", binaryOutput, "./cmd/main"},
		Dir:  root,
		Env: []string{
			"GOOS=" + target.GOOS,
			"GOARCH=" + target.GOARCH,
			"CGO_ENABLED=" + cgoValue(opts.CGOEnabled),
		},
		Stdout: stdout,
		Stderr: stderr,
	}); err != nil {
		return Artifact{}, err
	}

	if err := copyFile(binaryOutput, filepath.Join(packageRoot, binaryName)); err != nil {
		return Artifact{}, err
	}
	if err := os.Chmod(filepath.Join(packageRoot, binaryName), 0o755); err != nil {
		return Artifact{}, fmt.Errorf("chmod binary: %w", err)
	}
	if err := populateRuntimePackage(root, packageRoot, opts, includeWebUI); err != nil {
		return Artifact{}, err
	}

	archivePath := filepath.Join(outputDir, artifactBase+archiveExt(target))
	if err := createArchive(stagingRoot, artifactBase, archivePath, target); err != nil {
		return Artifact{}, err
	}
	return Artifact{Target: target, ArchivePath: archivePath}, nil
}

func applyDefaults(opts Options) Options {
	if len(opts.Targets) == 0 {
		opts.Targets = DefaultTargetStrings()
	}
	if strings.TrimSpace(opts.OutputDir) == "" {
		opts.OutputDir = DefaultOutputDir
	}
	if strings.TrimSpace(opts.WebUIBuildBaseURL) == "" {
		opts.WebUIBuildBaseURL = DefaultWebUIBuildBase
	}
	return opts
}

func parseTarget(value string) (Target, error) {
	value = strings.TrimSpace(value)
	parts := strings.Split(value, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return Target{}, fmt.Errorf("target %q must use goos/goarch format", value)
	}
	for _, part := range parts {
		if !isTargetPart(part) {
			return Target{}, fmt.Errorf("target %q contains unsupported characters", value)
		}
	}
	return Target{GOOS: parts[0], GOARCH: parts[1]}, nil
}

func isTargetPart(value string) bool {
	for _, r := range value {
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if r == '_' {
			continue
		}
		return false
	}
	return true
}

func artifactBaseName(target Target) string {
	return fmt.Sprintf("%s_%s_%s_%s", constants.AppName, constants.AppVersion, target.GOOS, target.GOARCH)
}

func archiveExt(target Target) string {
	if target.GOOS == "windows" {
		return ".zip"
	}
	return ".tar.gz"
}

func cgoValue(enabled bool) string {
	if enabled {
		return "1"
	}
	return "0"
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func requireWebUIDist(root string) error {
	indexPath := filepath.Join(root, defaultWebUIDistDir, "index.html")
	info, err := os.Stat(indexPath)
	if err != nil || info.IsDir() {
		return fmt.Errorf("%w: %s", ErrWebUIDistMissing, indexPath)
	}
	return nil
}

func webUIDistExists(root string) (bool, error) {
	indexPath := filepath.Join(root, defaultWebUIDistDir, "index.html")
	info, err := os.Stat(indexPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat admin webui static dist: %w", err)
	}
	return !info.IsDir(), nil
}

func populateRuntimePackage(root, packageRoot string, opts Options, includeWebUI bool) error {
	operations := []struct {
		src string
		dst string
		dir bool
	}{
		{src: filepath.Join(root, "deploy", "config.production.example.yaml"), dst: filepath.Join(packageRoot, "configs", "config.yaml")},
		{src: filepath.Join(root, "configs", "config.example.yaml"), dst: filepath.Join(packageRoot, "configs", "config.example.yaml")},
		{src: filepath.Join(root, "configs", "locales"), dst: filepath.Join(packageRoot, "configs", "locales"), dir: true},
		{src: filepath.Join(root, "internal", "migrations"), dst: filepath.Join(packageRoot, "internal", "migrations"), dir: true},
		{src: filepath.Join(root, "plugins", "demo1", "plugin.yaml"), dst: filepath.Join(packageRoot, "plugins", "demo1", "plugin.yaml")},
	}
	if includeWebUI {
		operations = append(operations, struct {
			src string
			dst string
			dir bool
		}{src: filepath.Join(root, defaultWebUIDistDir), dst: filepath.Join(packageRoot, defaultWebUIPackageDir), dir: true})
	}
	for _, op := range operations {
		var err error
		if op.dir {
			err = copyDir(op.src, op.dst)
		} else {
			err = copyFile(op.src, op.dst)
		}
		if err != nil {
			return err
		}
	}
	for _, name := range []string{"data", "logs"} {
		if err := os.MkdirAll(filepath.Join(packageRoot, name), 0o755); err != nil {
			return fmt.Errorf("create runtime dir %s: %w", name, err)
		}
	}
	return writeReleaseReadme(packageRoot, opts, includeWebUI)
}

func writeReleaseReadme(packageRoot string, opts Options, includeWebUI bool) error {
	lines := []string{
		constants.AppName + " release package",
		"",
		"Run:",
		"  ./go-scaffold-server server --config=./configs/config.yaml",
		"",
		"Windows:",
		"  .\\go-scaffold-server.exe server --config=.\\configs\\config.yaml",
		"",
	}
	if includeWebUI {
		lines = append(lines, "This package includes Admin WebUI static files under web/admin/.output/public.")
	} else {
		lines = append(lines, "This package does not include Admin WebUI static files.")
	}
	lines = append(lines,
		"Default packages are built with CGO_ENABLED=0 unless --cgo was used.",
		"With CGO disabled, the SQLite driver is not available at runtime; use MySQL/Postgres or rebuild with --cgo on a compatible toolchain.",
		"Nuxt base URL: "+opts.WebUIBuildBaseURL,
	)
	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(filepath.Join(packageRoot, "README.txt"), []byte(content), 0o644)
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}
	if info.IsDir() {
		return fmt.Errorf("copy file %s: source is directory", src)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %s: %w", dst, err)
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open %s: %w", src, err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("create %s: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy %s to %s: %w", src, dst, err)
	}
	return nil
}

func createArchive(stagingRoot, artifactBase, archivePath string, target Target) error {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0o755); err != nil {
		return fmt.Errorf("create archive parent: %w", err)
	}
	if target.GOOS == "windows" {
		return createZipArchive(stagingRoot, artifactBase, archivePath)
	}
	return createTarGzArchive(stagingRoot, artifactBase, archivePath)
}

func createZipArchive(stagingRoot, artifactBase, archivePath string) error {
	out, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create zip archive: %w", err)
	}
	defer out.Close()

	zipWriter := zip.NewWriter(out)
	defer zipWriter.Close()

	entries, err := archiveEntries(filepath.Join(stagingRoot, artifactBase))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := addZipEntry(zipWriter, stagingRoot, entry); err != nil {
			return err
		}
	}
	return nil
}

func addZipEntry(zipWriter *zip.Writer, stagingRoot, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(stagingRoot, path)
	if err != nil {
		return err
	}
	name := filepath.ToSlash(rel)
	if info.IsDir() {
		name += "/"
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = name
	if !info.IsDir() {
		header.Method = zip.Deflate
	}
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	return writeFileToArchive(path, writer)
}

func createTarGzArchive(stagingRoot, artifactBase, archivePath string) error {
	out, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create tar.gz archive: %w", err)
	}
	defer out.Close()

	gzipWriter := gzip.NewWriter(out)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	entries, err := archiveEntries(filepath.Join(stagingRoot, artifactBase))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := addTarEntry(tarWriter, stagingRoot, entry); err != nil {
			return err
		}
	}
	return nil
}

func addTarEntry(tarWriter *tar.Writer, stagingRoot, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(stagingRoot, path)
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(rel)
	if info.IsDir() {
		header.Name += "/"
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	if info.IsDir() {
		return nil
	}
	return writeFileToArchive(path, tarWriter)
}

func archiveEntries(root string) ([]string, error) {
	entries := []string{}
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		entries = append(entries, path)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(entries)
	return entries, nil
}

func writeFileToArchive(path string, writer io.Writer) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(writer, in)
	return err
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, req CommandRequest) error {
	cmd := exec.CommandContext(ctx, req.Name, req.Args...)
	cmd.Dir = req.Dir
	cmd.Env = append(os.Environ(), req.Env...)
	cmd.Stdout = req.Stdout
	cmd.Stderr = req.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", req.Name, strings.Join(req.Args, " "), err)
	}
	return nil
}
