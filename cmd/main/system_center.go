package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/rei0721/go-scaffold/internal/app/cliapp"
	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/types/constants"
)

func NewSystemCenterCommands() []cli.CommandSpec {
	return []cli.CommandSpec{
		newRunCommandSpec(),
		newServiceCommandSpec(),
		newInitCommandSpec(),
	}
}

func newRunCommandSpec() cli.CommandSpec {
	configFlag := cli.FlagSpec{
		Name:        "config",
		ShortName:   "c",
		Type:        cli.FlagTypeString,
		Default:     constants.AppDefaultConfigPath,
		Description: "Config file path",
		EnvVar:      appconfig.EnvConfigPathName(),
	}
	return cli.CommandSpec{
		Name:        "run",
		Description: "Start managed background services",
		HomeLabel:   "启动 / run",
		HomeOrder:   10,
		Flags:       []cli.FlagSpec{configFlag},
		Run: func(ctx *cli.Context) error {
			return cliapp.RunStartFlow(ctx)
		},
		Commands: []cli.CommandSpec{
			{
				Name:        constants.AppServerCommandName,
				Use:         "server [--config=<path>]",
				Description: "Start server as a managed background process",
				Flags:       []cli.FlagSpec{configFlag},
				Run: func(ctx *cli.Context) error {
					state, err := cliapp.NewManager().StartServer(ctx.Context, ctx.GetString("config"))
					if err != nil {
						return err
					}
					cliapp.PrintServiceState(ctx.Stdout, state)
					return nil
				},
			},
		},
	}
}

func newServiceCommandSpec() cli.CommandSpec {
	linesFlag := cli.FlagSpec{Name: "lines", Type: cli.FlagTypeInt, Default: 100, Description: "History lines to print"}
	followFlag := cli.FlagSpec{Name: "follow", ShortName: "f", Type: cli.FlagTypeBool, Default: false, Description: "Follow appended logs"}
	return cli.CommandSpec{
		Name:        "service",
		Description: "Inspect and control managed services",
		HomeLabel:   "服务 / service",
		HomeOrder:   20,
		Run: func(ctx *cli.Context) error {
			return cliapp.RunServiceFlow(ctx)
		},
		Commands: []cli.CommandSpec{
			serviceStatusSpec(),
			serviceInfoSpec(),
			{
				Name:        "logs",
				Use:         "logs [server]",
				Description: "Show managed service logs",
				Flags:       []cli.FlagSpec{linesFlag, followFlag},
				Args:        validateOptionalServerArg,
				Run: func(ctx *cli.Context) error {
					state, err := cliapp.NewManager().Status(ctx.Context, cliapp.ServiceServer)
					if err != nil {
						return err
					}
					return cliapp.PrintServiceLogs(ctx.Context, ctx.Stdout, state, ctx.GetInt("lines"), ctx.GetBool("follow"))
				},
			},
			{
				Name:        "terminal",
				Use:         "terminal [server]",
				Description: "Attach to managed service log terminal",
				Flags:       []cli.FlagSpec{linesFlag},
				Args:        validateOptionalServerArg,
				Run: func(ctx *cli.Context) error {
					state, err := cliapp.NewManager().Status(ctx.Context, cliapp.ServiceServer)
					if err != nil {
						return err
					}
					return cliapp.PrintServiceLogs(ctx.Context, ctx.Stdout, state, ctx.GetInt("lines"), true)
				},
			},
			{
				Name:        "restart",
				Use:         "restart [server]",
				Description: "Restart managed service",
				Args:        validateOptionalServerArg,
				Run: func(ctx *cli.Context) error {
					state, err := cliapp.NewManager().RestartServer(ctx.Context)
					if err != nil {
						return err
					}
					cliapp.PrintServiceState(ctx.Stdout, state)
					return nil
				},
			},
			{
				Name:        "stop",
				Use:         "stop [server]",
				Description: "Stop managed service",
				Args:        validateOptionalServerArg,
				Run: func(ctx *cli.Context) error {
					state, err := cliapp.NewManager().StopServer(ctx.Context)
					if err != nil {
						return err
					}
					cliapp.PrintServiceState(ctx.Stdout, state)
					return nil
				},
			},
		},
	}
}

