package main

import (
	"context"
	"io"
	"strings"

	"github.com/rei0721/go-scaffold/internal/app/buildapp"
	"github.com/rei0721/go-scaffold/pkg/cli"
)

type buildRunnerFunc func(context.Context, buildapp.Options, io.Writer, io.Writer) error

type BuildCommand struct {
	runner buildRunnerFunc
}

func NewBuildCommand() *BuildCommand {
	return &BuildCommand{}
}

func (c *BuildCommand) Name() string {
	return "build"
}

func (c *BuildCommand) Description() string {
	return "Build release packages for multiple platforms"
}

func (c *BuildCommand) Usage() string {
	return "build [flags]"
}

func (c *BuildCommand) Flags() []cli.FlagSpec {
	return []cli.FlagSpec{
		{
			Name:        "yes",
			Type:        cli.FlagTypeBool,
			Default:     false,
			Description: "Run build with defaults or explicit flags without interactive confirmation",
		},
		{
			Name:        "target",
			Type:        cli.FlagTypeStringSlice,
			Default:     buildapp.DefaultTargetStrings(),
			Description: "Target platform in goos/goarch format",
		},
		{
			Name:        "output",
			Type:        cli.FlagTypeString,
			Default:     buildapp.DefaultOutputDir,
			Description: "Release package output directory",
		},
		{
			Name:        "cgo",
			Type:        cli.FlagTypeBool,
			Default:     false,
			Description: "Build with CGO_ENABLED=1",
		},
		{
			Name:        "skip-web-generate",
			Type:        cli.FlagTypeBool,
			Default:     false,
			Description: "Skip pnpm generate; package existing Admin WebUI dist when present",
		},
		{
			Name:        "webui-build-base-url",
			Type:        cli.FlagTypeString,
			Default:     buildapp.DefaultWebUIBuildBase,
			Description: "Nuxt NUXT_APP_BASE_URL used by Admin WebUI generation",
		},
		{
			Name:        "webui-api-base-url",
			Type:        cli.FlagTypeString,
			Default:     buildapp.DefaultWebUIAPIBase,
			Description: "Nuxt NUXT_PUBLIC_API_BASE_URL used by Admin WebUI generation",
		},
		{
			Name:        "webui-show-demo-todo",
			Type:        cli.FlagTypeBool,
			Default:     buildapp.DefaultWebUIShowDemo,
			Description: "Nuxt NUXT_PUBLIC_SHOW_DEMO_TODO used by Admin WebUI generation",
		},
	}
}

func (c *BuildCommand) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:        c.Name(),
		Use:         c.Usage(),
		Description: c.Description(),
		HomeLabel:   "打包 / build",
		HomeOrder:   40,
		Flags:       c.Flags(),
		Run:         c.Execute,
	}
}

func (c *BuildCommand) Execute(ctx *cli.Context) error {
	runner := c.runner
	if runner == nil {
		runner = buildapp.Build
	}
	opts := buildOptionsFromContext(ctx)
	if !buildDirectRequested(ctx) {
		prompted, proceed, err := buildapp.PromptOptions(ctx.Context, ctx.UI, opts)
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}
		opts = prompted
	}
	return runner(ctx.Context, opts, ctx.Stdout, ctx.Stderr)
}

func buildOptionsFromContext(ctx *cli.Context) buildapp.Options {
	return buildapp.Options{
		Targets:           ctx.GetStringSlice("target"),
		OutputDir:         ctx.GetString("output"),
		CGOEnabled:        ctx.GetBool("cgo"),
		SkipWebGenerate:   ctx.GetBool("skip-web-generate"),
		WebUIBuildBaseURL: ctx.GetString("webui-build-base-url"),
		WebUIAPIBaseURL:   ctx.GetString("webui-api-base-url"),
		WebUIShowDemoTodo: ctx.GetBool("webui-show-demo-todo"),
	}
}

func buildDirectRequested(ctx *cli.Context) bool {
	if ctx == nil {
		return false
	}
	if ctx.GetBool("yes") {
		return true
	}
	for name, changed := range ctx.ChangedFlags {
		if changed && strings.TrimSpace(name) != "" {
			return true
		}
	}
	return false
}
