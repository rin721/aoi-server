package config

import (
	"fmt"
	"strings"
)

const (
	DefaultAccessTokenTTLSeconds       = 900
	DefaultRefreshTokenTTLSeconds      = 604800
	DefaultInvitationTTLSeconds        = 86400
	DefaultPasswordResetTTLSeconds     = 1800
	DefaultCasbinReloadIntervalSeconds = 300
	DefaultLoginMaxFailures            = 5
	DefaultLoginLockMinutes            = 15
	DefaultCaptchaTTLSeconds           = 120
	DefaultMFAIssuer                   = "go-scaffold"
	DefaultNotificationDriver          = "debug"
	DefaultPasswordMinLength           = 8
)

// AuthConfig controls the local-account IAM module.
type AuthConfig struct {
	Enabled                     bool                 `mapstructure:"enabled" envname:"AUTH_ENABLED" json:"enabled" yaml:"enabled" toml:"enabled"`
	SelfSignupEnabled           bool                 `mapstructure:"self_signup_enabled" envname:"AUTH_SELF_SIGNUP_ENABLED" json:"self_signup_enabled" yaml:"self_signup_enabled" toml:"self_signup_enabled"`
	Issuer                      string               `mapstructure:"issuer" envname:"AUTH_ISSUER" json:"issuer" yaml:"issuer" toml:"issuer"`
	Audience                    []string             `mapstructure:"audience" envname:"AUTH_AUDIENCE" json:"audience" yaml:"audience" toml:"audience"`
	SigningKey                  string               `mapstructure:"signing_key" envname:"AUTH_SIGNING_KEY" json:"signing_key" yaml:"signing_key" toml:"signing_key"`
	AccessTokenTTLSeconds       int                  `mapstructure:"access_token_ttl_seconds" envname:"AUTH_ACCESS_TOKEN_TTL_SECONDS" json:"access_token_ttl_seconds" yaml:"access_token_ttl_seconds" toml:"access_token_ttl_seconds"`
	RefreshTokenTTLSeconds      int                  `mapstructure:"refresh_token_ttl_seconds" envname:"AUTH_REFRESH_TOKEN_TTL_SECONDS" json:"refresh_token_ttl_seconds" yaml:"refresh_token_ttl_seconds" toml:"refresh_token_ttl_seconds"`
	RefreshTokenPepper          string               `mapstructure:"refresh_token_pepper" envname:"AUTH_REFRESH_TOKEN_PEPPER" json:"refresh_token_pepper" yaml:"refresh_token_pepper" toml:"refresh_token_pepper"`
	MFAIssuer                   string               `mapstructure:"mfa_issuer" envname:"AUTH_MFA_ISSUER" json:"mfa_issuer" yaml:"mfa_issuer" toml:"mfa_issuer"`
	MFASecretKey                string               `mapstructure:"mfa_secret_key" envname:"AUTH_MFA_SECRET_KEY" json:"mfa_secret_key" yaml:"mfa_secret_key" toml:"mfa_secret_key"`
	LoginMaxFailures            int                  `mapstructure:"login_max_failures" envname:"AUTH_LOGIN_MAX_FAILURES" json:"login_max_failures" yaml:"login_max_failures" toml:"login_max_failures"`
	LoginLockMinutes            int                  `mapstructure:"login_lock_minutes" envname:"AUTH_LOGIN_LOCK_MINUTES" json:"login_lock_minutes" yaml:"login_lock_minutes" toml:"login_lock_minutes"`
	LoginCaptchaEnabled         bool                 `mapstructure:"login_captcha_enabled" envname:"AUTH_LOGIN_CAPTCHA_ENABLED" json:"login_captcha_enabled" yaml:"login_captcha_enabled" toml:"login_captcha_enabled"`
	CaptchaTTLSeconds           int                  `mapstructure:"captcha_ttl_seconds" envname:"AUTH_CAPTCHA_TTL_SECONDS" json:"captcha_ttl_seconds" yaml:"captcha_ttl_seconds" toml:"captcha_ttl_seconds"`
	InvitationTTLSeconds        int                  `mapstructure:"invitation_ttl_seconds" envname:"AUTH_INVITATION_TTL_SECONDS" json:"invitation_ttl_seconds" yaml:"invitation_ttl_seconds" toml:"invitation_ttl_seconds"`
	PasswordResetTTLSeconds     int                  `mapstructure:"password_reset_ttl_seconds" envname:"AUTH_PASSWORD_RESET_TTL_SECONDS" json:"password_reset_ttl_seconds" yaml:"password_reset_ttl_seconds" toml:"password_reset_ttl_seconds"`
	NotificationDriver          string               `mapstructure:"notification_driver" envname:"AUTH_NOTIFICATION_DRIVER" json:"notification_driver" yaml:"notification_driver" toml:"notification_driver"`
	SMTP                        SMTPConfig           `mapstructure:"smtp" json:"smtp" yaml:"smtp" toml:"smtp"`
	PasswordPolicy              PasswordPolicyConfig `mapstructure:"password_policy" json:"password_policy" yaml:"password_policy" toml:"password_policy"`
	CasbinReloadIntervalSeconds int                  `mapstructure:"casbin_reload_interval_seconds" envname:"AUTH_CASBIN_RELOAD_INTERVAL_SECONDS" json:"casbin_reload_interval_seconds" yaml:"casbin_reload_interval_seconds" toml:"casbin_reload_interval_seconds"`
}

