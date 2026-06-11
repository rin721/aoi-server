package initapp

// 本文件属于应用初始化装配层，负责把配置、基础设施、业务模块或传输层拼接为可运行的分层对象。

import (
	"context"
	"fmt"
	"time"

	"github.com/rei0721/go-scaffold/internal/app/dbapp"
	demohandler "github.com/rei0721/go-scaffold/internal/modules/demo/handler"
	demorepository "github.com/rei0721/go-scaffold/internal/modules/demo/repository"
	demoservice "github.com/rei0721/go-scaffold/internal/modules/demo/service"
	iamhandler "github.com/rei0721/go-scaffold/internal/modules/iam/handler"
	iamrepository "github.com/rei0721/go-scaffold/internal/modules/iam/repository"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/pkg/authorization"
	"github.com/rei0721/go-scaffold/pkg/crypto"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/migrator"
	"github.com/rei0721/go-scaffold/pkg/token"
)

// NewModules 根据配置装配业务模块。
//
// Demo 模块是脚手架示例能力；只有启用 Demo 时才会创建 Todo 仓储、服务和 handler。
// 启动期是否应用 Demo schema 由 demo.apply_schema_on_start 控制。
func NewModules(core Core, infra Infrastructure) (Modules, error) {
	if err := ApplyConfiguredMigrations(core, infra); err != nil {
		return Modules{}, err
	}

	var demoModule DemoModule
	if core.Config.Demo.EnabledValue() {
		if core.Config.Demo.ApplySchemaOnStartValue() {
			if _, err := ApplyDemoSchemaForTrigger(infra.Database, core.Config.Database.Driver, core.Logger, DemoSchemaTriggerServerStart); err != nil {
				return Modules{}, err
			}
		}
		demoModule = NewDemoModule(infra.Database, core.Logger)
	} else if core.Logger != nil {
		core.Logger.Info("demo module disabled")
	}

	var iamModule IAMModule
	if core.Config.Auth.Enabled {
		module, err := NewIAMModule(core, infra)
		if err != nil {
			return Modules{}, err
		}
		iamModule = module
	} else if core.Logger != nil {
		core.Logger.Info("iam module disabled")
	}
	return Modules{
		Demo: demoModule,
		IAM:  iamModule,
	}, nil
}

// DemoSchemaTrigger 表示触发 Demo 表结构应用策略的运行场景。
type DemoSchemaTrigger string

const (
	// DemoSchemaTriggerServerStart 表示服务启动期，可按配置应用 Demo schema。
	DemoSchemaTriggerServerStart DemoSchemaTrigger = "server-start"
	// DemoSchemaTriggerReload 表示配置热更新期，不允许隐式修改表结构。
	DemoSchemaTriggerReload DemoSchemaTrigger = "reload"
)

// DemoSchemaPolicy 描述某个触发场景是否允许应用 Demo schema 以及原因。
type DemoSchemaPolicy struct {
	Trigger DemoSchemaTrigger
	Apply   bool
	Reason  string
}

// DemoSchemaPolicyFor 返回 Demo schema 的场景策略。
//
// reload 明确跳过 schema 变更，避免配置热更新时产生不可预期的数据结构副作用。
func DemoSchemaPolicyFor(trigger DemoSchemaTrigger) DemoSchemaPolicy {
	switch trigger {
	case DemoSchemaTriggerServerStart:
		return DemoSchemaPolicy{
			Trigger: trigger,
			Apply:   true,
			Reason:  "demo server startup keeps the local development schema ready through sqlgen",
		}
	case DemoSchemaTriggerReload:
		return DemoSchemaPolicy{
			Trigger: trigger,
			Apply:   false,
			Reason:  "database reload must not perform implicit schema changes",
		}
	default:
		return DemoSchemaPolicy{
			Trigger: trigger,
			Apply:   false,
			Reason:  "unknown schema trigger requires an explicit policy",
		}
	}
}

// ApplyDemoSchema 使用默认 server-start 策略应用 Demo Todo 表结构。
func ApplyDemoSchema(db database.Database, driver string, log logger.Logger) error {
	_, err := ApplyDemoSchemaForTrigger(db, driver, log, DemoSchemaTriggerServerStart)
	return err
}

