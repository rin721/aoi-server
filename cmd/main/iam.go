package main

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/rei0721/go-scaffold/internal/app/initapp"
	appconfig "github.com/rei0721/go-scaffold/internal/config"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/types/constants"
)

type IAMCommand struct{}

func NewIAMCommand() *IAMCommand {
	return &IAMCommand{}
}

func (c *IAMCommand) Spec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:        "iam",
		Use:         "iam",
		Description: "Manage IAM users, organizations, roles, and bootstrap tasks",
		Commands: []cli.CommandSpec{
			c.BootstrapAdminSpec(),
		},
	}
}

func (c *IAMCommand) BootstrapAdminSpec() cli.CommandSpec {
	return cli.CommandSpec{
		Name:        "bootstrap-admin",
		Use:         "bootstrap-admin [flags]",
		Description: "Create the initial IAM organization owner",
		Flags: []cli.FlagSpec{
			{Name: "config", Type: cli.FlagTypeString, Default: constants.AppDefaultConfigPath, Description: "Config file path", EnvVar: appconfig.EnvConfigPathName()},
			{Name: "org-code", Type: cli.FlagTypeString, Required: true, Description: "Organization code"},
			{Name: "org-name", Type: cli.FlagTypeString, Description: "Organization name"},
			{Name: "username", Type: cli.FlagTypeString, Required: true, Description: "Admin username"},
			{Name: "email", Type: cli.FlagTypeString, Required: true, Description: "Admin email"},
			{Name: "display-name", Type: cli.FlagTypeString, Description: "Admin display name"},
			{Name: "password", Type: cli.FlagTypeString, Description: "Admin password; prefer --password-stdin in automation"},
			{Name: "password-stdin", Type: cli.FlagTypeBool, Description: "Read admin password from stdin"},
		},
		Run: c.ExecuteBootstrapAdmin,
	}
}

func (c *IAMCommand) ExecuteBootstrapAdmin(ctx *cli.Context) error {
	password := ctx.GetString("password")
	if ctx.GetBool("password-stdin") {
		raw, err := io.ReadAll(ctx.Stdin)
		if err != nil {
			return err
		}
		password = strings.TrimSpace(string(raw))
	}
	if password == "" {
		return &cli.UsageError{Command: ctx.CommandPath, Message: "password is required; pass --password or --password-stdin"}
	}

	core, err := initapp.NewCore(ctx.GetString("config"))
	if err != nil {
		return fmt.Errorf("initialize core: %w", err)
	}
	defer func() {
		if core.Logger != nil {
			_ = core.Logger.Sync()
		}
	}()
	infra, err := initapp.NewInfrastructure(core)
	if err != nil {
		return err
	}
	defer func() {
		if infra.Database != nil {
			_ = infra.Database.Close()
		}
	}()
	if err := initapp.ApplyConfiguredMigrations(core, infra); err != nil {
		return err
	}
	module, err := initapp.NewIAMModule(core, infra)
	if err != nil {
		return err
	}
	principal, err := module.Service.BootstrapAdmin(ctx.Context, iamservice.BootstrapAdminInput{
		OrgCode:     ctx.GetString("org-code"),
		OrgName:     ctx.GetString("org-name"),
		Username:    ctx.GetString("username"),
		Email:       ctx.GetString("email"),
		DisplayName: ctx.GetString("display-name"),
		Password:    password,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(ctx.Stdout).Encode(principal)
}
