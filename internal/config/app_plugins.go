package config

import (
	"fmt"
	"strings"
)

const (
	DefaultPluginHealthTimeoutSeconds = 3
	DefaultPluginProxyTimeoutSeconds  = 30
)

// PluginsConfig 控制外部 sidecar 插件发现、健康检查和代理行为。
type PluginsConfig struct {
	Enabled              bool                   `mapstructure:"enabled" envname:"PLUGINS_ENABLED" json:"enabled" yaml:"enabled" toml:"enabled"`
	Manifests            []string               `mapstructure:"manifests" envname:"PLUGINS_MANIFESTS" json:"manifests" yaml:"manifests" toml:"manifests"`
	Items                []PluginManifestConfig `mapstructure:"items" json:"items" yaml:"items" toml:"items"`
	HealthTimeoutSeconds int                    `mapstructure:"health_timeout_seconds" envname:"PLUGINS_HEALTH_TIMEOUT_SECONDS" json:"health_timeout_seconds" yaml:"health_timeout_seconds" toml:"health_timeout_seconds"`
	ProxyTimeoutSeconds  int                    `mapstructure:"proxy_timeout_seconds" envname:"PLUGINS_PROXY_TIMEOUT_SECONDS" json:"proxy_timeout_seconds" yaml:"proxy_timeout_seconds" toml:"proxy_timeout_seconds"`
}

