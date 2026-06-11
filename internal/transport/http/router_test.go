package httptransport

// 本测试文件固定 HTTP 路由、中间件顺序和错误响应契约，防止注释补全和后续重构改变外部可观察行为。

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/web"
	apperrors "github.com/rei0721/go-scaffold/types/errors"
)

type fakeDatabase struct {
	database.Database
	pingErr error
}

// Close 实现测试桩的资源关闭入口，用于验证生命周期调用而不释放外部资源。
func (db *fakeDatabase) Close() error {
	return nil
}

// Ping 实现数据库测试桩的健康检查入口，按测试需要返回成功或预设错误。
func (db *fakeDatabase) Ping(context.Context) error {
	return db.pingErr
}

// Reload 实现测试桩的配置重载入口，用于验证调用路径而不触发真实资源替换。
func (db *fakeDatabase) Reload(*database.Config) error {
	return nil
}

type routerResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
}

// TestNewRouterHealthEndpoint 固定 HTTP 路由、中间件顺序和错误响应契约，确保后续注释补全或结构调整不改变该场景。
func TestNewRouterHealthEndpoint(t *testing.T) {
	router := newTestRouter(RouterDeps{})

	recorder, body := performRouterRequest(t, router, http.MethodGet, "/health")

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected /health status %d, got %d with body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	assertSuccessResponse(t, body)
	assertDataValue(t, body.Data, "status", "ok")
}

// TestNewRouterReadyEndpoint 固定 HTTP 路由、中间件顺序和错误响应契约，确保后续注释补全或结构调整不改变该场景。
func TestNewRouterReadyEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		db             database.Database
		wantHTTPStatus int
		wantCode       int
		wantMessage    string
		wantStatus     string
		wantDBCheck    string
	}{
		{
			name:           "missing database",
			db:             nil,
			wantHTTPStatus: http.StatusServiceUnavailable,
			wantCode:       apperrors.ErrDatabaseError,
			wantMessage:    "not ready",
			wantStatus:     "not_ready",
			wantDBCheck:    "missing",
		},
		{
			name:           "ping failure",
			db:             &fakeDatabase{pingErr: errors.New("db offline")},
			wantHTTPStatus: http.StatusServiceUnavailable,
			wantCode:       apperrors.ErrDatabaseError,
			wantMessage:    "not ready",
			wantStatus:     "not_ready",
			wantDBCheck:    "db offline",
		},
		{
			name:           "ready",
			db:             &fakeDatabase{},
			wantHTTPStatus: http.StatusOK,
			wantCode:       0,
			wantMessage:    "success",
			wantStatus:     "ready",
			wantDBCheck:    "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestRouter(RouterDeps{Database: tt.db})

			recorder, body := performRouterRequest(t, router, http.MethodGet, "/ready")

			if recorder.Code != tt.wantHTTPStatus {
				t.Fatalf("expected /ready status %d, got %d with body %s", tt.wantHTTPStatus, recorder.Code, recorder.Body.String())
			}
			if body.Code != tt.wantCode {
				t.Fatalf("expected response code %d, got %d", tt.wantCode, body.Code)
			}
			if body.Message != tt.wantMessage {
				t.Fatalf("expected response message %q, got %q", tt.wantMessage, body.Message)
			}
			if body.Data == nil {
				t.Fatal("expected response data to be present")
			}
			assertDataValue(t, body.Data, "status", tt.wantStatus)
			checks, ok := body.Data["checks"].(map[string]any)
			if !ok {
				t.Fatalf("expected data.checks to be an object, got %#v", body.Data["checks"])
			}
			assertDataValue(t, checks, "database", tt.wantDBCheck)
		})
	}
}

// TestNewRouterDoesNotRegisterRemovedUserManagementRoutes 固定 HTTP 路由、中间件顺序和错误响应契约，确保后续注释补全或结构调整不改变该场景。
func TestNewRouterDoesNotRegisterRemovedUserManagementRoutes(t *testing.T) {
	router := newTestRouter(RouterDeps{})

	for _, path := range []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/users",
		"/api/v1/roles",
		"/api/v1/permissions",
	} {
		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, path, nil)

		router.ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Fatalf("expected %s to be unregistered with status %d, got %d", path, http.StatusNotFound, recorder.Code)
		}
	}
}

