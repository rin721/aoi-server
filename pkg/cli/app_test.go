package cli

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestRunRoutesCobraCommandAndParsesFlags(t *testing.T) {
	t.Setenv("CLI_TEST_OUTPUT", "from-env")

	app, err := NewApp(Config{Name: "tool", Version: "1.2.3", Description: "test cli"})
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	var got *Context
	err = app.AddCommand(CommandSpec{
		Name:        "run",
		Description: "run command",
		Flags: []FlagSpec{
			{Name: "name", Type: FlagTypeString, Required: true},
			{Name: "count", Type: FlagTypeInt, Default: 1},
			{Name: "verbose", Type: FlagTypeBool},
			{Name: "tags", Type: FlagTypeStringSlice},
			{Name: "output", Type: FlagTypeString, EnvVar: "CLI_TEST_OUTPUT"},
		},
		Run: func(ctx *Context) error {
			got = ctx
			return nil
		},
	})
	if err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}

	var stdout, stderr bytes.Buffer
	err = app.RunWithIO(context.Background(), []string{
		"run",
		"--name", "alice",
		"--count", "3",
		"--verbose",
		"--tags", "alpha,beta",
		"positional",
	}, strings.NewReader("input"), &stdout, &stderr)
	if err != nil {
		t.Fatalf("RunWithIO() error = %v", err)
	}

	if got == nil {
		t.Fatal("command was not executed")
	}
	if got.CommandName != "run" {
		t.Fatalf("CommandName = %q, want run", got.CommandName)
	}
	if got.GetString("name") != "alice" {
		t.Fatalf("name = %q, want alice", got.GetString("name"))
	}
	if got.GetInt("count") != 3 {
		t.Fatalf("count = %d, want 3", got.GetInt("count"))
	}
	if !got.GetBool("verbose") {
		t.Fatal("verbose = false, want true")
	}
	if want := []string{"alpha", "beta"}; !reflect.DeepEqual(got.GetStringSlice("tags"), want) {
		t.Fatalf("tags = %#v, want %#v", got.GetStringSlice("tags"), want)
	}
	if got.GetString("output") != "from-env" {
		t.Fatalf("output = %q, want from-env", got.GetString("output"))
	}
	if want := []string{"positional"}; !reflect.DeepEqual(got.Args, want) {
		t.Fatalf("Args = %#v, want %#v", got.Args, want)
	}
}