// PluginManifestConfig 是 Aoi Admin v1 插件清单的配置侧结构。
type PluginManifestConfig struct {
	ID          string                   `mapstructure:"id" json:"id" yaml:"id" toml:"id"`
	Name        string                   `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	Version     string                   `mapstructure:"version" json:"version" yaml:"version" toml:"version"`
	BaseURL     string                   `mapstructure:"base_url" json:"baseURL" yaml:"baseURL" toml:"baseURL"`
	HealthPath  string                   `mapstructure:"health_path" json:"healthPath" yaml:"healthPath" toml:"healthPath"`
	Frontend    PluginFrontendConfig     `mapstructure:"frontend" json:"frontend" yaml:"frontend" toml:"frontend"`
	Menus       []PluginMenuConfig       `mapstructure:"menus" json:"menus" yaml:"menus" toml:"menus"`
	Permissions []PluginPermissionConfig `mapstructure:"permissions" json:"permissions" yaml:"permissions" toml:"permissions"`
	Proxy       PluginProxyConfig        `mapstructure:"proxy" json:"proxy" yaml:"proxy" toml:"proxy"`
	SecretRef   string                   `mapstructure:"secret_ref" json:"secretRef" yaml:"secretRef" toml:"secretRef"`
}

type PluginFrontendConfig struct {
	Entry string `mapstructure:"entry" json:"entry" yaml:"entry" toml:"entry"`
}

type PluginMenuConfig struct {
	Code       string `mapstructure:"code" json:"code" yaml:"code" toml:"code"`
	Label      string `mapstructure:"label" json:"label" yaml:"label" toml:"label"`
	Icon       string `mapstructure:"icon" json:"icon" yaml:"icon" toml:"icon"`
	Path       string `mapstructure:"path" json:"path" yaml:"path" toml:"path"`
	Permission string `mapstructure:"permission" json:"permission" yaml:"permission" toml:"permission"`
	Order      int    `mapstructure:"order" json:"order" yaml:"order" toml:"order"`
}

type PluginPermissionConfig struct {
	Code        string `mapstructure:"code" json:"code" yaml:"code" toml:"code"`
	Name        string `mapstructure:"name" json:"name" yaml:"name" toml:"name"`
	Description string `mapstructure:"description" json:"description" yaml:"description" toml:"description"`
}

type PluginProxyConfig struct {
	Prefixes []string `mapstructure:"prefixes" json:"prefixes" yaml:"prefixes" toml:"prefixes"`
}

func (c *PluginsConfig) ValidateName() string {
	return AppPluginsName
}

func (c *PluginsConfig) ValidateRequired() bool {
	return false
}

func (c *PluginsConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	c.ApplyDefaults()
	seen := map[string]struct{}{}
	for i := range c.Items {
		if err := c.Items[i].Validate(); err != nil {
			return fmt.Errorf("items[%d]: %w", i, err)
		}
		id := strings.TrimSpace(c.Items[i].ID)
		if _, ok := seen[id]; ok {
			return fmt.Errorf("duplicate plugin id %q", id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func (c *PluginsConfig) ApplyDefaults() {
	if c.HealthTimeoutSeconds == 0 {
		c.HealthTimeoutSeconds = DefaultPluginHealthTimeoutSeconds
	}
	if c.ProxyTimeoutSeconds == 0 {
		c.ProxyTimeoutSeconds = DefaultPluginProxyTimeoutSeconds
	}
}

func (c *PluginManifestConfig) Validate() error {
	c.ID = strings.TrimSpace(c.ID)
	c.Name = strings.TrimSpace(c.Name)
	c.Version = strings.TrimSpace(c.Version)
	c.BaseURL = strings.TrimRight(strings.TrimSpace(c.BaseURL), "/")
	c.HealthPath = normalizePluginPath(c.HealthPath, "/health")
	c.Frontend.Entry = strings.TrimSpace(c.Frontend.Entry)
	c.SecretRef = strings.TrimSpace(c.SecretRef)

	if c.ID == "" {
		return fmt.Errorf("id is required")
	}
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("baseURL is required")
	}
	for i := range c.Menus {
		if err := c.Menus[i].Validate(); err != nil {
			return fmt.Errorf("menus[%d]: %w", i, err)
		}
	}
	for i := range c.Permissions {
		if err := c.Permissions[i].Validate(); err != nil {
			return fmt.Errorf("permissions[%d]: %w", i, err)
		}
	}
	for i := range c.Proxy.Prefixes {
		c.Proxy.Prefixes[i] = normalizePluginPath(c.Proxy.Prefixes[i], "")
		if c.Proxy.Prefixes[i] == "" {
			return fmt.Errorf("proxy.prefixes[%d] is required", i)
		}
	}
	if len(c.Proxy.Prefixes) > 0 && c.SecretRef == "" {
		return fmt.Errorf("secretRef is required when proxy prefixes are configured")
	}
	return nil
}

func (c *PluginMenuConfig) Validate() error {
	c.Code = strings.TrimSpace(c.Code)
	c.Label = strings.TrimSpace(c.Label)
	c.Icon = strings.TrimSpace(c.Icon)
	c.Path = normalizePluginPath(c.Path, "/")
	c.Permission = strings.TrimSpace(c.Permission)
	if c.Code == "" {
		return fmt.Errorf("code is required")
	}
	if c.Label == "" {
		return fmt.Errorf("label is required")
	}
	return nil
}

func (c *PluginPermissionConfig) Validate() error {
	c.Code = strings.TrimSpace(c.Code)
	c.Name = strings.TrimSpace(c.Name)
	c.Description = strings.TrimSpace(c.Description)
	if c.Code == "" {
		return fmt.Errorf("code is required")
	}
	if c.Name == "" {
		c.Name = c.Code
	}
	return nil
}

func normalizePluginPath(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	return "/" + strings.Trim(strings.TrimRight(value, "/"), "/")
}

func copyPluginsConfig(src PluginsConfig) PluginsConfig {
	dst := src
	dst.Manifests = append([]string(nil), src.Manifests...)
	dst.Items = append([]PluginManifestConfig(nil), src.Items...)
	for i := range dst.Items {
		dst.Items[i].Menus = append([]PluginMenuConfig(nil), src.Items[i].Menus...)
		dst.Items[i].Permissions = append([]PluginPermissionConfig(nil), src.Items[i].Permissions...)
		dst.Items[i].Proxy.Prefixes = append([]string(nil), src.Items[i].Proxy.Prefixes...)
	}
	return dst
}