// ApplyDemoSchemaForTrigger 按触发策略应用 Demo Todo 表结构。
//
// 返回策略用于测试和审计；当策略不允许应用或数据库为空时不会执行 SQL。
func ApplyDemoSchemaForTrigger(db database.Database, driver string, log logger.Logger, trigger DemoSchemaTrigger) (DemoSchemaPolicy, error) {
	policy := DemoSchemaPolicyFor(trigger)
	if !policy.Apply {
		logDemoSchemaSkipped(log, policy)
		return policy, nil
	}
	if db == nil {
		return policy, nil
	}
	if _, err := dbapp.ApplyDemoSchema(context.Background(), db, driver); err != nil {
		return policy, fmt.Errorf("apply demo schema: %w", err)
	}
	logDemoSchemaApplied(log, policy)
	return policy, nil
}

// logDemoSchemaApplied 记录 Demo schema 已执行的上下文，帮助区分 server 启动和命令行触发。
func logDemoSchemaApplied(log logger.Logger, policy DemoSchemaPolicy) {
	if log == nil {
		return
	}
	log.Info("demo schema applied", "trigger", policy.Trigger, "reason", policy.Reason)
}

// logDemoSchemaSkipped 记录 Demo schema 被策略跳过的原因，避免静默忽略配置意图。
func logDemoSchemaSkipped(log logger.Logger, policy DemoSchemaPolicy) {
	if log == nil {
		return
	}
	log.Debug("demo schema apply skipped", "trigger", policy.Trigger, "reason", policy.Reason)
}

// NewDemoModule 装配 Demo Todo 的仓储、服务和 HTTP handler。
func NewDemoModule(db database.Database, log logger.Logger) DemoModule {
	todoRepo := demorepository.NewTodoRepository(db)
	todoService := demoservice.NewTodoService(db, todoRepo)
	todoHandler := demohandler.NewTodoHandler(todoService, log)

	return DemoModule{
		TodoRepository: todoRepo,
		TodoService:    todoService,
		TodoHandler:    todoHandler,
	}
}

func ApplyConfiguredMigrations(core Core, infra Infrastructure) error {
	core.Config.Migration.ApplyDefaults()
	if !core.Config.Migration.AutoApply {
		return nil
	}
	runner, err := migrator.New(infra.Database, migrator.Config{
		Driver: string(core.Config.Database.Driver),
		Dir:    core.Config.Migration.Dir,
	})
	if err != nil {
		return fmt.Errorf("initialize migrator: %w", err)
	}
	if err := runner.Up(context.Background()); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	if core.Logger != nil {
		core.Logger.Info("database migrations applied", "dir", core.Config.Migration.Dir)
	}
	return nil
}

func NewIAMModule(core Core, infra Infrastructure) (IAMModule, error) {
	authCfg := core.Config.Auth
	authCfg.ApplyDefaults()

	passwords, err := crypto.NewBcrypt()
	if err != nil {
		return IAMModule{}, fmt.Errorf("initialize password crypto: %w", err)
	}
	tokenManager, err := token.New(token.Config{
		Issuer:        authCfg.Issuer,
		Audience:      authCfg.Audience,
		SigningKey:    authCfg.SigningKey,
		AccessTTL:     time.Duration(authCfg.AccessTokenTTLSeconds) * time.Second,
		RefreshTTL:    time.Duration(authCfg.RefreshTokenTTLSeconds) * time.Second,
		RefreshPepper: authCfg.RefreshTokenPepper,
	})
	if err != nil {
		return IAMModule{}, fmt.Errorf("initialize token manager: %w", err)
	}
	enforcer, err := authorization.New()
	if err != nil {
		return IAMModule{}, fmt.Errorf("initialize authorization enforcer: %w", err)
	}
	repo := iamrepository.New(infra.Database)
	service := iamservice.New(infra.Database, repo, passwords, tokenManager, enforcer, core.IDGenerator, iamservice.Config{
		MFAIssuer:         authCfg.MFAIssuer,
		MFASecretKey:      authCfg.MFASecretKey,
		LoginMaxFailures:  authCfg.LoginMaxFailures,
		LoginLockDuration: time.Duration(authCfg.LoginLockMinutes) * time.Minute,
		InvitationTTL:     time.Duration(authCfg.InvitationTTLSeconds) * time.Second,
		PasswordResetTTL:  time.Duration(authCfg.PasswordResetTTLSeconds) * time.Second,
	}, iamservice.NoopNotifier{})
	if err := service.LoadPolicies(context.Background()); err != nil && core.Logger != nil {
		core.Logger.Warn("failed to load iam policies", "error", err)
	}
	return IAMModule{
		Repository: repo,
		Service:    service,
		Handler:    iamhandler.New(service, core.Logger),
	}, nil
}
