package config

// 本文件属于配置子系统，处理配置加载、环境变量覆盖、运行时快照或跨分区校验。

import "fmt"

// Configurable 表示可被聚合配置校验器识别的配置分区契约。
type Configurable interface {
	Validate() error
}

// Config 聚合应用所有运行时配置分区，是配置管理器发布给组装层的完整快照。
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	I18n      I18nConfig      `mapstructure:"i18n"`
	Executor  ExecutorConfig  `mapstructure:"executor"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Demo      DemoConfig      `mapstructure:"demo"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RPC       RPCConfig       `mapstructure:"rpc"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Migration MigrationConfig `mapstructure:"migration"`
	WebUI     WebUIConfig     `mapstructure:"webui"`
}

// Validator 表示可以对自身执行配置校验的分区接口。
type Validator interface {
	Validate() error
	ValidateName() string
	ValidateRequired() bool
}

// Validate 对完整应用配置执行跨分区校验，返回第一个阻断启动的配置错误。
func (c *Config) Validate() error {
	validators := []Validator{
		&c.Server,
		&c.Database,
		&c.Redis,
		&c.Logger,
		&c.I18n,
		&c.Executor,
		&c.Storage,
		&c.Demo,
		&c.CORS,
		&c.RPC,
		&c.Auth,
		&c.Migration,
		&c.WebUI,
	}
	for _, validator := range validators {
		if validator == nil {
			continue
		}
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("%s config: %w", validator.ValidateName(), err)
		}
	}
	return nil
}

// ValidateOld 保留旧版校验入口，供兼容路径在迁移期复用当前校验逻辑。
func (c *Config) ValidateOld() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config: %w", err)
	}
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config: %w", err)
	}
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("redis config: %w", err)
	}
	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger config: %w", err)
	}
	if err := c.I18n.Validate(); err != nil {
		return fmt.Errorf("i18n config: %w", err)
	}
	if err := c.Executor.Validate(); err != nil {
		return fmt.Errorf("executor config: %w", err)
	}
	if err := c.WebUI.Validate(); err != nil {
		return fmt.Errorf("webui config: %w", err)
	}
	return nil
}
