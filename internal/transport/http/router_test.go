package httptransport

// 本测试文件固定 HTTP 路由、中间件顺序和错误响应契约，防止注释补全和后续重构改变外部可观察行为。

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	iamhandler "github.com/rei0721/go-scaffold/internal/modules/iam/handler"
	iammodel "github.com/rei0721/go-scaffold/internal/modules/iam/model"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
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

type fakeIAMService struct {
	setupStatusCalls  int
	initialSetupCalls int
	signupCalls       int
	loginCalls        int
}

func (s *fakeIAMService) BootstrapAdmin(context.Context, iamservice.BootstrapAdminInput) (*iamservice.Principal, error) {
	return nil, nil
}
func (s *fakeIAMService) SetupStatus(context.Context) (iamservice.SetupStatus, error) {
	s.setupStatusCalls++
	return iamservice.SetupStatus{Required: true}, nil
}
func (s *fakeIAMService) InitialAdminSetup(context.Context, iamservice.InitialAdminSetupInput) (iamservice.TokenPair, error) {
	s.initialSetupCalls++
	return iamservice.TokenPair{AccessToken: "access", RefreshToken: "refresh", AccessExpiresAt: time.Now().Add(time.Hour), RefreshExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (s *fakeIAMService) Signup(context.Context, iamservice.SignupInput) (iamservice.TokenPair, error) {
	s.signupCalls++
	return iamservice.TokenPair{AccessToken: "access", RefreshToken: "refresh", AccessExpiresAt: time.Now().Add(time.Hour), RefreshExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (s *fakeIAMService) Login(context.Context, iamservice.LoginInput) (iamservice.TokenPair, error) {
	s.loginCalls++
	return iamservice.TokenPair{AccessToken: "access", RefreshToken: "refresh", AccessExpiresAt: time.Now().Add(time.Hour), RefreshExpiresAt: time.Now().Add(time.Hour)}, nil
}
func (s *fakeIAMService) Refresh(context.Context, iamservice.RefreshInput) (iamservice.TokenPair, error) {
	return iamservice.TokenPair{}, nil
}
func (s *fakeIAMService) Logout(context.Context, iamservice.Principal) error { return nil }
func (s *fakeIAMService) SwitchOrg(context.Context, iamservice.Principal, int64, string, string) (iamservice.TokenPair, error) {
	return iamservice.TokenPair{}, nil
}
func (s *fakeIAMService) AuthenticateToken(context.Context, string) (iamservice.Principal, error) {
	return iamservice.Principal{UserID: 1, OrgID: 1, SessionID: 1, Username: "admin", Email: "admin@example.com"}, nil
}
func (s *fakeIAMService) Authorize(context.Context, iamservice.Principal, string, string) (bool, error) {
	return true, nil
}
func (s *fakeIAMService) Me(context.Context, iamservice.Principal) (*iammodel.User, error) {
	return nil, nil
}
func (s *fakeIAMService) ListMyOrganizations(context.Context, iamservice.Principal) ([]iammodel.Organization, error) {
	return nil, nil
}
func (s *fakeIAMService) ListOrganizations(context.Context, iamservice.Principal) ([]iammodel.Organization, error) {
	return nil, nil
}
func (s *fakeIAMService) CreateOrganization(context.Context, iamservice.Principal, string, string) (*iammodel.Organization, error) {
	return nil, nil
}
func (s *fakeIAMService) UpdateOrganization(context.Context, iamservice.UpdateOrganizationInput) (*iammodel.Organization, error) {
	return nil, nil
}
func (s *fakeIAMService) InviteUser(context.Context, iamservice.InviteUserInput) (iamservice.NotificationDelivery, error) {
	return iamservice.NotificationDelivery{}, nil
}
func (s *fakeIAMService) ListInvitations(context.Context, iamservice.Principal) ([]iammodel.Invitation, error) {
	return nil, nil
}
func (s *fakeIAMService) RevokeInvitation(context.Context, iamservice.Principal, int64, string, string) error {
	return nil
}
func (s *fakeIAMService) AcceptInvitation(context.Context, iamservice.AcceptInvitationInput) (*iamservice.Principal, error) {
	return nil, nil
}
func (s *fakeIAMService) ForgotPassword(context.Context, iamservice.ForgotPasswordInput) (iamservice.NotificationDelivery, error) {
	return iamservice.NotificationDelivery{}, nil
}
func (s *fakeIAMService) ResetPassword(context.Context, iamservice.ResetPasswordInput) error {
	return nil
}
func (s *fakeIAMService) SetupMFA(context.Context, iamservice.Principal) (string, string, error) {
	return "", "", nil
}
func (s *fakeIAMService) VerifyMFA(context.Context, iamservice.Principal, string) error { return nil }
func (s *fakeIAMService) ListUsers(context.Context, iamservice.Principal) ([]iamservice.OrganizationUser, error) {
	return nil, nil
}
func (s *fakeIAMService) UpdateUser(context.Context, iamservice.UpdateUserInput) (*iamservice.OrganizationUser, error) {
	return nil, nil
}
func (s *fakeIAMService) ListRoles(context.Context, iamservice.Principal) ([]iammodel.Role, error) {
	return nil, nil
}
func (s *fakeIAMService) CreateRole(context.Context, iamservice.CreateRoleInput) (*iammodel.Role, error) {
	return nil, nil
}
func (s *fakeIAMService) UpdateRole(context.Context, iamservice.UpdateRoleInput) (*iammodel.Role, error) {
	return nil, nil
}
func (s *fakeIAMService) ListPermissions(context.Context, iamservice.Principal) ([]iammodel.Permission, error) {
	return nil, nil
}
func (s *fakeIAMService) ListSessions(context.Context, iamservice.Principal, int64) ([]iammodel.Session, error) {
	return nil, nil
}
func (s *fakeIAMService) RevokeSession(context.Context, iamservice.Principal, int64) error {
	return nil
}
func (s *fakeIAMService) ListAuditLogs(context.Context, iamservice.Principal, iamservice.AuditLogFilter) ([]iammodel.AuditLog, error) {
	return nil, nil
}
func (s *fakeIAMService) RecordAudit(context.Context, iamservice.Principal, string, string, string, string, string, map[string]any) error {
	return nil
}
func (s *fakeIAMService) LoadPolicies(context.Context) error { return nil }

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

func TestNewRouterSignupEndpointIsPublic(t *testing.T) {
	iamSvc := &fakeIAMService{}
	router := newTestRouter(RouterDeps{
		IAMHandler: iamhandler.New(iamSvc, nil),
		IAMAuth:    iamSvc,
		IAMAuthz:   iamSvc,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/signup", bytes.NewBufferString(`{"orgCode":"acme","orgName":"Acme","username":"owner","email":"owner@example.com","password":"password123"}`))
	request.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected public signup status %d, got %d body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if iamSvc.signupCalls != 1 {
		t.Fatalf("expected signup call count 1, got %d", iamSvc.signupCalls)
	}
}

func TestNewRouterSetupEndpointsArePublic(t *testing.T) {
	iamSvc := &fakeIAMService{}
	router := newTestRouter(RouterDeps{
		IAMHandler: iamhandler.New(iamSvc, nil),
		IAMAuth:    iamSvc,
		IAMAuthz:   iamSvc,
	})

	statusRecorder := httptest.NewRecorder()
	statusRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/setup/status", nil)
	router.ServeHTTP(statusRecorder, statusRequest)
	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("expected setup status %d, got %d body %s", http.StatusOK, statusRecorder.Code, statusRecorder.Body.String())
	}

	setupRecorder := performJSONRouterRequest(router, http.MethodPost, "/api/v1/auth/setup/initial-admin", `{"orgCode":"acme","orgName":"Acme","username":"owner","email":"owner@example.com","password":"password123"}`)
	if setupRecorder.Code != http.StatusOK {
		t.Fatalf("expected setup initial-admin %d, got %d body %s", http.StatusOK, setupRecorder.Code, setupRecorder.Body.String())
	}
	if iamSvc.setupStatusCalls != 1 || iamSvc.initialSetupCalls != 1 {
		t.Fatalf("unexpected setup call counts: status=%d initial=%d", iamSvc.setupStatusCalls, iamSvc.initialSetupCalls)
	}
}

func TestNewRouterRateLimitsPublicAuthEndpoints(t *testing.T) {
	iamSvc := &fakeIAMService{}
	router := newTestRouter(RouterDeps{
		IAMHandler: iamhandler.New(iamSvc, nil),
		IAMAuth:    iamSvc,
		IAMAuthz:   iamSvc,
	})

	for i := 0; i < 20; i++ {
		recorder := performJSONRouterRequest(router, http.MethodPost, "/api/v1/auth/login", `{"identifier":"owner@example.com","password":"password123"}`)
		if recorder.Code != http.StatusOK {
			t.Fatalf("request %d expected status %d, got %d body %s", i+1, http.StatusOK, recorder.Code, recorder.Body.String())
		}
	}
	limited := performJSONRouterRequest(router, http.MethodPost, "/api/v1/auth/login", `{"identifier":"owner@example.com","password":"password123"}`)
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("expected rate limited status %d, got %d body %s", http.StatusTooManyRequests, limited.Code, limited.Body.String())
	}
}

func TestOpenAPICoversIAMProductRoutes(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "api", "openapi.yaml"))
	if err != nil {
		t.Fatalf("read openapi.yaml: %v", err)
	}
	spec := string(raw)
	for _, path := range []string{
		"/api/v1/auth/setup/status:",
		"/api/v1/auth/setup/initial-admin:",
		"/api/v1/auth/signup:",
		"/api/v1/orgs/{orgId}:",
		"/api/v1/orgs/{orgId}/users/{userId}:",
		"/api/v1/orgs/{orgId}/invitations:",
		"/api/v1/orgs/{orgId}/invitations/{invitationId}:",
		"/api/v1/orgs/{orgId}/roles/{roleId}:",
	} {
		if !strings.Contains(spec, path) {
			t.Fatalf("openapi.yaml missing path %s", path)
		}
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

func performJSONRouterRequest(router http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
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