// TestNewRouterServesAdminWebUI 固定管理台静态产物的挂载和 SPA fallback 行为，避免 API 路由被前端回退吞掉。
func TestNewRouterServesAdminWebUI(t *testing.T) {
	distDir := newAdminWebUIDist(t)
	router := newTestRouter(RouterDeps{
		WebUI: WebUIDeps{Enabled: true, MountPath: "/admin", DistDir: distDir},
	})

	for _, path := range []string{"/admin", "/admin/", "/admin/login", "/admin/users"} {
		recorder := performRawRouterRequest(router, http.MethodGet, path)
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected %s status %d, got %d with body %s", path, http.StatusOK, recorder.Code, recorder.Body.String())
		}
		if !strings.Contains(recorder.Body.String(), "admin-shell") {
			t.Fatalf("expected %s to serve index.html, got %q", path, recorder.Body.String())
		}
	}

	asset := performRawRouterRequest(router, http.MethodGet, "/admin/_nuxt/app.js")
	if asset.Code != http.StatusOK {
		t.Fatalf("expected static asset status %d, got %d with body %s", http.StatusOK, asset.Code, asset.Body.String())
	}
	if strings.Contains(asset.Body.String(), "admin-shell") || !strings.Contains(asset.Body.String(), "console.log") {
		t.Fatalf("expected asset response instead of index fallback, got %q", asset.Body.String())
	}

	missingAsset := performRawRouterRequest(router, http.MethodGet, "/admin/_nuxt/missing.js")
	if missingAsset.Code != http.StatusNotFound || strings.Contains(missingAsset.Body.String(), "admin-shell") {
		t.Fatalf("expected missing asset to return 404 instead of index, got status %d body %s", missingAsset.Code, missingAsset.Body.String())
	}
}

// TestNewRouterKeepsAPIAndProbesOutsideWebUI 固定 WebUI fallback 不能覆盖健康检查或 API 前缀。
func TestNewRouterKeepsAPIAndProbesOutsideWebUI(t *testing.T) {
	distDir := newAdminWebUIDist(t)
	router := newTestRouter(RouterDeps{
		WebUI: WebUIDeps{Enabled: true, MountPath: "/admin", DistDir: distDir},
	})

	health := performRawRouterRequest(router, http.MethodGet, "/health")
	if health.Code != http.StatusOK || strings.Contains(health.Body.String(), "admin-shell") {
		t.Fatalf("expected /health to stay API response, got status %d body %s", health.Code, health.Body.String())
	}

	ready := performRawRouterRequest(router, http.MethodGet, "/ready")
	if ready.Code != http.StatusServiceUnavailable || strings.Contains(ready.Body.String(), "admin-shell") {
		t.Fatalf("expected /ready to stay probe response, got status %d body %s", ready.Code, ready.Body.String())
	}

	login := performRawRouterRequest(router, http.MethodPost, "/api/v1/auth/login")
	if login.Code != http.StatusNotFound || strings.Contains(login.Body.String(), "admin-shell") {
		t.Fatalf("expected /api/v1/auth/login to stay outside SPA fallback, got status %d body %s", login.Code, login.Body.String())
	}
}

// TestNewRouterSkipsWebUIWhenDistMissing 固定缺少静态产物时后端仍可创建路由，管理台前缀返回 404。
func TestNewRouterSkipsWebUIWhenDistMissing(t *testing.T) {
	router := newTestRouter(RouterDeps{
		WebUI: WebUIDeps{Enabled: true, MountPath: "/admin", DistDir: filepath.Join(t.TempDir(), "missing")},
	})

	recorder := performRawRouterRequest(router, http.MethodGet, "/admin")
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing WebUI dist to return %d, got %d with body %s", http.StatusNotFound, recorder.Code, recorder.Body.String())
	}
}

// newTestRouter 构造当前测试场景所需的最小依赖集合，避免测试直接耦合生产装配流程。
func newTestRouter(deps RouterDeps) *web.Engine {
	if deps.Mode == "" {
		deps.Mode = "test"
	}
	return NewRouter(deps)
}

func newAdminWebUIDist(t *testing.T) string {
	t.Helper()

	distDir := t.TempDir()
	nuxtDir := filepath.Join(distDir, "_nuxt")
	if err := os.MkdirAll(nuxtDir, 0755); err != nil {
		t.Fatalf("mkdir _nuxt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(distDir, "index.html"), []byte(`<!doctype html><div id="admin-shell"></div>`), 0644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nuxtDir, "app.js"), []byte(`console.log("admin")`), 0644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}
	return distDir
}

// performRouterRequest 执行测试 HTTP 请求并返回响应记录器，封装路由调用细节。
func performRouterRequest(t *testing.T, router http.Handler, method string, path string) (*httptest.ResponseRecorder, routerResponse) {
	t.Helper()

	request := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	var body routerResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body %q: %v", recorder.Body.String(), err)
	}
	return recorder, body
}

func performRawRouterRequest(router http.Handler, method string, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// assertSuccessResponse 校验测试响应或状态中的关键字段，使测试断言聚焦在对外契约而非重复解析细节。
func assertSuccessResponse(t *testing.T, body routerResponse) {
	t.Helper()

	if body.Code != 0 {
		t.Fatalf("expected response code 0, got %d", body.Code)
	}
	if body.Message != "success" {
		t.Fatalf("expected response message success, got %q", body.Message)
	}
	if body.Data == nil {
		t.Fatal("expected response data to be present")
	}
}

// assertDataValue 校验测试响应或状态中的关键字段，使测试断言聚焦在对外契约而非重复解析细节。
func assertDataValue(t *testing.T, data map[string]any, key string, want string) {
	t.Helper()

	got, ok := data[key].(string)
	if !ok {
		t.Fatalf("expected data.%s to be a string, got %#v", key, data[key])
	}
	if got != want {
		t.Fatalf("expected data.%s %q, got %q", key, want, got)
	}
}
