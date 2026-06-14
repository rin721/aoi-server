package initapp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rei0721/go-scaffold/internal/config"
)

func TestPluginsServiceConfigLoadsManifestFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plugin.yaml")
	raw := []byte(`
id: demo
name: Demo
version: 0.1.0
baseURL: http://127.0.0.1:10098
healthPath: /healthz
frontend:
  entry: /assets/remote.js
menus:
  - code: demo.home
    label: Demo
    path: /
proxy:
  prefixes:
    - /api
secretRef: AOI_PLUGIN_TEST_SECRET
`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	cfg, err := PluginsServiceConfig(config.PluginsConfig{
		Enabled:   true,
		Manifests: []string{path},
	})
	if err != nil {
		t.Fatalf("PluginsServiceConfig() error = %v", err)
	}
	if len(cfg.Inline) != 1 {
		t.Fatalf("inline manifests = %d, want 1", len(cfg.Inline))
	}
	manifest := cfg.Inline[0]
	if manifest.BaseURL != "http://127.0.0.1:10098" || manifest.HealthPath != "/healthz" || manifest.SecretRef != "AOI_PLUGIN_TEST_SECRET" {
		t.Fatalf("manifest fields not decoded: %#v", manifest)
	}
}
