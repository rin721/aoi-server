package initapp

// 本文件属于应用初始化装配层，负责把配置、基础设施、业务模块或传输层拼接为可运行的分层对象。

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/internal/app/dbapp"
	"github.com/rei0721/go-scaffold/internal/config"
	demohandler "github.com/rei0721/go-scaffold/internal/modules/demo/handler"
	demorepository "github.com/rei0721/go-scaffold/internal/modules/demo/repository"
	demoservice "github.com/rei0721/go-scaffold/internal/modules/demo/service"
	iamhandler "github.com/rei0721/go-scaffold/internal/modules/iam/handler"
	iammodel "github.com/rei0721/go-scaffold/internal/modules/iam/model"
	iamrepository "github.com/rei0721/go-scaffold/internal/modules/iam/repository"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	pluginhandler "github.com/rei0721/go-scaffold/internal/modules/plugins/handler"
	pluginservice "github.com/rei0721/go-scaffold/internal/modules/plugins/service"
	systemhandler "github.com/rei0721/go-scaffold/internal/modules/system/handler"
	systemmodel "github.com/rei0721/go-scaffold/internal/modules/system/model"
	systemrepository "github.com/rei0721/go-scaffold/internal/modules/system/repository"
	systemservice "github.com/rei0721/go-scaffold/internal/modules/system/service"
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

	pluginsModule, err := NewPluginsModule(core, iamModule)
	if err != nil {
		return Modules{}, err
	}
	systemModule := NewSystemModule(core, infra, iamModule)
	return Modules{
		Demo:    demoModule,
		IAM:     iamModule,
		Plugins: pluginsModule,
		System:  systemModule,
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
	notifier, err := NewIAMNotifier(authCfg)
	if err != nil {
		return IAMModule{}, err
	}
	service := iamservice.New(infra.Database, repo, passwords, tokenManager, enforcer, core.IDGenerator, iamservice.Config{
		SelfSignupEnabled:  authCfg.SelfSignupEnabled,
		MFAIssuer:          authCfg.MFAIssuer,
		MFASecretKey:       authCfg.MFASecretKey,
		LoginMaxFailures:   authCfg.LoginMaxFailures,
		LoginLockDuration:  time.Duration(authCfg.LoginLockMinutes) * time.Minute,
		CaptchaEnabled:     authCfg.LoginCaptchaEnabled,
		CaptchaTTL:         time.Duration(authCfg.CaptchaTTLSeconds) * time.Second,
		InvitationTTL:      time.Duration(authCfg.InvitationTTLSeconds) * time.Second,
		PasswordResetTTL:   time.Duration(authCfg.PasswordResetTTLSeconds) * time.Second,
		NotificationDriver: authCfg.NotificationDriver,
		PublicBaseURL:      webUIPublicBaseURL(core.Config.WebUI),
		PasswordPolicy: iamservice.PasswordPolicy{
			MinLength:     authCfg.PasswordPolicy.MinLength,
			RequireLower:  authCfg.PasswordPolicy.RequireLower,
			RequireUpper:  authCfg.PasswordPolicy.RequireUpper,
			RequireNumber: authCfg.PasswordPolicy.RequireNumber,
			RequireSymbol: authCfg.PasswordPolicy.RequireSymbol,
		},
	}, notifier)
	if err := service.LoadPolicies(context.Background()); err != nil && core.Logger != nil {
		core.Logger.Warn("failed to load iam policies", "error", err)
	}
	return IAMModule{
		Repository: repo,
		Service:    service,
		Handler:    iamhandler.New(service, core.Logger),
	}, nil
}

func NewIAMNotifier(cfg config.AuthConfig) (iamservice.Notifier, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.NotificationDriver)) {
	case "smtp":
		return iamservice.NewSMTPNotifier(iamservice.SMTPNotifierConfig{
			Host:     cfg.SMTP.Host,
			Port:     cfg.SMTP.Port,
			Username: cfg.SMTP.Username,
			Password: cfg.SMTP.Password,
			From:     cfg.SMTP.From,
			FromName: cfg.SMTP.FromName,
			StartTLS: cfg.SMTP.StartTLS,
		})
	default:
		return iamservice.NoopNotifier{}, nil
	}
}

