package configloader

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateYAMLScalarsRejectsEnvPlaceholderByDefault(t *testing.T) {
	configPath := writeYAMLScalarTestFile(t)

	err := UpdateYAMLScalars(configPath, []YAMLScalarUpdate{
		{Kind: YAMLScalarString, Path: "auth.signing_key", Value: "updated-signing-secret-at-least-32-bytes"},
	})
	if err == nil || !strings.Contains(err.Error(), "managed by environment placeholder") {
		t.Fatalf("expected environment placeholder error, got %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if !strings.Contains(string(content), "${AUTH_SIGNING_KEY:dev-secret}") {
		t.Fatalf("placeholder should remain unchanged:\n%s", content)
	}
}

func TestUpdateYAMLScalarsAllowsEnvPlaceholderOverwriteWithOption(t *testing.T) {
	configPath := writeYAMLScalarTestFile(t)

	if err := UpdateYAMLScalars(configPath, []YAMLScalarUpdate{
		{Kind: YAMLScalarString, Path: "auth.signing_key", Value: "updated-signing-secret-at-least-32-bytes"},
	}, WithEnvPlaceholderOverwrite()); err != nil {
		t.Fatalf("UpdateYAMLScalars() error = %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, `signing_key: "updated-signing-secret-at-least-32-bytes"`) {
		t.Fatalf("updated config missing forced value:\n%s", text)
	}
	if strings.Contains(text, "${AUTH_SIGNING_KEY") {
		t.Fatalf("placeholder should be overwritten:\n%s", text)
	}
}

func TestYAMLPathContainsEnvPlaceholder(t *testing.T) {
	configPath := writeYAMLScalarTestFile(t)

	hasPlaceholder, err := YAMLPathContainsEnvPlaceholder(configPath, "auth.signing_key")
	if err != nil {
		t.Fatalf("YAMLPathContainsEnvPlaceholder(signing_key) error = %v", err)
	}
	if !hasPlaceholder {
		t.Fatal("expected signing key path to contain env placeholder")
	}

	hasPlaceholder, err = YAMLPathContainsEnvPlaceholder(configPath, "auth.issuer")
	if err != nil {
		t.Fatalf("YAMLPathContainsEnvPlaceholder(issuer) error = %v", err)
	}
	if hasPlaceholder {
		t.Fatal("issuer path should not contain env placeholder")
	}
}

func writeYAMLScalarTestFile(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte(`
auth:
  issuer: go-scaffold
  signing_key: ${AUTH_SIGNING_KEY:dev-secret}
`)
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return configPath
}
