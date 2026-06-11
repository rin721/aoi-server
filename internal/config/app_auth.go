package config

import "fmt"

const (
	DefaultAccessTokenTTLSeconds       = 900
	DefaultRefreshTokenTTLSeconds      = 604800
	DefaultInvitationTTLSeconds        = 86400
	DefaultPasswordResetTTLSeconds     = 1800
	DefaultCasbinReloadIntervalSeconds = 300
	DefaultLoginMaxFailures            = 5
	DefaultLoginLockMinutes            = 15
	DefaultMFAIssuer                   = "go-scaffold"
)

// AuthConfig controls the local-account IAM module.
type AuthConfig struct {
	Enabled                     bool     `mapstructure:"enabled" envname:"AUTH_ENABLED" json:"enabled" yaml:"enabled" toml:"enabled"`
	Issuer                      string   `mapstructure:"issuer" envname:"AUTH_ISSUER" json:"issuer" yaml:"issuer" toml:"issuer"`
	Audience                    []string `mapstructure:"audience" envname:"AUTH_AUDIENCE" json:"audience" yaml:"audience" toml:"audience"`
	SigningKey                  string   `mapstructure:"signing_key" envname:"AUTH_SIGNING_KEY" json:"signing_key" yaml:"signing_key" toml:"signing_key"`
	AccessTokenTTLSeconds       int      `mapstructure:"access_token_ttl_seconds" envname:"AUTH_ACCESS_TOKEN_TTL_SECONDS" json:"access_token_ttl_seconds" yaml:"access_token_ttl_seconds" toml:"access_token_ttl_seconds"`
	RefreshTokenTTLSeconds      int      `mapstructure:"refresh_token_ttl_seconds" envname:"AUTH_REFRESH_TOKEN_TTL_SECONDS" json:"refresh_token_ttl_seconds" yaml:"refresh_token_ttl_seconds" toml:"refresh_token_ttl_seconds"`
	RefreshTokenPepper          string   `mapstructure:"refresh_token_pepper" envname:"AUTH_REFRESH_TOKEN_PEPPER" json:"refresh_token_pepper" yaml:"refresh_token_pepper" toml:"refresh_token_pepper"`
	MFAIssuer                   string   `mapstructure:"mfa_issuer" envname:"AUTH_MFA_ISSUER" json:"mfa_issuer" yaml:"mfa_issuer" toml:"mfa_issuer"`
	MFASecretKey                string   `mapstructure:"mfa_secret_key" envname:"AUTH_MFA_SECRET_KEY" json:"mfa_secret_key" yaml:"mfa_secret_key" toml:"mfa_secret_key"`
	LoginMaxFailures            int      `mapstructure:"login_max_failures" envname:"AUTH_LOGIN_MAX_FAILURES" json:"login_max_failures" yaml:"login_max_failures" toml:"login_max_failures"`
	LoginLockMinutes            int      `mapstructure:"login_lock_minutes" envname:"AUTH_LOGIN_LOCK_MINUTES" json:"login_lock_minutes" yaml:"login_lock_minutes" toml:"login_lock_minutes"`
	InvitationTTLSeconds        int      `mapstructure:"invitation_ttl_seconds" envname:"AUTH_INVITATION_TTL_SECONDS" json:"invitation_ttl_seconds" yaml:"invitation_ttl_seconds" toml:"invitation_ttl_seconds"`
	PasswordResetTTLSeconds     int      `mapstructure:"password_reset_ttl_seconds" envname:"AUTH_PASSWORD_RESET_TTL_SECONDS" json:"password_reset_ttl_seconds" yaml:"password_reset_ttl_seconds" toml:"password_reset_ttl_seconds"`
	CasbinReloadIntervalSeconds int      `mapstructure:"casbin_reload_interval_seconds" envname:"AUTH_CASBIN_RELOAD_INTERVAL_SECONDS" json:"casbin_reload_interval_seconds" yaml:"casbin_reload_interval_seconds" toml:"casbin_reload_interval_seconds"`
}

func (c *AuthConfig) ValidateName() string {
	return AppAuthName
}

func (c *AuthConfig) ValidateRequired() bool {
	return false
}

func (c *AuthConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	c.ApplyDefaults()
	if c.Issuer == "" {
		return fmt.Errorf("issuer is required")
	}
	if len(c.SigningKey) < 32 {
		return fmt.Errorf("signing_key must be at least 32 bytes")
	}
	if c.RefreshTokenPepper == "" {
		return fmt.Errorf("refresh_token_pepper is required")
	}
	if len(c.MFASecretKey) < 32 {
		return fmt.Errorf("mfa_secret_key must be at least 32 bytes")
	}
	if c.AccessTokenTTLSeconds <= 0 || c.RefreshTokenTTLSeconds <= 0 {
		return fmt.Errorf("token ttl values must be positive")
	}
	if c.InvitationTTLSeconds <= 0 || c.PasswordResetTTLSeconds <= 0 {
		return fmt.Errorf("invitation and password reset ttl values must be positive")
	}
	if c.LoginMaxFailures <= 0 || c.LoginLockMinutes <= 0 {
		return fmt.Errorf("login lock policy values must be positive")
	}
	return nil
}

func (c *AuthConfig) ApplyDefaults() {
	if c.Issuer == "" {
		c.Issuer = "go-scaffold"
	}
	if len(c.Audience) == 0 {
		c.Audience = []string{"go-scaffold-api"}
	}
	if c.AccessTokenTTLSeconds == 0 {
		c.AccessTokenTTLSeconds = DefaultAccessTokenTTLSeconds
	}
	if c.RefreshTokenTTLSeconds == 0 {
		c.RefreshTokenTTLSeconds = DefaultRefreshTokenTTLSeconds
	}
	if c.MFAIssuer == "" {
		c.MFAIssuer = DefaultMFAIssuer
	}
	if c.LoginMaxFailures == 0 {
		c.LoginMaxFailures = DefaultLoginMaxFailures
	}
	if c.LoginLockMinutes == 0 {
		c.LoginLockMinutes = DefaultLoginLockMinutes
	}
	if c.InvitationTTLSeconds == 0 {
		c.InvitationTTLSeconds = DefaultInvitationTTLSeconds
	}
	if c.PasswordResetTTLSeconds == 0 {
		c.PasswordResetTTLSeconds = DefaultPasswordResetTTLSeconds
	}
	if c.CasbinReloadIntervalSeconds == 0 {
		c.CasbinReloadIntervalSeconds = DefaultCasbinReloadIntervalSeconds
	}
}