func NewPluginsModule(core Core, iam IAMModule) (PluginsModule, error) {
	pluginService, err := pluginservice.New(PluginsServiceConfig(core.Config.Plugins), core.Logger)
	if err != nil {
		return PluginsModule{}, err
	}
	if !core.Config.Plugins.Enabled {
		if core.Logger != nil {
			core.Logger.Info("plugins module disabled")
		}
		return PluginsModule{Service: pluginService}, nil
	}
	handler := pluginhandler.New(pluginService, iam.Service, iam.Service, core.Logger)
	return PluginsModule{
		Service: pluginService,
		Handler: handler,
	}, nil
}

func NewSystemModule(core Core, infra Infrastructure, iam IAMModule) SystemModule {
	options := []systemservice.Option{systemservice.WithIDGenerator(core.IDGenerator)}
	if infra.Database != nil {
		options = append(options, systemservice.WithRepository(systemrepository.New(infra.Database)))
	}
	if infra.Storage != nil {
		options = append(options, systemservice.WithStorage(infra.Storage))
	}
	if iam.Repository != nil {
		options = append(options, systemservice.WithPermissionStore(newSystemPermissionStore(iam.Repository, core.IDGenerator)))
	}
	service := systemservice.New(systemservice.Config{
		DemoEnabled:   core.Config.Demo.EnabledValue(),
		RuntimeConfig: SystemConfigSnapshot(core.Config),
		StartTime:     time.Now().UTC(),
	}, options...)
	seedSystemDefaults(core, service)
	return SystemModule{
		Service: service,
		Handler: systemhandler.New(service, iam.Service, core.Logger),
	}
}

func seedSystemDefaults(core Core, service systemservice.Service) {
	if core.Config == nil || service == nil || !core.Config.System.SeedDefaultsOnStartValue() {
		return
	}
	result, err := service.SeedDefaults(context.Background())
	if err != nil {
		if core.Logger != nil {
			core.Logger.Warn("system defaults seed failed", "error", err)
		}
		return
	}
	if core.Logger != nil {
		core.Logger.Info(
			"system defaults seed completed",
			"storage", result.StorageStatus,
			"dictionaries", result.DictionariesCreated,
			"dictionary_items", result.DictionaryItemsCreated,
			"parameters", result.ParametersCreated,
		)
	}
}

type systemPermissionIDGenerator interface {
	NextID() int64
}

type systemPermissionStore struct {
	repo iamrepository.Repository
	ids  systemPermissionIDGenerator
}

func newSystemPermissionStore(repo iamrepository.Repository, ids systemPermissionIDGenerator) systemservice.PermissionStore {
	return &systemPermissionStore{repo: repo, ids: ids}
}

func (s *systemPermissionStore) ListPermissions(ctx context.Context) ([]systemmodel.PermissionEntry, error) {
	permissions, err := s.repo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]systemmodel.PermissionEntry, 0, len(permissions))
	for _, permission := range permissions {
		out = append(out, systemmodel.PermissionEntry{
			Code:        permission.Code,
			Name:        permission.Name,
			Description: permission.Description,
		})
	}
	return out, nil
}

