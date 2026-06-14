package cliapp

import (
	"context"
	"fmt"
	"io"
	"strings"

	appconfig "github.com/rei0721/go-scaffold/internal/config"
	"github.com/rei0721/go-scaffold/pkg/cli"
	"github.com/rei0721/go-scaffold/types/constants"
)

var (
	newFlowManager        = NewManager
	executeInitialization = ExecuteInitialization
)

const (
	privacyCoreActionGenerateFile = "generate_file"
	privacyCoreActionManual       = "manual"
)

// RunStartFlow 执行由 pkg/cli UI 驱动的服务启动流程。
func RunStartFlow(ctx *cli.Context) error {
	if ctx == nil {
		return fmt.Errorf("cli context is required")
	}
	if ctx.Context == nil {
		ctx.Context = context.Background()
	}
	ui, err := requireUI(ctx)
	if err != nil {
		return err
	}

	service, err := cli.SelectKey(ctx.Context, ui, "service", "选择服务", startServiceOptions())
	if err != nil {
		return err
	}
	service = normalizeStartService(service)
	if !isSupportedStartService(service) {
		return fmt.Errorf("unsupported service %q; expected one of: server, db, iam, cache, storage", service)
	}

	configPath, err := selectConfigPath(ctx)
	if err != nil {
		return err
	}
	recoveredCoreSecrets, err := printConfigSummaryForStart(ctx, service, configPath)
	if err != nil {
		return err
	}
	if service != ServiceServer {
		return printDependencyServiceInfo(ctx.Stdout, service, configPath)
	}
	if recoveredCoreSecrets {
		return startServer(ctx, configPath)
	}
	ok, err := cli.ConfirmKey(ctx.Context, ui, "privacy", "是否填写或生成隐私配置？", false)
	if err != nil {
		return err
	}
	if !ok {
		return startServer(ctx, configPath)
	}
	if isExampleConfig(configPath) {
		return fmt.Errorf("示例配置只读，不能写入隐私配置")
	}
	updates, err := promptPrivacyUpdates(ctx, configPath)
	if err != nil {
		return err
	}
	if updates.hasChanges() {
		if err := ApplyPrivacyRuntimeEnvOnly(configPath, updates.runtimeEnvOnlyPaths); err != nil {
			return err
		}
		if err := ApplyPrivacyUpdates(configPath, updates.fileUpdates); err != nil {
			return err
		}
		if err := ApplyPrivacyUpdates(configPath, updates.forceFileUpdates, appconfig.WithEnvManagedPersistMode(appconfig.EnvManagedPersistForceFile)); err != nil {
			return err
		}
		_ = ui.Info("隐私配置已处理。")
	}
	return startServer(ctx, configPath)
}

func printConfigSummaryForStart(ctx *cli.Context, service string, configPath string) (bool, error) {
	err := PrintConfigSummary(ctx.Stdout, configPath)
	if err == nil {
		return false, nil
	}
	if service != ServiceServer || !isCoreSecretConfigError(err) {
		return false, err
	}
	if !canPromptCoreSecretRecovery(ctx) {
		return false, coreSecretConfigError(configPath, err)
	}
	recovered, recoverErr := promptCoreSecretRecovery(ctx, configPath)
	if recoverErr != nil {
		return false, recoverErr
	}
	if !recovered {
		return false, coreSecretConfigError(configPath, err)
	}
	if err := PrintConfigSummary(ctx.Stdout, configPath); err != nil {
		if isCoreSecretConfigError(err) {
			return false, coreSecretConfigError(configPath, err)
		}
		return false, err
	}
	return true, nil
}

func canPromptCoreSecretRecovery(ctx *cli.Context) bool {
	if ctx == nil || ctx.UI == nil || ctx.GetBool("yes") {
		return false
	}
	if value, ok := cli.PromptAnswer(ctx.UI, "privacy"); ok {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "false" || value == "f" || value == "no" || value == "n" || value == "0" {
			return false
		}
	}
	return true
}