// SMTPConfig 定义 IAM 邀请和找回密码邮件通知的 SMTP 连接参数。
type SMTPConfig struct {
	Host     string `mapstructure:"host" envname:"AUTH_SMTP_HOST" json:"host" yaml:"host" toml:"host"`
	Port     int    `mapstructure:"port" envname:"AUTH_SMTP_PORT" json:"port" yaml:"port" toml:"port"`
	Username string `mapstructure:"username" envname:"AUTH_SMTP_USERNAME" json:"username" yaml:"username" toml:"username"`
	Password string `mapstructure:"password" envname:"AUTH_SMTP_PASSWORD" json:"password" yaml:"password" toml:"password"`
	From     string `mapstructure:"from" envname:"AUTH_SMTP_FROM" json:"from" yaml:"from" toml:"from"`
	FromName string `mapstructure:"from_name" envname:"AUTH_SMTP_FROM_NAME" json:"from_name" yaml:"from_name" toml:"from_name"`
	StartTLS bool   `mapstructure:"starttls" envname:"AUTH_SMTP_STARTTLS" json:"starttls" yaml:"starttls" toml:"starttls"`
}

// PasswordPolicyConfig 定义账号创建和密码重置时的最低密码要求。
type PasswordPolicyConfig struct {
	MinLength     int  `mapstructure:"min_length" envname:"AUTH_PASSWORD_MIN_LENGTH" json:"min_length" yaml:"min_length" toml:"min_length"`
	RequireLower  bool `mapstructure:"require_lower" envname:"AUTH_PASSWORD_REQUIRE_LOWER" json:"require_lower" yaml:"require_lower" toml:"require_lower"`
	RequireUpper  bool `mapstructure:"require_upper" envname:"AUTH_PASSWORD_REQUIRE_UPPER" json:"require_upper" yaml:"require_upper" toml:"require_upper"`
	RequireNumber bool `mapstructure:"require_number" envname:"AUTH_PASSWORD_REQUIRE_NUMBER" json:"require_number" yaml:"require_number" toml:"require_number"`
	RequireSymbol bool `mapstructure:"require_symbol" envname:"AUTH_PASSWORD_REQUIRE_SYMBOL" json:"require_symbol" yaml:"require_symbol" toml:"require_symbol"`
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
	if c.LoginCaptchaEnabled && c.CaptchaTTLSeconds <= 0 {
		return fmt.Errorf("captcha_ttl_seconds must be positive when login captcha is enabled")
	}
	if c.PasswordPolicy.MinLength <= 0 {
		return fmt.Errorf("password policy min_length must be positive")
	}
	if strings.EqualFold(c.NotificationDriver, "smtp") {
		if c.SMTP.Host == "" || c.SMTP.Port <= 0 || c.SMTP.From == "" {
			return fmt.Errorf("smtp host, port, and from are required when notification_driver is smtp")
		}
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
	if c.CaptchaTTLSeconds == 0 {
		c.CaptchaTTLSeconds = DefaultCaptchaTTLSeconds
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
	if c.NotificationDriver == "" {
		c.NotificationDriver = DefaultNotificationDriver
	}
	if c.SMTP.Port == 0 {
		c.SMTP.Port = 587
	}
	if c.PasswordPolicy.MinLength == 0 {
		c.PasswordPolicy.MinLength = DefaultPasswordMinLength
	}
}
