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
	systemhandler "github.com/rei0721/go-scaffold/internal/modules/system/handler"
	systemmodel "github.com/rei0721/go-scaffold/internal/modules/system/model"
	systemservice "github.com/rei0721/go-scaffold/internal/modules/system/service"
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

type menuResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Code  string `json:"code"`
		Items []struct {
			Code       string `json:"code"`
			Permission string `json:"permission"`
		} `json:"items"`
	} `json:"data"`
}

type apiCatalogResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    []apiCatalogGroup `json:"data"`
}

type apiCatalogGroup struct {
	Code  string           `json:"code"`
	Items []apiCatalogItem `json:"items"`
}

type apiCatalogItem struct {
	Access     string `json:"access"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Permission string `json:"permission"`
}

type apiSyncResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Created       int               `json:"created"`
		Groups        []apiCatalogGroup `json:"groups"`
		Persisted     bool              `json:"persisted"`
		Stale         int               `json:"stale"`
		StorageStatus string            `json:"storageStatus"`
		Total         int               `json:"total"`
		Updated       int               `json:"updated"`
	} `json:"data"`
}

type apiPermissionSyncResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Created       int    `json:"created"`
		Persisted     bool   `json:"persisted"`
		Skipped       int    `json:"skipped"`
		StorageStatus string `json:"storageStatus"`
		Total         int    `json:"total"`
	} `json:"data"`
}

type fakeIAMService struct {
	setupStatusCalls  int
	initialSetupCalls int
	signupCalls       int
	loginCalls        int
	orgListCalls      int
	lastOrgFilter     iamservice.OrganizationListFilter
	userListCalls     int
	lastUserFilter    iamservice.UserListFilter
	sessionListCalls  int
	lastSessionFilter iamservice.SessionListFilter
}

type permissionAuthorizer map[string]bool

func (a permissionAuthorizer) Authorize(_ context.Context, _ iamservice.Principal, obj string, act string) (bool, error) {
	return a[obj+":"+act], nil
}

func (s *fakeIAMService) BootstrapAdmin(context.Context, iamservice.BootstrapAdminInput) (*iamservice.Principal, error) {
	return nil, nil
}

func (s *fakeIAMService) Captcha(context.Context) (iamservice.CaptchaChallenge, error) {
	return iamservice.CaptchaChallenge{Enabled: false}, nil
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
func (s *fakeIAMService) ListOrganizations(_ context.Context, _ iamservice.Principal, filter iamservice.OrganizationListFilter) (iamservice.OrganizationPage, error) {
	s.orgListCalls++
	s.lastOrgFilter = filter
	return iamservice.OrganizationPage{Items: []iammodel.Organization{}, Page: 1, PageSize: 10, StorageStatus: "persisted"}, nil
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
func (s *fakeIAMService) ListUsers(_ context.Context, _ iamservice.Principal, filter iamservice.UserListFilter) (iamservice.OrganizationUserPage, error) {
	s.userListCalls++
	s.lastUserFilter = filter
	return iamservice.OrganizationUserPage{Items: []iamservice.OrganizationUser{}, Page: 1, PageSize: 10, StorageStatus: "persisted"}, nil
}
func (s *fakeIAMService) UpdateUser(context.Context, iamservice.UpdateUserInput) (*iamservice.OrganizationUser, error) {
	return nil, nil
}
func (s *fakeIAMService) CreateAPIToken(context.Context, iamservice.CreateAPITokenInput) (iamservice.CreateAPITokenResult, error) {
	return iamservice.CreateAPITokenResult{}, nil
}
func (s *fakeIAMService) ListAPITokens(context.Context, iamservice.Principal, iamservice.APITokenFilter) (iamservice.APITokenPage, error) {
	return iamservice.APITokenPage{}, nil
}
func (s *fakeIAMService) RevokeAPIToken(context.Context, iamservice.RevokeAPITokenInput) error {
	return nil
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
func (s *fakeIAMService) ListSessions(_ context.Context, _ iamservice.Principal, filter iamservice.SessionListFilter) (iamservice.SessionPage, error) {
	s.sessionListCalls++
	s.lastSessionFilter = filter
	return iamservice.SessionPage{Items: []iammodel.Session{}, Page: 1, PageSize: 10, StorageStatus: "persisted"}, nil
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

func TestNewRouterOrgUsersPassesListFilters(t *testing.T) {
	iamSvc := &fakeIAMService{}
	router := newTestRouter(RouterDeps{
		IAMHandler: iamhandler.New(iamSvc, nil),
		IAMAuth:    iamSvc,
		IAMAuthz:   iamSvc,
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/orgs/1/users?keyword=alice&username=ali&displayName=Alice&email=alice%40example.com&roleCode=admin&status=active&page=2&pageSize=30&orderKey=username&desc=true", nil)
	request.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected users status %d, got %d body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if iamSvc.userListCalls != 1 {
		t.Fatalf("expected one ListUsers call, got %d", iamSvc.userListCalls)
	}
	filter := iamSvc.lastUserFilter
	if filter.Keyword != "alice" || filter.Username != "ali" || filter.DisplayName != "Alice" || filter.Email != "alice@example.com" || filter.RoleCode != "admin" || filter.Status != "active" || filter.Page != 2 || filter.PageSize != 30 || filter.OrderKey != "username" || !filter.Desc {
		t.Fatalf("unexpected user list filter: %#v", filter)
	}
}

func TestNewRouterOrganizationsPassesListFilters(t *testing.T) {
	iamSvc := &fakeIAMService{}
	router := newTestRouter(RouterDeps{
		IAMHandler: iamhandler.New(iamSvc, nil),
		IAMAuth:    iamSvc,
		IAMAuthz:   iamSvc,
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/orgs?keyword=team&code=alpha&name=Alpha%20Team&status=active&page=3&pageSize=20&orderKey=code&desc=1", nil)
	request.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected organizations status %d, got %d body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if iamSvc.orgListCalls != 1 {
		t.Fatalf("expected one ListOrganizations call, got %d", iamSvc.orgListCalls)
	}
	filter := iamSvc.lastOrgFilter
	if filter.Keyword != "team" || filter.Code != "alpha" || filter.Name != "Alpha Team" || filter.Status != "active" || filter.Page != 3 || filter.PageSize != 20 || filter.OrderKey != "code" || !filter.Desc {
		t.Fatalf("unexpected organization list filter: %#v", filter)
	}
}

func TestNewRouterSessionsPassesListFilters(t *testing.T) {
	iamSvc := &fakeIAMService{}
	router := newTestRouter(RouterDeps{
		IAMHandler: iamhandler.New(iamSvc, nil),
		IAMAuth:    iamSvc,
		IAMAuthz:   iamSvc,
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/orgs/1/sessions?scope=org&keyword=edge&userId=42&ipAddress=127.0.0.1&status=active&page=2&pageSize=30&orderKey=last_used_at&desc=true", nil)
	request.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected sessions status %d, got %d body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if iamSvc.sessionListCalls != 1 {
		t.Fatalf("expected one ListSessions call, got %d", iamSvc.sessionListCalls)
	}
	filter := iamSvc.lastSessionFilter
	if filter.Scope != "org" || filter.Keyword != "edge" || filter.UserID != 42 || filter.IPAddress != "127.0.0.1" || filter.Status != "active" || filter.Page != 2 || filter.PageSize != 30 || filter.OrderKey != "last_used_at" || !filter.Desc {
		t.Fatalf("unexpected session list filter: %#v", filter)
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
		"/api/v1/auth/captcha:",
		"/api/v1/orgs/{orgId}:",
		"/api/v1/orgs/{orgId}/users/{userId}:",
		"/api/v1/orgs/{orgId}/invitations:",
		"/api/v1/orgs/{orgId}/invitations/{invitationId}:",
		"/api/v1/orgs/{orgId}/roles/{roleId}:",
		"/api/v1/system/apis:",
		"/api/v1/system/apis/permissions/sync:",
		"/api/v1/system/apis/sync:",
		"/api/v1/system/config:",
		"/api/v1/system/dictionaries:",
		"/api/v1/system/menus:",
		"/api/v1/system/operation-records:",
		"/api/v1/system/parameters:",
		"/api/v1/system/parameters/{parameterId}:",
		"/api/v1/system/parameters/value:",
		"/api/v1/system/server-info:",
		"/api/v1/system/versions:",
		"/api/v1/system/versions/export:",
		"/api/v1/system/versions/import:",
		"/api/v1/system/versions/sources:",
		"/api/v1/system/versions/{versionId}:",
		"/api/v1/system/versions/{versionId}/download:",
	} {
		if !strings.Contains(spec, path) {
			t.Fatalf("openapi.yaml missing path %s", path)
		}
	}
}

func TestNewRouterSystemMenusRequireAuthAndFilterPermissions(t *testing.T) {
	auth := &fakeIAMService{}
	systemHandler := systemhandler.New(systemservice.New(systemservice.Config{DemoEnabled: true}), permissionAuthorizer{
		"audit:read":      true,
		"dictionary:read": true,
		"config:read":     true,
		"org:read":        true,
		"operation:read":  true,
		"parameter:read":  true,
		"permission:read": true,
		"role:read":       true,
		"server:read":     true,
		"version:read":    true,
	}, nil)
	router := newTestRouter(RouterDeps{
		IAMAuth:       auth,
		SystemHandler: systemHandler,
	})

	unauthorized := performRawRouterRequest(router, http.MethodGet, "/api/v1/system/menus")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected system menus to require auth, got status %d body %s", unauthorized.Code, unauthorized.Body.String())
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/system/menus", nil)
	request.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected system menus status %d, got %d body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	var body menuResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode system menu response: %v", err)
	}
	if body.Code != 0 || body.Message != "success" {
		t.Fatalf("expected success response, got %#v", body)
	}
	if !menuContains(body, "workspace", "organizations") || !menuContains(body, "workspace", "roles") {
		t.Fatalf("expected allowed workspace menus in %#v", body.Data)
	}
	if !menuContains(body, "system", "menus") || !menuContains(body, "system", "apis") {
		t.Fatalf("expected allowed system menus in %#v", body.Data)
	}
	if !menuContains(body, "system", "dictionaries") {
		t.Fatalf("expected dictionary management menu in %#v", body.Data)
	}
	if !menuContains(body, "system", "operation-records") {
		t.Fatalf("expected operation history menu in %#v", body.Data)
	}
	if !menuContains(body, "security", "login-logs") {
		t.Fatalf("expected login log menu in %#v", body.Data)
	}
	if !menuContains(body, "security", "error-logs") {
		t.Fatalf("expected error log menu in %#v", body.Data)
	}
	if !menuContains(body, "system", "parameters") {
		t.Fatalf("expected parameter management menu in %#v", body.Data)
	}
	if !menuContains(body, "system", "system-config") {
		t.Fatalf("expected system config menu in %#v", body.Data)
	}
	if !menuContains(body, "system", "server-info") {
		t.Fatalf("expected server info menu in %#v", body.Data)
	}
	if !menuContains(body, "system", "versions") {
		t.Fatalf("expected version management menu in %#v", body.Data)
	}
	if menuContains(body, "workspace", "users") {
		t.Fatalf("expected users menu to be hidden without user:read permission: %#v", body.Data)
	}
	if !menuContains(body, "examples", "todos") {
		t.Fatalf("expected demo menu when demo is enabled: %#v", body.Data)
	}
}

// TestNewRouterDoesNotRegisterRemovedUserManagementRoutes 固定 HTTP 路由、中间件顺序和错误响应契约，确保后续注释补全或结构调整不改变该场景。
func TestNewRouterSystemAPIsRequirePermissionAndListCatalog(t *testing.T) {
	auth := &fakeIAMService{}
	systemHandler := systemhandler.New(systemservice.New(systemservice.Config{}), permissionAuthorizer{
		"permission:read": true,
	}, nil)
	router := newTestRouter(RouterDeps{
		IAMAuth:       auth,
		IAMAuthz:      permissionAuthorizer{"permission:read": true},
		SystemHandler: systemHandler,
	})

	unauthorized := performRawRouterRequest(router, http.MethodGet, "/api/v1/system/apis")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected system apis to require auth, got status %d body %s", unauthorized.Code, unauthorized.Body.String())
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/system/apis", nil)
	request.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected system apis status %d, got %d body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	var body apiCatalogResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode system api response: %v", err)
	}
	if body.Code != 0 || body.Message != "success" {
		t.Fatalf("expected success response, got %#v", body)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/apis", "permission:read") {
		t.Fatalf("expected system api catalog to include itself with permission: %#v", body.Data)
	}
	if !apiCatalogAccess(body, http.MethodGet, "/api/v1/system/apis", "permission") {
		t.Fatalf("expected system api route to require permission in API catalog: %#v", body.Data)
	}
	if !apiCatalogAccess(body, http.MethodGet, "/api/v1/system/menus", "authenticated") {
		t.Fatalf("expected system menu route to require authentication in API catalog: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/config", "config:read") {
		t.Fatalf("expected system api catalog to include config with permission: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodPatch, "/api/v1/system/config", "config:update") {
		t.Fatalf("expected system api catalog to include config update with permission: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/server-info", "server:read") {
		t.Fatalf("expected system api catalog to include server info with permission: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodPost, "/api/v1/system/apis/sync", "permission:read") {
		t.Fatalf("expected system api catalog to include sync route with permission: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodPost, "/api/v1/system/apis/permissions/sync", "permission:sync") {
		t.Fatalf("expected system api catalog to include permission sync route with permission: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/dictionaries", "dictionary:read") {
		t.Fatalf("expected system api catalog to include dictionaries: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodPost, "/api/v1/system/dictionaries", "dictionary:create") {
		t.Fatalf("expected system api catalog to include dictionary create: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/operation-records", "operation:read") {
		t.Fatalf("expected system api catalog to include operation history list: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodDelete, "/api/v1/system/operation-records", "operation:delete") {
		t.Fatalf("expected system api catalog to include operation history delete: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/parameters", "parameter:read") {
		t.Fatalf("expected system api catalog to include parameter list: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodPost, "/api/v1/system/parameters", "parameter:create") {
		t.Fatalf("expected system api catalog to include parameter create: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/parameters/:parameterId", "parameter:read") {
		t.Fatalf("expected system api catalog to include parameter detail: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodPatch, "/api/v1/system/parameters/:parameterId", "parameter:update") {
		t.Fatalf("expected system api catalog to include parameter update: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodDelete, "/api/v1/system/parameters/:parameterId", "parameter:delete") {
		t.Fatalf("expected system api catalog to include parameter delete: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/menus", "") {
		t.Fatalf("expected system api catalog to include menus: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/versions", "version:read") {
		t.Fatalf("expected system api catalog to include versions list: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodPost, "/api/v1/system/versions/export", "version:create") {
		t.Fatalf("expected system api catalog to include version export: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodPost, "/api/v1/system/versions/import", "version:import") {
		t.Fatalf("expected system api catalog to include version import: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodGet, "/api/v1/system/versions/:versionId/download", "version:download") {
		t.Fatalf("expected system api catalog to include version download: %#v", body.Data)
	}
	if !apiCatalogContains(body, http.MethodDelete, "/api/v1/system/versions/:versionId", "version:delete") {
		t.Fatalf("expected system api catalog to include version delete: %#v", body.Data)
	}
}

func TestAPIRouteAccessClassifiesPublicAuthenticatedAndPermissionRoutes(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		permission string
		want       string
	}{
		{name: "login is public", path: "/api/v1/auth/login", want: "public"},
		{name: "captcha is public", path: "/api/v1/auth/captcha", want: "public"},
		{name: "invitation accept is public", path: "/api/v1/invitations/:token/accept", want: "public"},
		{name: "logout is authenticated", path: "/api/v1/auth/logout", want: "authenticated"},
		{name: "me is authenticated", path: "/api/v1/me", want: "authenticated"},
		{name: "permission route wins", path: "/api/v1/system/apis", permission: "permission:read", want: "permission"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := apiRouteAccess(tt.path, tt.permission); got != tt.want {
				t.Fatalf("apiRouteAccess(%q, %q) = %q, want %q", tt.path, tt.permission, got, tt.want)
			}
		})
	}
}

func TestSanitizeOperationRequestBodyRedactsConfigUpdateValues(t *testing.T) {
	body := sanitizeOperationRequestBody(http.MethodPatch, "/api/v1/system/config", `{"persist":true,"items":[{"key":"database.password","value":"secret"},{"key":"server.port","value":18083}]}`)

	if strings.Contains(body, "secret") || strings.Contains(body, "18083") {
		t.Fatalf("expected config update body to be redacted, got %s", body)
	}
	if !strings.Contains(body, "database.password") || !strings.Contains(body, "[redacted]") {
		t.Fatalf("expected config update body to retain keys and redact values, got %s", body)
	}
	if !strings.Contains(body, `"persist":true`) {
		t.Fatalf("expected config update body to retain persist flag, got %s", body)
	}
}

func TestNewRouterSystemConfigUpdateRequiresPermission(t *testing.T) {
	auth := &fakeIAMService{}
	systemHandler := systemhandler.New(systemservice.New(systemservice.Config{
		ConfigUpdater: func(context.Context, systemservice.UpdateConfigInput) (systemmodel.ConfigSnapshot, error) {
			return systemmodel.ConfigSnapshot{Sections: []systemmodel.ConfigSection{
				{
					Code: "server",
					Items: []systemmodel.ConfigItem{
						{Key: "server.port", Value: 18083},
					},
				},
			}}, nil
		},
	}), nil, nil)

	routerWithoutPermission := newTestRouter(RouterDeps{
		IAMAuth:       auth,
		IAMAuthz:      permissionAuthorizer{},
		SystemHandler: systemHandler,
	})
	forbiddenRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/system/config", bytes.NewBufferString(`{"items":[{"key":"server.port","value":18083}]}`))
	forbiddenRequest.Header.Set("Authorization", "Bearer token")
	forbiddenRequest.Header.Set("Content-Type", "application/json")
	forbiddenRecorder := httptest.NewRecorder()
	routerWithoutPermission.ServeHTTP(forbiddenRecorder, forbiddenRequest)
	if forbiddenRecorder.Code != http.StatusForbidden {
		t.Fatalf("expected config update to require config:update, got status %d body %s", forbiddenRecorder.Code, forbiddenRecorder.Body.String())
	}

	routerWithPermission := newTestRouter(RouterDeps{
		IAMAuth:       auth,
		IAMAuthz:      permissionAuthorizer{"config:update": true},
		SystemHandler: systemHandler,
	})
	allowedRequest := httptest.NewRequest(http.MethodPatch, "/api/v1/system/config", bytes.NewBufferString(`{"items":[{"key":"server.port","value":18083}]}`))
	allowedRequest.Header.Set("Authorization", "Bearer token")
	allowedRequest.Header.Set("Content-Type", "application/json")
	allowedRecorder := httptest.NewRecorder()
	routerWithPermission.ServeHTTP(allowedRecorder, allowedRequest)
	if allowedRecorder.Code != http.StatusOK {
		t.Fatalf("expected config update status %d, got %d body %s", http.StatusOK, allowedRecorder.Code, allowedRecorder.Body.String())
	}
}

func TestNewRouterSystemAPISyncReturnsLiveCatalogWithoutStorage(t *testing.T) {
	auth := &fakeIAMService{}
	systemHandler := systemhandler.New(systemservice.New(systemservice.Config{}), permissionAuthorizer{
		"permission:read": true,
	}, nil)
	router := newTestRouter(RouterDeps{
		IAMAuth:       auth,
		IAMAuthz:      permissionAuthorizer{"permission:read": true},
		SystemHandler: systemHandler,
	})

	unauthorized := performRawRouterRequest(router, http.MethodPost, "/api/v1/system/apis/sync")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected system api sync to require auth, got status %d body %s", unauthorized.Code, unauthorized.Body.String())
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/system/apis/sync", nil)
	request.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected system api sync status %d, got %d body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	var body apiSyncResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode system api sync response: %v", err)
	}
	if body.Code != 0 || body.Message != "success" {
		t.Fatalf("expected success response, got %#v", body)
	}
	if body.Data.Persisted || body.Data.StorageStatus != "memory" {
		t.Fatalf("expected in-memory sync result without repository, got %#v", body.Data)
	}
	if body.Data.Total == 0 || !apiGroupsContain(body.Data.Groups, http.MethodPost, "/api/v1/system/apis/sync", "permission:read") {
		t.Fatalf("expected sync result to include live catalog routes, got %#v", body.Data.Groups)
	}
}

func TestNewRouterSystemAPIPermissionSyncReturnsUnavailableWithoutStore(t *testing.T) {
	auth := &fakeIAMService{}
	systemHandler := systemhandler.New(systemservice.New(systemservice.Config{}), permissionAuthorizer{
		"permission:read": true,
		"permission:sync": true,
	}, nil)
	router := newTestRouter(RouterDeps{
		IAMAuth:       auth,
		IAMAuthz:      permissionAuthorizer{"permission:sync": true},
		SystemHandler: systemHandler,
	})

	unauthorized := performRawRouterRequest(router, http.MethodPost, "/api/v1/system/apis/permissions/sync")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected system api permission sync to require auth, got status %d body %s", unauthorized.Code, unauthorized.Body.String())
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/system/apis/permissions/sync", nil)
	request.Header.Set("Authorization", "Bearer token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected system api permission sync status %d, got %d body %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	var body apiPermissionSyncResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode system api permission sync response: %v", err)
	}
	if body.Code != 0 || body.Message != "success" {
		t.Fatalf("expected success response, got %#v", body)
	}
	if body.Data.Persisted || body.Data.StorageStatus != "unavailable" {
		t.Fatalf("expected unavailable permission sync without store, got %#v", body.Data)
	}
	if body.Data.Total == 0 {
		t.Fatalf("expected permission sync result to count route permissions, got %#v", body.Data)
	}
}

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

// TestNewRouterReturns404WhenWebUIDistMissing 固定缺少静态产物时后端仍可创建路由，管理台前缀返回 404。
func TestNewRouterReturns404WhenWebUIDistMissing(t *testing.T) {
	router := newTestRouter(RouterDeps{
		WebUI: WebUIDeps{Enabled: true, MountPath: "/admin", DistDir: filepath.Join(t.TempDir(), "missing")},
	})

	recorder := performRawRouterRequest(router, http.MethodGet, "/admin")
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected missing WebUI dist to return %d, got %d with body %s", http.StatusNotFound, recorder.Code, recorder.Body.String())
	}
}

func TestNewRouterServesAdminWebUIAfterLateStaticGeneration(t *testing.T) {
	distDir := t.TempDir()
	router := newTestRouter(RouterDeps{
		WebUI: WebUIDeps{Enabled: true, MountPath: "/admin", DistDir: distDir},
	})

	before := performRawRouterRequest(router, http.MethodGet, "/admin")
	if before.Code != http.StatusNotFound {
		t.Fatalf("expected missing WebUI dist to return %d before generation, got %d with body %s", http.StatusNotFound, before.Code, before.Body.String())
	}

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

	after := performRawRouterRequest(router, http.MethodGet, "/admin/login")
	if after.Code != http.StatusOK {
		t.Fatalf("expected late generated WebUI status %d, got %d with body %s", http.StatusOK, after.Code, after.Body.String())
	}
	if !strings.Contains(after.Body.String(), "admin-shell") {
		t.Fatalf("expected late generated WebUI to serve index.html, got %q", after.Body.String())
	}
	asset := performRawRouterRequest(router, http.MethodGet, "/admin/_nuxt/app.js")
	if asset.Code != http.StatusOK || !strings.Contains(asset.Body.String(), "console.log") {
		t.Fatalf("expected late generated asset response, got status %d body %s", asset.Code, asset.Body.String())
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

func menuContains(body menuResponse, groupCode string, itemCode string) bool {
	for _, group := range body.Data {
		if group.Code != groupCode {
			continue
		}
		for _, item := range group.Items {
			if item.Code == itemCode {
				return true
			}
		}
	}
	return false
}

// assertSuccessResponse 校验测试响应或状态中的关键字段，使测试断言聚焦在对外契约而非重复解析细节。
func apiCatalogContains(body apiCatalogResponse, method string, path string, permission string) bool {
	return apiGroupsContain(body.Data, method, path, permission)
}

func apiGroupsContain(groups []apiCatalogGroup, method string, path string, permission string) bool {
	for _, group := range groups {
		for _, item := range group.Items {
			if item.Method == method && item.Path == path && item.Permission == permission {
				return true
			}
		}
	}
	return false
}

func apiCatalogAccess(body apiCatalogResponse, method string, path string, access string) bool {
	for _, group := range body.Data {
		for _, item := range group.Items {
			if item.Method == method && item.Path == path && item.Access == access {
				return true
			}
		}
	}
	return false
}

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