func promptCoreSecretRecovery(ctx *cli.Context, configPath string) (bool, error) {
	if isExampleConfig(configPath) {
		return false, fmt.Errorf("example config is read-only for secret generation; copy it to a real config file or set %s", coreSecretEnvHelp())
	}
	if coreSecretValueAnswersProvided(ctx.UI) {
		return promptAndWriteCoreSecrets(ctx, configPath)
	}
	action, err := cli.SelectKey(ctx.Context, ctx.UI, "privacy.core_secrets.action", "IAM core secrets are missing; choose a repair action", []cli.SelectOption{
		{Value: privacyCoreActionGenerateFile, Label: "生成并写入配置文件", Description: "生成稳定随机密钥并禁用这些路径的环境变量覆盖"},
		{Value: privacyActionRuntimeEnvOnly, Label: "继续使用环境变量", Description: "校验真实环境变量，配置文件不写入密钥"},
		{Value: privacyCoreActionManual, Label: "逐项输入", Description: "手动输入；也可以输入 generate 自动生成单项"},
		{Value: privacyActionSkip, Label: "跳过", Description: "保留当前配置并返回可操作错误"},
	})
	if err != nil {
		return false, err
	}
	switch action {
	case privacyCoreActionGenerateFile:
		updates := map[string]string{}
		for _, path := range coreSecretPaths {
			updates[path] = randomSecret()
		}
		if err := applyPrivacyForceFileUpdates(configPath, updates); err != nil {
			return false, err
		}
		_ = ctx.UI.Info("IAM core secrets generated and written to config.")
		return true, nil
	case privacyActionRuntimeEnvOnly:
		if err := applyPrivacyRuntimeEnvOnlyDirect(configPath, coreSecretPaths); err != nil {
			return false, err
		}
		_ = ctx.UI.Info("IAM core secrets will be read from environment variables.")
		return true, nil
	case privacyCoreActionManual:
		return promptAndWriteCoreSecrets(ctx, configPath)
	case privacyActionSkip, "":
		return false, nil
	default:
		return false, fmt.Errorf("unknown IAM core secret repair action %q", action)
	}
}

func coreSecretValueAnswersProvided(ui cli.PromptUI) bool {
	for _, path := range coreSecretPaths {
		if _, ok := cli.PromptAnswer(ui, "privacy."+path+".value"); ok {
			return true
		}
	}
	return false
}

func promptAndWriteCoreSecrets(ctx *cli.Context, configPath string) (bool, error) {
	updates := map[string]string{}
	for _, path := range coreSecretPaths {
		value, err := promptPrivacyValue(ctx, path)
		if err != nil {
			return false, err
		}
		if value != "" {
			updates[path] = value
		}
	}
	if len(updates) == 0 {
		return false, nil
	}
	if err := applyPrivacyForceFileUpdates(configPath, updates); err != nil {
		return false, err
	}
	_ = ctx.UI.Info("IAM core secrets written to config.")
	return true, nil
}

func startServiceOptions() []cli.SelectOption {
	return []cli.SelectOption{
		{Value: ServiceServer, Label: "server", Description: "后台托管 HTTP 服务进程"},
		{Value: "db", Label: "db", Description: "数据库配置、迁移和初始化能力"},
		{Value: "iam", Label: "iam", Description: "IAM 管理员、角色权限和 API Token"},
		{Value: "cache", Label: "cache", Description: "Redis 缓存依赖状态"},
		{Value: "storage", Label: "storage", Description: "存储配置和依赖状态"},
	}
}

func normalizeStartService(service string) string {
	return strings.ToLower(strings.TrimSpace(service))
}

func isSupportedStartService(service string) bool {
	switch normalizeStartService(service) {
	case ServiceServer, "db", "iam", "cache", "storage":
		return true
	default:
		return false
	}
}

func startServer(ctx *cli.Context, configPath string) error {
	state, err := newFlowManager().StartServer(ctx.Context, configPath)
	if err != nil {
		return err
	}
	PrintServiceState(ctx.Stdout, state)
	return nil
}

// RunServiceFlow 执行由 pkg/cli UI 驱动的服务管理流程。
func RunServiceFlow(ctx *cli.Context) error {
	ui, err := requireUI(ctx)
	if err != nil {
		return err
	}
	manager := newFlowManager()
	_, singleAction := cli.PromptAnswer(ui, "action")
	for {
		action, err := cli.SelectKey(ctx.Context, ui, "action", "服务管理：server", []cli.SelectOption{
			{Value: "status", Label: "查看运行状态"},
			{Value: "info", Label: "查看服务信息"},
			{Value: "logs", Label: "查看服务日志"},
			{Value: "terminal", Label: "进入服务终端"},
			{Value: "restart", Label: "重启服务"},
			{Value: "stop", Label: "停止服务"},
			{Value: "back", Label: "退出服务"},
		})
		if err != nil {
			return err
		}
		switch action {
		case "status":
			state, err := manager.Status(ctx.Context, ServiceServer)
			if err != nil {
				return err
			}
			PrintServiceState(ctx.Stdout, state)
		case "info":
			state, err := manager.Status(ctx.Context, ServiceServer)
			if err != nil {
				return err
			}
			PrintServiceState(ctx.Stdout, state)
		case "logs":
			state, err := manager.Status(ctx.Context, ServiceServer)
			if err != nil {
				return err
			}
			follow, err := cli.ConfirmKey(ctx.Context, ui, "logs.follow", "是否实时跟随日志？", false)
			if err != nil {
				return err
			}
			if err := PrintServiceLogs(ctx.Context, ctx.Stdout, state, 100, follow); err != nil {
				return err
			}
		case "terminal":
			state, err := manager.Status(ctx.Context, ServiceServer)
			if err != nil {
				return err
			}
			if err := PrintServiceLogs(ctx.Context, ctx.Stdout, state, 100, true); err != nil {
				return err
			}
		case "restart":
			state, err := manager.RestartServer(ctx.Context)
			if err != nil {
				return err
			}
			PrintServiceState(ctx.Stdout, state)
		case "stop":
			state, err := manager.StopServer(ctx.Context)
			if err != nil {
				return err
			}
			PrintServiceState(ctx.Stdout, state)
		case "back":
			return nil
		}
		if singleAction {
			return nil
		}
	}
}

