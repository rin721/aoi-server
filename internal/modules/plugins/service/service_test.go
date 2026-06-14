package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProxyAddsIdentityAndSignatureHeaders(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/hello" {
			t.Fatalf("upstream path = %s, want /api/hello", r.URL.Path)
		}
		for _, header := range []string{HeaderPluginID, HeaderUserID, HeaderOrgID, HeaderTraceID, HeaderSignature, HeaderSignatureTimestamp} {
			if r.Header.Get(header) == "" {
				t.Fatalf("expected %s header", header)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer upstream.Close()
	t.Setenv("AOI_PLUGIN_TEST_SECRET", "test-plugin-secret")

	svc, err := New(Config{
		Enabled: true,
		Inline: []Manifest{{
			ID:        "demo",
			Name:      "Demo",
			Version:   "0.1.0",
			BaseURL:   upstream.URL,
			Proxy:     Proxy{Prefixes: []string{"/api"}},
			SecretRef: "AOI_PLUGIN_TEST_SECRET",
		}},
		ProxyTimeout: time.Second,
	}, nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	resp, err := svc.Proxy(context.Background(), ProxyRequest{
		PluginID: "demo",
		Method:   http.MethodGet,
		Path:     "/api/hello",
		Headers:  http.Header{"Authorization": []string{"Bearer secret"}},
		Body:     strings.NewReader(""),
		Identity: ProxyIdentity{UserID: "10", OrgID: "20", TraceID: "trace-1"},
	})
	if err != nil {
		t.Fatalf("Proxy() error = %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Proxy() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestProxyRejectsPathOutsidePrefixes(t *testing.T) {
	t.Setenv("AOI_PLUGIN_TEST_SECRET", "test-plugin-secret")
	svc, err := New(Config{
		Enabled: true,
		Inline: []Manifest{{
			ID:        "demo",
			Name:      "Demo",
			Version:   "0.1.0",
			BaseURL:   "http://127.0.0.1:1",
			Proxy:     Proxy{Prefixes: []string{"/api"}},
			SecretRef: "AOI_PLUGIN_TEST_SECRET",
		}},
	}, nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = svc.Proxy(context.Background(), ProxyRequest{
		PluginID: "demo",
		Method:   http.MethodGet,
		Path:     "/private",
		Identity: ProxyIdentity{UserID: "10", OrgID: "20", TraceID: "trace-1"},
	})
	if err != ErrProxyForbidden {
		t.Fatalf("Proxy() error = %v, want %v", err, ErrProxyForbidden)
	}
}
