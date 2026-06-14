package initapp

import (
	"context"
	"fmt"
	"io"

	iammodel "github.com/rei0721/go-scaffold/internal/modules/iam/model"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	systemmodel "github.com/rei0721/go-scaffold/internal/modules/system/model"
	systemservice "github.com/rei0721/go-scaffold/internal/modules/system/service"
	"github.com/rei0721/go-scaffold/pkg/migrator"
)

// InitialSetupSource 标识初始化请求来源，用于审计和后续扩展来源差异。
type InitialSetupSource string

const (
	InitialSetupSourceCLI InitialSetupSource = "cli"
	InitialSetupSourceWeb InitialSetupSource = "web"
)

// InitialSetupInput 汇总 CLI init 和 WebUI 首次初始化共享的入参。
type InitialSetupInput struct {
	Source             InitialSetupSource
	OrgCode            string
	OrgName            string
	AdminUsername      string
	AdminEmail         string
	AdminDisplayName   string
	AdminPassword      string
	UserAgent          string
	IPAddress          string
	IssueLoginTokens   bool
	CreateServiceToken bool
	ServiceTokenDays   int
	ServiceTokenRemark string
}

// InitialSetupResult 描述一次显式初始化写入和同步后的结果。
type InitialSetupResult struct {
	Seed               systemservice.SeedResult
	APISync            systemmodel.APISyncResult
	PermissionSync     systemmodel.PermissionSyncResult
	Admin              *iamservice.Principal
	LoginTokens        iamservice.TokenPair
	LoginTokensIssued  bool
	ServiceToken       iamservice.CreateAPITokenResult
	ServiceTokenIssued bool
}

// InitialSetupRunner 编排首次初始化所需的迁移、系统同步和 IAM owner 创建。
type InitialSetupRunner struct {
	Core    Core
	Infra   Infrastructure
	Modules Modules
	Stdout  io.Writer
}

// NewInitialSetupRunner 创建显式初始化编排器。
func NewInitialSetupRunner(core Core, infra Infrastructure, modules Modules, stdout io.Writer) InitialSetupRunner {
	return InitialSetupRunner{Core: core, Infra: infra, Modules: modules, Stdout: stdout}
}

func attachWebInitialSetupService(core Core, infra Infrastructure, modules Modules) {
	if modules.IAM.Handler == nil || modules.IAM.Service == nil {
		return
	}
	modules.IAM.Handler.UseSetupService(webInitialSetupService{
		runner: NewInitialSetupRunner(core, infra, modules, nil),
	})
}

