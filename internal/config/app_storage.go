package config

// 本文件定义一个配置分区及其校验规则，是外部配置进入运行时基础设施前的类型化边界。

import (
	"fmt"

	"github.com/rei0721/go-scaffold/pkg/storage"
)

// StorageConfig 保存 Storage 文件服务配置
type StorageConfig struct {
	// Enabled 是否启用文件服务
	Enabled bool `mapstructure:"enabled" envname:"STORAGE_ENABLED" json:"enabled" yaml:"enabled" toml:"enabled"`

	// FSType 文件系统类型
	// 可选值: os, memory, readonly, basepath
	FSType string `mapstructure:"fs_type" envname:"STORAGE_FS_TYPE" json:"fs_type" yaml:"fs_type" toml:"fs_type"`

	// BasePath 基础路径 (仅basepath类型需要)
	BasePath string `mapstructure:"base_path" envname:"STORAGE_BASE_PATH" json:"base_path" yaml:"base_path" toml:"base_path"`

	// EnableWatch 是否启用文件监听功能
	EnableWatch bool `mapstructure:"enable_watch" envname:"STORAGE_ENABLE_WATCH" json:"enable_watch" yaml:"enable_watch" toml:"enable_watch"`

	// WatchBufferSize 文件监听事件缓冲区大小
	WatchBufferSize int `mapstructure:"watch_buffer_size" envname:"STORAGE_WATCH_BUFFER_SIZE" json:"watch_buffer_size" yaml:"watch_buffer_size" toml:"watch_buffer_size"`
}

// ValidateName 返回配置名称
func (c *StorageConfig) ValidateName() string {
	return AppStorageName
}

// ValidateRequired 返回是否必需
func (c *StorageConfig) ValidateRequired() bool {
	return false
}

// Validate 验证配置有效性
func (c *StorageConfig) Validate() error {
	// 如果未启用,跳过验证
	if !c.Enabled {
		return nil
	}

	// 验证文件系统类型
	validTypes := []string{"os", "memory", "readonly", "basepath"}
	valid := false
	for _, t := range validTypes {
		if c.FSType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("storage: invalid fs_type %s, must be one of: os, memory, readonly, basepath", c.FSType)
	}

	// 验证基础路径
	if c.FSType == "basepath" && c.BasePath == "" {
		return fmt.Errorf("storage: base_path is required when fs_type is basepath")
	}

	// 验证监听缓冲区大小
	if c.WatchBufferSize < 0 {
		return fmt.Errorf("storage: watch_buffer_size must be non-negative")
	}

	return nil
}

// DefaultConfig 设置默认配置
func (c *StorageConfig) DefaultConfig() {
	if c.FSType == "" {
		c.FSType = "os"
	}
	if c.BasePath == "" {
		c.BasePath = "./data"
	}
	if c.WatchBufferSize == 0 {
		c.WatchBufferSize = 100
	}
}

// OverrideConfig 从环境变量覆盖配置
func (c *StorageConfig) OverrideConfig() {
	overrideConfigFromEnv(c)
}

// ToPkgConfig 转换为 pkg/storage.Config
func (c *StorageConfig) ToPkgConfig() *storage.Config {
	return &storage.Config{
		FSType:          storage.FSType(c.FSType),
		BasePath:        c.BasePath,
		EnableWatch:     c.EnableWatch,
		WatchBufferSize: c.WatchBufferSize,
	}
}
