package cliapp

import (
	"context"
	"io"
	"time"

	"github.com/rei0721/go-scaffold/internal/app/initapp"
	"github.com/rei0721/go-scaffold/internal/app/lifecycleapp"
	"github.com/rei0721/go-scaffold/types/constants"
)

// InitializationInput 描述一次交互式初始化需要的输入。
type InitializationInput struct {
	ConfigPath         string
	OrgCode            string
	OrgName            string
	AdminUsername      string
	AdminEmail         string
	AdminDisplayName   string
	AdminPassword      string
	CreateServiceToken bool
	ServiceTokenDays   int
	ServiceTokenRemark string
}

// ExecuteInitialization 执行数据库、IAM 和系统默认数据初始化。
func ExecuteInitialization(ctx context.Context, stdout io.Writer, input InitializationInput) error {
	if input.ConfigPath == "" {
		input.ConfigPath = constants.AppDefaultConfigPath
	}
	core, err := initapp.NewCore(input.ConfigPath)
	if err != nil {
		return err
	}
	infra, err := initapp.NewInfrastructure(core)
	if err != nil {
		return err
	}
	var transport initapp.Transport
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = lifecycleapp.Shutdown(shutdownCtx, core, infra, transport)
	}()

	var demo initapp.DemoModule
	if core.Config.Demo.EnabledValue() {
		demo = initapp.NewDemoModule(infra.Database, core.Logger)
	}

	var iam initapp.IAMModule
	if core.Config.Auth.Enabled {
		iam, err = initapp.NewIAMModule(core, infra)
		if err != nil {
			return err
		}
	}
	plugins, err := initapp.NewPluginsModule(core, iam)
	if err != nil {
		return err
	}
	system := initapp.NewSystemModule(core, infra, iam)
	modules := initapp.Modules{Demo: demo, IAM: iam, Plugins: plugins, System: system}
	transport, err = initapp.NewTransport(core, infra, modules)
	if err != nil {
		return err
	}

	runner := initapp.NewInitialSetupRunner(core, infra, modules, stdout)
	_, err = runner.Run(ctx, initapp.InitialSetupInput{
		Source:             initapp.InitialSetupSourceCLI,
		OrgCode:            input.OrgCode,
		OrgName:            input.OrgName,
		AdminUsername:      input.AdminUsername,
		AdminEmail:         input.AdminEmail,
		AdminDisplayName:   input.AdminDisplayName,
		AdminPassword:      input.AdminPassword,
		CreateServiceToken: input.CreateServiceToken,
		ServiceTokenDays:   input.ServiceTokenDays,
		ServiceTokenRemark: input.ServiceTokenRemark,
	})
	return err
}
