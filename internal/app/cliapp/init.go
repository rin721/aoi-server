package cliapp

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rei0721/go-scaffold/internal/app/initapp"
	"github.com/rei0721/go-scaffold/internal/app/lifecycleapp"
	iammodel "github.com/rei0721/go-scaffold/internal/modules/iam/model"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/pkg/migrator"
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
		return fmt.Errorf("initialize core: %w", err)
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

	fmt.Fprintln(stdout, "应用数据库迁移...")
	core.Config.Migration.ApplyDefaults()
	runner, err := migrator.New(infra.Database, migrator.Config{
		Driver: string(core.Config.Database.Driver),
		Dir:    core.Config.Migration.Dir,
	})
	if err != nil {
		return err
	}
	if err := runner.Up(ctx); err != nil {
		return err
	}

	var demo initapp.DemoModule
	if core.Config.Demo.EnabledValue() {
		if core.Config.Demo.ApplySchemaOnStartValue() {
			fmt.Fprintln(stdout, "应用 Demo schema...")
			if _, err := initapp.ApplyDemoSchemaForTrigger(infra.Database, core.Config.Database.Driver, core.Logger, initapp.DemoSchemaTriggerServerStart); err != nil {
				return err
			}
		}
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

	fmt.Fprintln(stdout, "补齐系统默认数据...")
	seed, err := system.Service.SeedDefaults(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "系统默认数据：dictionaries=%d items=%d parameters=%d status=%s\n", seed.DictionariesCreated, seed.DictionaryItemsCreated, seed.ParametersCreated, seed.StorageStatus)

	apiSync, err := system.Service.SyncAPIs(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "API 目录同步：created=%d updated=%d stale=%d status=%s\n", apiSync.Created, apiSync.Updated, apiSync.Stale, apiSync.StorageStatus)

	permissionSync, err := system.Service.SyncPermissions(ctx)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "权限同步：created=%d skipped=%d status=%s\n", permissionSync.Created, permissionSync.Skipped, permissionSync.StorageStatus)

	if iam.Service == nil {
		fmt.Fprintln(stdout, "IAM 未启用，跳过管理员和服务账户创建。")
		return nil
	}
	if input.AdminPassword == "" {
		fmt.Fprintln(stdout, "未输入管理员密码，跳过 IAM 管理员创建。")
		return nil
	}
	admin, err := iam.Service.BootstrapAdmin(ctx, iamservice.BootstrapAdminInput{
		OrgCode:     input.OrgCode,
		OrgName:     input.OrgName,
		Username:    input.AdminUsername,
		Email:       input.AdminEmail,
		DisplayName: input.AdminDisplayName,
		Password:    input.AdminPassword,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "管理员已就绪：user=%s org=%d\n", admin.Username, admin.OrgID)

	if input.CreateServiceToken {
		created, err := iam.Service.CreateAPIToken(ctx, iamservice.CreateAPITokenInput{
			Principal: *admin,
			UserID:    admin.UserID,
			RoleCode:  iammodel.RoleOwner,
			Days:      input.ServiceTokenDays,
			Remark:    input.ServiceTokenRemark,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "服务账户 API Token 已创建，prefix=%s\n", created.Item.TokenPrefix)
		fmt.Fprintf(stdout, "完整 token 仅显示一次：%s\n", created.Token)
	}
	return nil
}