// RunInitializationFlow 执行由 pkg/cli UI 驱动的初始化流程。
func RunInitializationFlow(ctx *cli.Context, input InitializationInput) error {
	configPath, err := selectConfigPath(ctx)
	if err != nil {
		return err
	}
	input.ConfigPath = configPath
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	ui, err := requireUI(ctx)
	if err != nil {
		return err
	}
	if cfg.Auth.Enabled && input.AdminPassword == "" {
		if !ctx.IsFlagChanged("org-code") {
			input.OrgCode, err = cli.InputKey(ctx.Context, ui, "org-code", "组织 code", defaultString(input.OrgCode, "acme"))
			if err != nil {
				return err
			}
		}
		if !ctx.IsFlagChanged("org-name") {
			input.OrgName, err = cli.InputKey(ctx.Context, ui, "org-name", "组织名称", defaultString(input.OrgName, input.OrgCode))
			if err != nil {
				return err
			}
		}
		if !ctx.IsFlagChanged("admin-username") {
			input.AdminUsername, err = cli.InputKey(ctx.Context, ui, "admin-username", "管理员用户名", defaultString(input.AdminUsername, "admin"))
			if err != nil {
				return err
			}
		}
		if !ctx.IsFlagChanged("admin-email") {
			input.AdminEmail, err = cli.InputKey(ctx.Context, ui, "admin-email", "管理员邮箱", defaultString(input.AdminEmail, "admin@example.com"))
			if err != nil {
				return err
			}
		}
		if !ctx.IsFlagChanged("admin-display-name") {
			input.AdminDisplayName, err = cli.InputKey(ctx.Context, ui, "admin-display-name", "管理员显示名", defaultString(input.AdminDisplayName, input.AdminUsername))
			if err != nil {
				return err
			}
		}
		input.AdminPassword, err = cli.PasswordKey(ctx.Context, ui, "admin-password", "管理员密码，留空跳过管理员创建")
		if err != nil {
			return err
		}
		if input.AdminPassword != "" && !ctx.IsFlagChanged("create-service-token") {
			input.CreateServiceToken, err = cli.ConfirmKey(ctx.Context, ui, "create-service-token", "是否创建服务账户 API Token？", false)
			if err != nil {
				return err
			}
		}
	}
	if input.CreateServiceToken {
		if !ctx.IsFlagChanged("service-token-days") {
			days, err := cli.InputKey(ctx.Context, ui, "service-token-days", "Token 有效天数，-1 表示永不过期", fmt.Sprint(defaultInt(input.ServiceTokenDays, 30)))
			if err != nil {
				return err
			}
			fmt.Sscanf(days, "%d", &input.ServiceTokenDays)
		}
		if !ctx.IsFlagChanged("service-token-remark") {
			input.ServiceTokenRemark, err = cli.InputKey(ctx.Context, ui, "service-token-remark", "Token 备注", defaultString(input.ServiceTokenRemark, "created by cli init"))
			if err != nil {
				return err
			}
		}
	}
	return executeInitialization(ctx.Context, ctx.Stdout, input)
}

func selectConfigPath(ctx *cli.Context) (string, error) {
	if ctx.IsFlagChanged("config") && strings.TrimSpace(ctx.GetString("config")) != "" {
		return ctx.GetString("config"), nil
	}
	if value, ok := cli.PromptAnswer(ctx.UI, "config"); ok {
		value = strings.TrimSpace(value)
		if value == "" {
			return constants.AppDefaultConfigPath, nil
		}
		return value, nil
	}
	files := DiscoverConfigFiles()
	if len(files) == 0 {
		return constants.AppDefaultConfigPath, nil
	}
	if _, err := requireUI(ctx); err != nil {
		return "", err
	}
	options := make([]cli.SelectOption, 0, len(files)+1)
	for _, file := range files {
		description := ""
		if isExampleConfig(file) {
			description = "示例配置，隐私配置只读"
		}
		options = append(options, cli.SelectOption{Value: file, Label: file, Description: description})
	}
	options = append(options, cli.SelectOption{Value: "__custom__", Label: "手动输入路径"})
	selected, err := cli.SelectKey(ctx.Context, ctx.UI, "config", "选择配置文件", options)
	if err != nil {
		return "", err
	}
	if selected == "__custom__" {
		return cli.InputKey(ctx.Context, ctx.UI, "config.custom", "配置文件路径", constants.AppDefaultConfigPath)
	}
	return selected, nil
}