func (s *systemPermissionStore) CreatePermission(ctx context.Context, permission systemmodel.PermissionEntry) error {
	now := time.Now().UTC()
	return s.repo.CreatePermission(ctx, &iammodel.Permission{
		ID:          s.ids.NextID(),
		Code:        permission.Code,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func PluginsServiceConfig(cfg config.PluginsConfig) pluginservice.Config {
	cfg.ApplyDefaults()
	out := pluginservice.Config{
		Enabled:       cfg.Enabled,
		ManifestPaths: append([]string(nil), cfg.Manifests...),
		HealthTimeout: time.Duration(cfg.HealthTimeoutSeconds) * time.Second,
		ProxyTimeout:  time.Duration(cfg.ProxyTimeoutSeconds) * time.Second,
		Inline:        make([]pluginservice.Manifest, 0, len(cfg.Items)),
	}
	for _, item := range cfg.Items {
		out.Inline = append(out.Inline, pluginservice.Manifest{
			ID:         item.ID,
			Name:       item.Name,
			Version:    item.Version,
			BaseURL:    item.BaseURL,
			HealthPath: item.HealthPath,
			Frontend: pluginservice.Frontend{
				Entry: item.Frontend.Entry,
			},
			Menus:       pluginMenusConfig(item.Menus),
			Permissions: pluginPermissionsConfig(item.Permissions),
			Proxy: pluginservice.Proxy{
				Prefixes: append([]string(nil), item.Proxy.Prefixes...),
			},
			SecretRef: item.SecretRef,
		})
	}
	return out
}

func pluginMenusConfig(items []config.PluginMenuConfig) []pluginservice.Menu {
	menus := make([]pluginservice.Menu, 0, len(items))
	for _, item := range items {
		menus = append(menus, pluginservice.Menu{
			Code:       item.Code,
			Label:      item.Label,
			Icon:       item.Icon,
			Path:       item.Path,
			Permission: item.Permission,
			Order:      item.Order,
		})
	}
	return menus
}

func pluginPermissionsConfig(items []config.PluginPermissionConfig) []pluginservice.Permission {
	permissions := make([]pluginservice.Permission, 0, len(items))
	for _, item := range items {
		permissions = append(permissions, pluginservice.Permission{
			Code:        item.Code,
			Name:        item.Name,
			Description: item.Description,
		})
	}
	return permissions
}

func webUIPublicBaseURL(cfg config.WebUIConfig) string {
	cfg.ApplyDefaults()
	if cfg.PublicBaseURL != "" {
		return cfg.PublicBaseURL
	}
	return cfg.MountPath
}

func SystemConfigSnapshot(configSnapshot *config.Config) systemmodel.ConfigSnapshot {
	if configSnapshot == nil {
		return systemmodel.ConfigSnapshot{}
	}
	cfg := *configSnapshot
	cfg.Auth.ApplyDefaults()
	cfg.Migration.ApplyDefaults()
	cfg.Plugins.ApplyDefaults()
	cfg.RPC.ApplyDefaults()
	cfg.Storage.DefaultConfig()
	cfg.WebUI.ApplyDefaults()

	return systemmodel.ConfigSnapshot{Sections: []systemmodel.ConfigSection{
		{
			Code:        "server",
			Description: "HTTP 服务监听、运行模式和连接超时。",
			Icon:        "server",
			Label:       "系统服务",
			Order:       10,
			Items: []systemmodel.ConfigItem{
				configItem("server.host", "监听地址", cfg.Server.Host),
				configItem("server.port", "监听端口", cfg.Server.Port),
				configItem("server.mode", "运行模式", cfg.Server.Mode),
				configItem("server.read_timeout", "读取超时(秒)", cfg.Server.ReadTimeout),
				configItem("server.write_timeout", "写入超时(秒)", cfg.Server.WriteTimeout),
				configItem("server.idle_timeout", "空闲超时(秒)", cfg.Server.IdleTimeout),
			},
		},
		{
			Code:        "database",
			Description: "主数据库驱动、连接地址和连接池参数。",
			Icon:        "database",
			Label:       "数据库",
			Order:       20,
			Items: []systemmodel.ConfigItem{
				configItem("database.driver", "驱动", cfg.Database.Driver),
				configItem("database.host", "主机", cfg.Database.Host),
				configItem("database.port", "端口", cfg.Database.Port),
				configItem("database.user", "用户", cfg.Database.User),
				secretConfigItem("database.password", "密码", cfg.Database.Password),
				configItem("database.dbname", "数据库", cfg.Database.DBName),
				configItem("database.max_open_conns", "最大打开连接", cfg.Database.MaxOpenConns),
				configItem("database.max_idle_conns", "最大空闲连接", cfg.Database.MaxIdleConns),
			},
		},
		{
			Code:        "redis",
			Description: "Redis 缓存连接、重试和读写超时。",
			Icon:        "hard-drive",
			Label:       "Redis",
			Order:       30,
			Items: []systemmodel.ConfigItem{
				configItem("redis.enabled", "启用", cfg.Redis.Enabled),
				configItem("redis.host", "主机", cfg.Redis.Host),
				configItem("redis.port", "端口", cfg.Redis.Port),
				secretConfigItem("redis.password", "密码", cfg.Redis.Password),
				configItem("redis.db", "数据库", cfg.Redis.DB),
				configItem("redis.pool_size", "连接池", cfg.Redis.PoolSize),
				configItem("redis.min_idle_conns", "最小空闲连接", cfg.Redis.MinIdleConns),
				configItem("redis.max_retries", "最大重试", cfg.Redis.MaxRetries),
				configItem("redis.dial_timeout", "连接超时(秒)", cfg.Redis.DialTimeout),
				configItem("redis.read_timeout", "读取超时(秒)", cfg.Redis.ReadTimeout),
				configItem("redis.write_timeout", "写入超时(秒)", cfg.Redis.WriteTimeout),
			},
		},
		{
			Code:        "auth",
			Description: "IAM、令牌、MFA、登录锁定和通知策略。",
			Icon:        "shield-check",
			Label:       "认证安全",
			Order:       40,
			Items: []systemmodel.ConfigItem{
				configItem("auth.enabled", "启用 IAM", cfg.Auth.Enabled),
				configItem("auth.self_signup_enabled", "开放注册", cfg.Auth.SelfSignupEnabled),
				configItem("auth.issuer", "签发者", cfg.Auth.Issuer),
				configItem("auth.audience", "受众", cfg.Auth.Audience),
				secretConfigItem("auth.signing_key", "签名密钥", cfg.Auth.SigningKey),
				configItem("auth.access_token_ttl_seconds", "Access TTL(秒)", cfg.Auth.AccessTokenTTLSeconds),
				configItem("auth.refresh_token_ttl_seconds", "Refresh TTL(秒)", cfg.Auth.RefreshTokenTTLSeconds),
				secretConfigItem("auth.refresh_token_pepper", "Refresh Pepper", cfg.Auth.RefreshTokenPepper),
				configItem("auth.mfa_issuer", "MFA 签发者", cfg.Auth.MFAIssuer),
				secretConfigItem("auth.mfa_secret_key", "MFA 密钥", cfg.Auth.MFASecretKey),
				configItem("auth.login_max_failures", "登录失败锁定次数", cfg.Auth.LoginMaxFailures),
				configItem("auth.login_lock_minutes", "锁定时长(分钟)", cfg.Auth.LoginLockMinutes),
				configItem("auth.login_captcha_enabled", "登录验证码", cfg.Auth.LoginCaptchaEnabled),
				configItem("auth.captcha_ttl_seconds", "验证码 TTL(秒)", cfg.Auth.CaptchaTTLSeconds),
				configItem("auth.invitation_ttl_seconds", "邀请 TTL(秒)", cfg.Auth.InvitationTTLSeconds),
				configItem("auth.password_reset_ttl_seconds", "重置 TTL(秒)", cfg.Auth.PasswordResetTTLSeconds),
				configItem("auth.notification_driver", "通知驱动", cfg.Auth.NotificationDriver),
				configItem("auth.casbin_reload_interval_seconds", "权限策略刷新(秒)", cfg.Auth.CasbinReloadIntervalSeconds),
				configItem("auth.password_policy.min_length", "密码最小长度", cfg.Auth.PasswordPolicy.MinLength),
				configItem("auth.smtp.host", "SMTP 主机", cfg.Auth.SMTP.Host),
				configItem("auth.smtp.port", "SMTP 端口", cfg.Auth.SMTP.Port),
				configItem("auth.smtp.username", "SMTP 用户", cfg.Auth.SMTP.Username),
				secretConfigItem("auth.smtp.password", "SMTP 密码", cfg.Auth.SMTP.Password),
				configItem("auth.smtp.from", "SMTP 发件人", cfg.Auth.SMTP.From),
			},
		},
		{
			Code:        "logger",
			Description: "日志级别、格式和文件轮转。",
			Icon:        "scroll-text",
			Label:       "日志",
			Order:       50,
			Items: []systemmodel.ConfigItem{
				configItem("logger.level", "级别", cfg.Logger.Level),
				configItem("logger.format", "默认格式", cfg.Logger.Format),
				configItem("logger.console_format", "控制台格式", cfg.Logger.ConsoleFormat),
				configItem("logger.file_format", "文件格式", cfg.Logger.FileFormat),
				configItem("logger.output", "输出", cfg.Logger.Output),
				configItem("logger.file_path", "文件路径", cfg.Logger.FilePath),
				configItem("logger.max_size", "单文件大小(MB)", cfg.Logger.MaxSize),
				configItem("logger.max_backups", "备份数量", cfg.Logger.MaxBackups),
				configItem("logger.max_age", "保留天数", cfg.Logger.MaxAge),
			},
		},
		{
			Code:        "webui",
			Description: "内置管理台静态产物挂载和公开访问地址。",
			Icon:        "monitor",
			Label:       "WebUI",
			Order:       60,
			Items: []systemmodel.ConfigItem{
				configItem("webui.enabled", "启用", cfg.WebUI.EnabledValue()),
				configItem("webui.mount_path", "挂载路径", cfg.WebUI.MountPath),
				configItem("webui.dist_dir", "静态目录", cfg.WebUI.DistDir),
				configItem("webui.public_base_url", "公开地址", cfg.WebUI.PublicBaseURL),
			},
		},
		{
			Code:        "storage",
			Description: "文件服务类型、基础路径和监听策略。",
			Icon:        "folder",
			Label:       "文件存储",
			Order:       70,
			Items: []systemmodel.ConfigItem{
				configItem("storage.enabled", "启用", cfg.Storage.Enabled),
				configItem("storage.fs_type", "文件系统", cfg.Storage.FSType),
				configItem("storage.base_path", "基础路径", cfg.Storage.BasePath),
				configItem("storage.enable_watch", "监听变更", cfg.Storage.EnableWatch),
				configItem("storage.watch_buffer_size", "监听缓冲区", cfg.Storage.WatchBufferSize),
			},
		},
		{
			Code:        "runtime",
			Description: "跨域、国际化、迁移、Demo、执行器和 RPC 的运行策略。",
			Icon:        "settings",
			Label:       "运行策略",
			Order:       80,
			Items: []systemmodel.ConfigItem{
				configItem("cors.enabled", "启用 CORS", cfg.CORS.Enabled),
				configItem("cors.allow_origins", "允许来源", cfg.CORS.AllowOrigins),
				configItem("cors.allow_methods", "允许方法", cfg.CORS.AllowMethods),
				configItem("cors.allow_credentials", "允许凭证", cfg.CORS.AllowCredentials),
				configItem("cors.max_age", "预检缓存(秒)", cfg.CORS.MaxAge),
				configItem("i18n.default", "默认语言", cfg.I18n.Default),
				configItem("i18n.supported", "支持语言", cfg.I18n.Supported),
				configItem("i18n.messages_dir", "语言目录", cfg.I18n.MessagesDir),
				configItem("demo.enabled", "Demo 模块", cfg.Demo.EnabledValue()),
				configItem("demo.apply_schema_on_start", "启动建表示例", cfg.Demo.ApplySchemaOnStartValue()),
				configItem("migration.auto_apply", "自动迁移", cfg.Migration.AutoApply),
				configItem("migration.dir", "迁移目录", cfg.Migration.Dir),
				configItem("executor.enabled", "执行器", cfg.Executor.Enabled),
				configItem("executor.pools", "执行器池数量", len(cfg.Executor.Pools)),
				configItem("rpc.enabled", "RPC 入口", cfg.RPC.Enabled),
				configItem("rpc.host", "RPC 主机", cfg.RPC.Host),
				configItem("rpc.port", "RPC 端口", cfg.RPC.Port),
				configItem("rpc.read_timeout", "RPC 读取超时(秒)", cfg.RPC.ReadTimeout),
				configItem("rpc.write_timeout", "RPC 写入超时(秒)", cfg.RPC.WriteTimeout),
			},
		},
		{
			Code:        "plugins",
			Description: "Sidecar 插件发现、健康检查和代理策略。",
			Icon:        "blocks",
			Label:       "插件",
			Order:       90,
			Items: []systemmodel.ConfigItem{
				configItem("plugins.enabled", "启用", cfg.Plugins.Enabled),
				configItem("plugins.manifests", "Manifest 文件", cfg.Plugins.Manifests),
				configItem("plugins.items", "内联插件数量", len(cfg.Plugins.Items)),
				configItem("plugins.health_timeout_seconds", "健康检查超时(秒)", cfg.Plugins.HealthTimeoutSeconds),
				configItem("plugins.proxy_timeout_seconds", "代理超时(秒)", cfg.Plugins.ProxyTimeoutSeconds),
			},
		},
	}}
}

func configItem(key string, label string, value any) systemmodel.ConfigItem {
	return systemmodel.ConfigItem{
		Key:    key,
		Label:  label,
		Source: "runtime",
		Value:  value,
	}
}

func secretConfigItem(key string, label string, value string) systemmodel.ConfigItem {
	item := configItem(key, label, secretPresence(value))
	item.Secret = true
	return item
}

func secretPresence(value string) string {
	if strings.TrimSpace(value) == "" {
		return "未配置"
	}
	return "已配置"
}
