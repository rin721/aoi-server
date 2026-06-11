package config

import "testing"

// TestWebUIConfigDefaultsAndValidation 固定管理台静态托管配置的默认值和保留路径边界。
func TestWebUIConfigDefaultsAndValidation(t *testing.T) {
	cfg := WebUIConfig{}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() with defaults error = %v", err)
	}
	if !cfg.EnabledValue() || cfg.MountPath != DefaultWebUIMountPath || cfg.DistDir != DefaultWebUIDistDir {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}

	for _, mountPath := range []string{"/", "/api", "/api/v1", "/api/v1/admin", "/health", "/ready"} {
		cfg := WebUIConfig{Enabled: boolPtr(true), MountPath: mountPath, DistDir: "./dist"}
		if err := cfg.Validate(); err == nil {
			t.Fatalf("expected mount_path %q to be rejected", mountPath)
		}
	}
}