func serviceStatusSpec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:        "status",
		Use:         "status [server]",
		Description: "Show managed service status",
		Args:        validateOptionalServerArg,
		Run: func(ctx *cli.Context) error {
			state, err := cliapp.NewManager().Status(ctx.Context, cliapp.ServiceServer)
			if err != nil {
				return err
			}
			cliapp.PrintServiceState(ctx.Stdout, state)
			return nil
		},
	}
}

func serviceInfoSpec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:        "info",
		Use:         "info [server]",
		Description: "Show managed service metadata",
		Args:        validateOptionalServerArg,
		Run: func(ctx *cli.Context) error {
			state, err := cliapp.NewManager().Status(ctx.Context, cliapp.ServiceServer)
			if err != nil {
				return err
			}
			return json.NewEncoder(ctx.Stdout).Encode(state)
		},
	}
}

func newInitCommandSpec() cli.CommandSpec {
	configFlag := cli.FlagSpec{
		Name:        "config",
		ShortName:   "c",
		Type:        cli.FlagTypeString,
		Default:     constants.AppDefaultConfigPath,
		Description: "Config file path",
		EnvVar:      appconfig.EnvConfigPathName(),
	}
	return cli.CommandSpec{
		Name:        "init",
		Use:         "init [flags]",
		Description: "Initialize database, IAM admin, service account, and system defaults",
		HomeLabel:   "初始化 / init",
		HomeOrder:   30,
		Flags: []cli.FlagSpec{
			configFlag,
			{Name: "org-code", Type: cli.FlagTypeString, Default: "acme", Description: "IAM organization code for optional admin bootstrap"},
			{Name: "org-name", Type: cli.FlagTypeString, Default: "acme", Description: "IAM organization name for optional admin bootstrap"},
			{Name: "admin-username", Type: cli.FlagTypeString, Default: "admin", Description: "IAM admin username"},
			{Name: "admin-email", Type: cli.FlagTypeString, Default: "admin@example.com", Description: "IAM admin email"},
			{Name: "admin-display-name", Type: cli.FlagTypeString, Default: "admin", Description: "IAM admin display name"},
			{Name: "admin-password", Type: cli.FlagTypeString, Description: "IAM admin password; omit to skip admin bootstrap"},
			{Name: "admin-password-stdin", Type: cli.FlagTypeBool, Description: "Read IAM admin password from stdin"},
			{Name: "create-service-token", Type: cli.FlagTypeBool, Description: "Create API Token for the initialized admin"},
			{Name: "service-token-days", Type: cli.FlagTypeInt, Default: 30, Description: "API Token validity days; use -1 for no expiration"},
			{Name: "service-token-remark", Type: cli.FlagTypeString, Default: "created by cli init", Description: "API Token remark"},
		},
		Run: func(ctx *cli.Context) error {
			input, err := initializationInputFromContext(ctx)
			if err != nil {
				return err
			}
			return cliapp.RunInitializationFlow(ctx, input)
		},
	}
}

func initializationInputFromContext(ctx *cli.Context) (cliapp.InitializationInput, error) {
	password := ctx.GetString("admin-password")
	if ctx.GetBool("admin-password-stdin") {
		raw, err := io.ReadAll(ctx.Stdin)
		if err != nil {
			return cliapp.InitializationInput{}, err
		}
		password = strings.TrimSpace(string(raw))
	}
	return cliapp.InitializationInput{
		ConfigPath:         ctx.GetString("config"),
		OrgCode:            ctx.GetString("org-code"),
		OrgName:            ctx.GetString("org-name"),
		AdminUsername:      ctx.GetString("admin-username"),
		AdminEmail:         ctx.GetString("admin-email"),
		AdminDisplayName:   ctx.GetString("admin-display-name"),
		AdminPassword:      password,
		CreateServiceToken: ctx.GetBool("create-service-token"),
		ServiceTokenDays:   ctx.GetInt("service-token-days"),
		ServiceTokenRemark: ctx.GetString("service-token-remark"),
	}, nil
}

func validateOptionalServerArg(ctx *cli.Context) error {
	if len(ctx.Args) == 0 {
		return nil
	}
	if len(ctx.Args) == 1 && ctx.Args[0] == cliapp.ServiceServer {
		return nil
	}
	return &cli.UsageError{Command: ctx.CommandPath, Message: fmt.Sprintf("expected optional service name %q", cliapp.ServiceServer)}
}