func promptPrivacyUpdates(ctx *cli.Context, configPath string) (privacyPersistPlan, error) {
	paths, err := privacyPaths(configPath)
	if err != nil {
		return privacyPersistPlan{}, err
	}
	updates := newPrivacyPersistPlan()
	for _, path := range paths {
		envManaged, err := privacyPathIsEnvManaged(configPath, path)
		if err != nil {
			return privacyPersistPlan{}, err
		}
		if envManaged {
			action, err := cli.SelectKey(ctx.Context, ctx.UI, "privacy."+path+".action", path+" 由环境变量管理，选择处理方式", []cli.SelectOption{
				{Value: privacyActionForceFile, Label: "写入配置文件并禁用环境变量覆盖", Description: "替换配置文件中的环境变量占位符"},
				{Value: privacyActionRuntimeEnvOnly, Label: "恢复使用环境变量", Description: "校验真实环境变量并移除配置文件优先标记"},
				{Value: privacyActionSkip, Label: "跳过", Description: "保留当前配置"},
			})
			if err != nil {
				return privacyPersistPlan{}, err
			}
			switch action {
			case privacyActionRuntimeEnvOnly:
				updates.runtimeEnvOnlyPaths = append(updates.runtimeEnvOnlyPaths, path)
				continue
			case privacyActionSkip, "":
				continue
			case privacyActionForceFile:
				value, err := promptPrivacyValue(ctx, path)
				if err != nil {
					return privacyPersistPlan{}, err
				}
				if value != "" {
					updates.forceFileUpdates[path] = value
				}
				continue
			default:
				return privacyPersistPlan{}, fmt.Errorf("unknown privacy config action %q", action)
			}
		}

		value, err := promptPrivacyValue(ctx, path)
		if err != nil {
			return privacyPersistPlan{}, err
		}
		if value == "" {
			continue
		}
		updates.fileUpdates[path] = value
	}
	return updates, nil
}

func promptPrivacyValue(ctx *cli.Context, path string) (string, error) {
	hint := "留空跳过"
	if isGeneratedSecretPath(path) {
		hint = "输入 generate 自动生成，留空跳过"
	}
	value, err := cli.InputKey(ctx.Context, ctx.UI, "privacy."+path+".value", path+"（"+hint+"）", "")
	if err != nil {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if strings.EqualFold(value, "generate") && isGeneratedSecretPath(path) {
		return randomSecret(), nil
	}
	return value, nil
}

func privacyPaths(configPath string) ([]string, error) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return nil, err
	}
	paths := append([]string(nil), coreSecretPaths...)
	if !strings.EqualFold(cfg.Database.Driver, "sqlite") {
		paths = append(paths, "database.password")
	}
	if cfg.Redis.Enabled {
		paths = append(paths, "redis.password")
	}
	if strings.EqualFold(cfg.Auth.NotificationDriver, "smtp") {
		paths = append(paths, "auth.smtp.password")
	}
	return paths, nil
}

func printDependencyServiceInfo(w io.Writer, service string, configPath string) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		return err
	}
	switch service {
	case "db":
		fmt.Fprintf(w, "db：driver=%s dbname=%s\n", cfg.Database.Driver, cfg.Database.DBName)
	case "iam":
		fmt.Fprintf(w, "iam：enabled=%v issuer=%s\n", cfg.Auth.Enabled, cfg.Auth.Issuer)
	case "cache":
		fmt.Fprintf(w, "cache：redis.enabled=%v addr=%s:%d\n", cfg.Redis.Enabled, cfg.Redis.Host, cfg.Redis.Port)
	case "storage":
		fmt.Fprintf(w, "storage：enabled=%v fs_type=%s base_path=%s\n", cfg.Storage.Enabled, cfg.Storage.FSType, cfg.Storage.BasePath)
	}
	fmt.Fprintln(w, "v1 仅托管 server 后台进程；该服务作为配置、初始化、帮助和依赖状态展示。")
	return nil
}

func requireUI(ctx *cli.Context) (cli.PromptUI, error) {
	if ctx == nil || ctx.UI == nil {
		return nil, fmt.Errorf("interactive UI is not available")
	}
	if ctx.Context == nil {
		ctx.Context = context.Background()
	}
	return ctx.UI, nil
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func defaultInt(value int, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}