// Run 执行一次显式初始化。该方法只由 CLI init 或 WebUI setup 调用，不属于普通启动自动迁移。
func (r InitialSetupRunner) Run(ctx context.Context, input InitialSetupInput) (InitialSetupResult, error) {
	if r.Core.Config == nil {
		return InitialSetupResult{}, fmt.Errorf("initial setup: config is not initialized")
	}
	if r.Infra.Database == nil {
		return InitialSetupResult{}, fmt.Errorf("initial setup: database is not initialized")
	}

	r.printf("Applying database migrations...\n")
	if err := ApplyExplicitMigrations(ctx, r.Core, r.Infra); err != nil {
		return InitialSetupResult{}, fmt.Errorf("initial setup: apply migrations: %w", err)
	}

	if r.Core.Config.Demo.EnabledValue() && r.Core.Config.Demo.ApplySchemaOnStartValue() {
		r.printf("Applying demo schema...\n")
		if _, err := ApplyDemoSchemaForTrigger(r.Infra.Database, r.Core.Config.Database.Driver, r.Core.Logger, DemoSchemaTriggerServerStart); err != nil {
			return InitialSetupResult{}, fmt.Errorf("initial setup: apply demo schema: %w", err)
		}
	}

	var result InitialSetupResult
	if r.Modules.System.Service != nil {
		r.printf("Seeding system defaults...\n")
		seed, err := r.Modules.System.Service.SeedDefaults(ctx)
		if err != nil {
			return InitialSetupResult{}, fmt.Errorf("initial setup: seed system defaults: %w", err)
		}
		result.Seed = seed
		r.printf("System defaults: dictionaries=%d items=%d parameters=%d status=%s\n", seed.DictionariesCreated, seed.DictionaryItemsCreated, seed.ParametersCreated, seed.StorageStatus)

		apiSync, err := r.Modules.System.Service.SyncAPIs(ctx)
		if err != nil {
			return InitialSetupResult{}, fmt.Errorf("initial setup: sync api catalog: %w", err)
		}
		result.APISync = apiSync
		r.printf("API catalog sync: created=%d updated=%d stale=%d status=%s\n", apiSync.Created, apiSync.Updated, apiSync.Stale, apiSync.StorageStatus)

		permissionSync, err := r.Modules.System.Service.SyncPermissions(ctx)
		if err != nil {
			return InitialSetupResult{}, fmt.Errorf("initial setup: sync permissions: %w", err)
		}
		result.PermissionSync = permissionSync
		r.printf("Permission sync: created=%d skipped=%d status=%s\n", permissionSync.Created, permissionSync.Skipped, permissionSync.StorageStatus)
	}

	if r.Modules.IAM.Service == nil {
		r.printf("IAM is disabled; skipped admin and service token creation.\n")
		return result, nil
	}
	if input.AdminPassword == "" {
		if err := r.Modules.IAM.Service.LoadPolicies(ctx); err != nil {
			return InitialSetupResult{}, fmt.Errorf("initial setup: reload iam policies: %w", err)
		}
		r.printf("Admin password is empty; skipped IAM admin creation.\n")
		return result, nil
	}

	if input.IssueLoginTokens {
		pair, err := r.Modules.IAM.Service.InitialAdminSetup(ctx, iamservice.InitialAdminSetupInput{
			OrgCode:     input.OrgCode,
			OrgName:     input.OrgName,
			Username:    input.AdminUsername,
			Email:       input.AdminEmail,
			DisplayName: input.AdminDisplayName,
			Password:    input.AdminPassword,
			UserAgent:   input.UserAgent,
			IPAddress:   input.IPAddress,
		})
		if err != nil {
			return InitialSetupResult{}, err
		}
		result.LoginTokens = pair
		result.LoginTokensIssued = true
	} else {
		admin, err := r.Modules.IAM.Service.BootstrapAdmin(ctx, iamservice.BootstrapAdminInput{
			OrgCode:     input.OrgCode,
			OrgName:     input.OrgName,
			Username:    input.AdminUsername,
			Email:       input.AdminEmail,
			DisplayName: input.AdminDisplayName,
			Password:    input.AdminPassword,
		})
		if err != nil {
			return InitialSetupResult{}, err
		}
		result.Admin = admin
		r.printf("Admin ready: user=%s org=%d\n", admin.Username, admin.OrgID)

		if input.CreateServiceToken {
			created, err := r.Modules.IAM.Service.CreateAPIToken(ctx, iamservice.CreateAPITokenInput{
				Principal: *admin,
				UserID:    admin.UserID,
				RoleCode:  iammodel.RoleOwner,
				Days:      input.ServiceTokenDays,
				Remark:    input.ServiceTokenRemark,
			})
			if err != nil {
				return InitialSetupResult{}, err
			}
			result.ServiceToken = created
			result.ServiceTokenIssued = true
			r.printf("Service API token created, prefix=%s\n", created.Item.TokenPrefix)
			r.printf("Full token is shown only once: %s\n", created.Token)
		}
	}

	if err := r.Modules.IAM.Service.LoadPolicies(ctx); err != nil {
		return InitialSetupResult{}, fmt.Errorf("initial setup: reload iam policies: %w", err)
	}
	return result, nil
}

// ApplyExplicitMigrations 运行显式初始化迁移，不受 migration.auto_apply 启动期开关影响。
func ApplyExplicitMigrations(ctx context.Context, core Core, infra Infrastructure) error {
	return applyMigrations(ctx, core, infra, "initial-setup")
}

func applyMigrations(ctx context.Context, core Core, infra Infrastructure, trigger string) error {
	core.Config.Migration.ApplyDefaults()
	runner, err := migrator.New(infra.Database, migrator.Config{
		Driver: string(core.Config.Database.Driver),
		Dir:    core.Config.Migration.Dir,
	})
	if err != nil {
		return fmt.Errorf("initialize migrator: %w", err)
	}
	if err := runner.Up(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	if core.Logger != nil {
		core.Logger.Info("database migrations applied", "dir", core.Config.Migration.Dir, "trigger", trigger)
	}
	return nil
}

func (r InitialSetupRunner) printf(format string, args ...any) {
	if r.Stdout == nil {
		return
	}
	fmt.Fprintf(r.Stdout, format, args...)
}

type webInitialSetupService struct {
	runner InitialSetupRunner
}

func (s webInitialSetupService) SetupStatus(ctx context.Context) (iamservice.SetupStatus, error) {
	return s.runner.Modules.IAM.Service.SetupStatus(ctx)
}

func (s webInitialSetupService) InitialAdminSetup(ctx context.Context, input iamservice.InitialAdminSetupInput) (iamservice.TokenPair, error) {
	result, err := s.runner.Run(ctx, InitialSetupInput{
		Source:           InitialSetupSourceWeb,
		OrgCode:          input.OrgCode,
		OrgName:          input.OrgName,
		AdminUsername:    input.Username,
		AdminEmail:       input.Email,
		AdminDisplayName: input.DisplayName,
		AdminPassword:    input.Password,
		UserAgent:        input.UserAgent,
		IPAddress:        input.IPAddress,
		IssueLoginTokens: true,
	})
	if err != nil {
		return iamservice.TokenPair{}, err
	}
	return result.LoginTokens, nil
}
