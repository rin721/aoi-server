package config

// 本文件定义 Demo 配置分区及其校验规则，是外部配置进入运行时 Demo 模块前的类型化边界。

// DemoConfig 控制脚手架 Demo 模块是否注册，以及 server 启动期是否允许隐式建表。
type DemoConfig struct {
	Enabled            *bool `mapstructure:"enabled" envname:"DEMO_ENABLED" json:"enabled" yaml:"enabled" toml:"enabled"`
	ApplySchemaOnStart *bool `mapstructure:"apply_schema_on_start" envname:"DEMO_APPLY_SCHEMA_ON_START" json:"apply_schema_on_start" yaml:"apply_schema_on_start" toml:"apply_schema_on_start"`
}

// ValidateName 返回当前配置分区在聚合校验错误中的稳定名称。
func (c *DemoConfig) ValidateName() string {
	return AppDemoName
}

// ValidateRequired 声明当前配置分区是否必须出现在完整应用配置中。
func (c *DemoConfig) ValidateRequired() bool {
	return false
}

// Validate 校验当前配置分区的取值范围，并在不满足运行时约束时返回错误。
func (c *DemoConfig) Validate() error {
	return nil
}

// EnabledValue 解析 Demo 开关指针的有效值，nil 表示采用脚手架默认启用策略。
func (c DemoConfig) EnabledValue() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

// ApplySchemaOnStartValue 解析启动期建表开关，nil 表示沿用 server 启动的默认 schema 策略。
func (c DemoConfig) ApplySchemaOnStartValue() bool {
	if c.ApplySchemaOnStart == nil {
		return c.EnabledValue()
	}
	return *c.ApplySchemaOnStart
}

// copyDemoConfig 深拷贝 DemoConfig 的指针字段，避免配置快照之间共享可变布尔值。
func copyDemoConfig(src DemoConfig) DemoConfig {
	dst := src
	if src.Enabled != nil {
		value := *src.Enabled
		dst.Enabled = &value
	}
	if src.ApplySchemaOnStart != nil {
		value := *src.ApplySchemaOnStart
		dst.ApplySchemaOnStart = &value
	}
	return dst
}