func TestRunMapsUsageAndExecutionErrors(t *testing.T) {
	app, err := NewApp(Config{Name: "tool"})
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}

	cause := errors.New("boom")
	if err := app.AddCommand(CommandSpec{
		Name:  "run",
		Flags: []FlagSpec{{Name: "name", Type: FlagTypeString, Required: true}},
		Run: func(*Context) error {
			return cause
		},
	}); err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}

	err = app.RunWithIO(context.Background(), []string{"run"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	var usageErr *UsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("missing required error = %T, want *UsageError", err)
	}
	if got := GetExitCode(err); got != ExitUsage {
		t.Fatalf("usage exit = %d, want %d", got, ExitUsage)
	}

	err = app.RunWithIO(context.Background(), []string{"run", "--name", "alice"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	var commandErr *CommandError
	if !errors.As(err, &commandErr) {
		t.Fatalf("execution error = %T, want *CommandError", err)
	}
	if !errors.Is(err, cause) {
		t.Fatal("execution error does not wrap cause")
	}
	if got := GetExitCode(err); got != ExitError {
		t.Fatalf("command exit = %d, want %d", got, ExitError)
	}
}

func TestRunMapsInvalidEnvDefaultToUsageError(t *testing.T) {
	t.Setenv("CLI_TEST_COUNT", "not-an-int")

	app, err := NewApp(Config{Name: "tool"})
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	if err := app.AddCommand(CommandSpec{
		Name:  "run",
		Flags: []FlagSpec{{Name: "count", Type: FlagTypeInt, EnvVar: "CLI_TEST_COUNT"}},
		Run:   func(*Context) error { return nil },
	}); err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}

	err = app.RunWithIO(context.Background(), []string{"run"}, nil, &bytes.Buffer{}, &bytes.Buffer{})
	var usageErr *UsageError
	if !errors.As(err, &usageErr) {
		t.Fatalf("invalid env error = %T, want *UsageError", err)
	}
	if got := GetExitCode(err); got != ExitUsage {
		t.Fatalf("exit = %d, want %d", got, ExitUsage)
	}
}

func TestAddCommandRejectsDuplicateNames(t *testing.T) {
	app, err := NewApp(Config{Name: "tool"})
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	if err := app.AddCommand(CommandSpec{Name: "run"}); err != nil {
		t.Fatalf("first AddCommand() error = %v", err)
	}

	err = app.AddCommand(CommandSpec{Name: "run"})
	if err == nil {
		t.Fatal("second AddCommand() error = nil, want duplicate error")
	}
	if !strings.Contains(err.Error(), ErrMsgDuplicateCommand) {
		t.Fatalf("duplicate error = %q, want duplicate message", err.Error())
	}
}

func TestRunHelpAndVersionUseCobra(t *testing.T) {
	app, err := NewApp(Config{Name: "tool", Version: "1.2.3", Description: "test cli"})
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	if err := app.AddCommand(CommandSpec{Name: "run", Description: "run command"}); err != nil {
		t.Fatalf("AddCommand() error = %v", err)
	}

	var help bytes.Buffer
	if err := app.RunWithIO(context.Background(), []string{"--help"}, nil, &help, &bytes.Buffer{}); err != nil {
		t.Fatalf("RunWithIO(--help) error = %v", err)
	}
	helpText := help.String()
	for _, want := range []string{"test cli", "run command", "Usage:"} {
		if !strings.Contains(helpText, want) {
			t.Fatalf("help output %q does not contain %q", helpText, want)
		}
	}

	var version bytes.Buffer
	if err := app.RunWithIO(context.Background(), []string{"--version"}, nil, &version, &bytes.Buffer{}); err != nil {
		t.Fatalf("RunWithIO(--version) error = %v", err)
	}
	if got := strings.TrimSpace(version.String()); got != "tool version 1.2.3" {
		t.Fatalf("version output = %q, want %q", got, "tool version 1.2.3")
	}
}

func TestRunWithoutArgsStartsInteractiveHome(t *testing.T) {
	impl, err := NewApp(Config{Name: "tool", Version: "1.2.3", Description: "test cli"})
	if err != nil {
		t.Fatalf("NewApp() error = %v", err)
	}
	appImpl := impl.(*app)

	var called bool
	appImpl.runHome = func(_ context.Context, model homeModel, _ streams, _ []ProgramOption) error {
		called = true
		view := model.View()
		if !view.AltScreen {
			t.Fatal("home view AltScreen = false, want true")
		}
		if !strings.Contains(view.Content, "tool v1.2.3") {
			t.Fatalf("home content %q does not contain title", view.Content)
		}
		return &CancelledError{}
	}

	err = appImpl.RunWithIO(context.Background(), nil, nil, &bytes.Buffer{}, &bytes.Buffer{})
	var cancelled *CancelledError
	if !errors.As(err, &cancelled) {
		t.Fatalf("RunWithIO(nil) error = %T, want *CancelledError", err)
	}
	if !called {
		t.Fatal("interactive home runner was not called")
	}
}

func TestHomeModelNavigationHelpAndQuit(t *testing.T) {
	model := newHomeModel(homeConfig{
		Name:        "tool",
		Version:     "1.2.3",
		Description: "test cli",
		Theme:       DefaultTheme(),
		Commands: []homeCommand{
			{Name: "server", Description: "run server", Help: "server help"},
			{Name: "db", Description: "database tools", Help: "db help"},
		},
	})

	if view := model.View(); !strings.Contains(view.Content, "server") || !view.AltScreen {
		t.Fatalf("initial view = %#v", view)
	}

	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	model = updated.(homeModel)
	if model.width != 100 || model.height != 40 {
		t.Fatalf("size = %dx%d, want 100x40", model.width, model.height)
	}

	updated, _ = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	model = updated.(homeModel)
	if model.selected != 1 {
		t.Fatalf("selected = %d, want 1", model.selected)
	}

	updated, _ = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	model = updated.(homeModel)
	if !model.showingHelp || !strings.Contains(model.View().Content, "db help") {
		t.Fatalf("help state = %v, content = %q", model.showingHelp, model.View().Content)
	}

	updated, cmd := model.Update(tea.KeyPressMsg(tea.Key{Text: "q", Code: 'q'}))
	model = updated.(homeModel)
	if !model.cancelled {
		t.Fatal("cancelled = false, want true")
	}
	if cmd == nil {
		t.Fatal("quit command = nil")
	}
}

func TestDefaultHomeRunnerMapsBubbleTeaQuitToCancelled(t *testing.T) {
	model := newHomeModel(homeConfig{
		Name:  "tool",
		Theme: DefaultTheme(),
	})

	var stdout, stderr bytes.Buffer
	err := defaultHomeRunner(
		context.Background(),
		model,
		streams{
			stdin:  strings.NewReader("q"),
			stdout: &stdout,
			stderr: &stderr,
		},
		[]ProgramOption{tea.WithoutRenderer(), tea.WithoutSignals()},
	)
	var cancelled *CancelledError
	if !errors.As(err, &cancelled) {
		t.Fatalf("defaultHomeRunner() error = %T, want *CancelledError", err)
	}
}
