package httptransport

// 本文件定义 HTTP 传输层装配，把中间件顺序、健康检查和业务路由注册为 Gin Engine。

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/internal/middleware"
	demohandler "github.com/rei0721/go-scaffold/internal/modules/demo/handler"
	iamhandler "github.com/rei0721/go-scaffold/internal/modules/iam/handler"
	pluginhandler "github.com/rei0721/go-scaffold/internal/modules/plugins/handler"
	systemhandler "github.com/rei0721/go-scaffold/internal/modules/system/handler"
	systemmodel "github.com/rei0721/go-scaffold/internal/modules/system/model"
	systemservice "github.com/rei0721/go-scaffold/internal/modules/system/service"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/i18n"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/utils"
	"github.com/rei0721/go-scaffold/pkg/web"
	apperrors "github.com/rei0721/go-scaffold/types/errors"
	"github.com/rei0721/go-scaffold/types/result"
)

// RouterDeps 聚合 HTTP 路由装配所需依赖，允许测试或可选模块传入 nil 以裁剪路由。
type RouterDeps struct {
	Mode          string
	Logger        logger.Logger
	I18n          i18n.I18n
	Database      database.Database
	Middleware    middleware.MiddlewareConfig
	TodoHandler   *demohandler.TodoHandler
	IAMHandler    *iamhandler.Handler
	PluginHandler *pluginhandler.Handler
	SystemHandler *systemhandler.Handler
	IAMAuth       middleware.Authenticator
	IAMAuthz      middleware.Authorizer
	WebUI         WebUIDeps
}

// WebUIDeps 描述管理台静态产物挂载所需配置，避免 transport 层直接依赖应用配置结构。
type WebUIDeps struct {
	Enabled   bool
	MountPath string
	DistDir   string
}

// NewRouter 按固定顺序注册中间件、健康检查和业务路由，返回可直接交给 HTTPServer 的 Gin Engine。
func NewRouter(deps RouterDeps) *web.Engine {
	r := web.New(deps.Mode)

	if deps.I18n != nil {
		r.Use(middleware.I18n(deps.I18n))
	}
	r.Use(middleware.TraceID(deps.Middleware.TraceID))
	r.Use(middleware.CORSMiddleware(deps.Middleware.CORS))
	if deps.Logger != nil {
		r.Use(middleware.Logger(deps.Middleware.Logger, deps.Logger))
		r.Use(middleware.Recovery(deps.Middleware.Recovery, deps.Logger))
	} else {
		r.Use(web.Recovery())
	}

	r.GET("/health", health)
	r.GET("/ready", ready(deps.Database))

	v1 := r.Group("/api/v1")
	demo := v1.Group("/demo")
	if deps.TodoHandler != nil {
		todos := demo.Group("/todos")
		todos.POST("", deps.TodoHandler.Create)
		todos.GET("", deps.TodoHandler.List)
		todos.GET("/:id", deps.TodoHandler.Get)
		todos.PUT("/:id", deps.TodoHandler.Update)
		todos.DELETE("/:id", deps.TodoHandler.Delete)
	}
	if deps.IAMHandler != nil {
		registerIAMRoutes(v1, deps)
	}
	if deps.PluginHandler != nil {
		registerPluginRoutes(v1, deps)
	}
	if deps.SystemHandler != nil {
		registerSystemRoutes(v1, deps)
		deps.SystemHandler.RegisterAPIs(catalogAPIRoutes(r.Routes()))
	}
	registerWebUI(r, deps)

	return r
}

func registerWebUI(r *web.Engine, deps RouterDeps) {
	if !deps.WebUI.Enabled {
		return
	}
	err := r.MountStaticSPA(web.StaticSPAConfig{
		MountPath: deps.WebUI.MountPath,
		DistDir:   deps.WebUI.DistDir,
	})
	if err == nil {
		if deps.Logger != nil {
			deps.Logger.Info("admin webui mounted", "mount_path", deps.WebUI.MountPath, "dist_dir", deps.WebUI.DistDir)
		}
		return
	}
	if deps.Logger == nil {
		return
	}
	if errors.Is(err, web.ErrStaticSPAIndexMissing) {
		deps.Logger.Warn("admin webui static files missing", "mount_path", deps.WebUI.MountPath, "dist_dir", deps.WebUI.DistDir)
		return
	}
	deps.Logger.Warn("admin webui mount skipped", "mount_path", deps.WebUI.MountPath, "dist_dir", deps.WebUI.DistDir, "error", err)
}

func registerIAMRoutes(v1 web.Router, deps RouterDeps) {
	auth := v1.Group("/auth")
	auth.Use(middleware.RateLimit(middleware.RateLimitConfig{Enabled: true, Limit: 20, Window: time.Minute}))
	auth.GET("/setup/status", deps.IAMHandler.SetupStatus)
	auth.POST("/setup/initial-admin", deps.IAMHandler.InitialAdminSetup)
	auth.POST("/signup", deps.IAMHandler.Signup)
	auth.POST("/login", deps.IAMHandler.Login)
	auth.POST("/refresh", deps.IAMHandler.Refresh)
	auth.POST("/password/forgot", deps.IAMHandler.ForgotPassword)
	auth.POST("/password/reset", deps.IAMHandler.ResetPassword)

	invitations := v1.Group("/invitations")
	invitations.Use(middleware.RateLimit(middleware.RateLimitConfig{Enabled: true, Limit: 20, Window: time.Minute}))
	invitations.POST("/:token/accept", deps.IAMHandler.AcceptInvitation)

	protected := v1.Group("")
	protected.Use(middleware.Auth(deps.IAMAuth))
	protected.Use(OperationRecorder(deps.SystemHandler))
	protected.POST("/auth/logout", deps.IAMHandler.Logout)
	protected.POST("/auth/switch-org", deps.IAMHandler.SwitchOrg)
	protected.POST("/auth/mfa/setup", deps.IAMHandler.SetupMFA)
	protected.POST("/auth/mfa/verify", deps.IAMHandler.VerifyMFA)
	protected.GET("/me", deps.IAMHandler.Me)
	protected.GET("/me/orgs", deps.IAMHandler.MyOrganizations)

	orgs := protected.Group("/orgs")
	orgScoped := func(obj, act string, next web.HandlerFunc) web.HandlerFunc {
		return middleware.RequireOrgParam("orgId", middleware.RequirePermission(deps.IAMAuthz, obj, act, next))
	}
	orgs.GET("", middleware.RequirePermission(deps.IAMAuthz, "org", "read", deps.IAMHandler.ListOrganizations))
	orgs.POST("", middleware.RequirePermission(deps.IAMAuthz, "org", "create", deps.IAMHandler.CreateOrganization))
	orgs.PATCH("/:orgId", orgScoped("org", "update", deps.IAMHandler.UpdateOrganization))
	orgs.GET("/:orgId/users", orgScoped("user", "read", deps.IAMHandler.ListUsers))
	orgs.PATCH("/:orgId/users/:userId", orgScoped("user", "update", deps.IAMHandler.UpdateUser))
	orgs.POST("/:orgId/users/invitations", orgScoped("user", "invite", deps.IAMHandler.InviteUser))
	orgs.GET("/:orgId/invitations", orgScoped("user", "invite", deps.IAMHandler.ListInvitations))
	orgs.DELETE("/:orgId/invitations/:invitationId", orgScoped("user", "invite", deps.IAMHandler.RevokeInvitation))
	orgs.GET("/:orgId/roles", orgScoped("role", "read", deps.IAMHandler.ListRoles))
	orgs.POST("/:orgId/roles", orgScoped("role", "create", deps.IAMHandler.CreateRole))
	orgs.PATCH("/:orgId/roles/:roleId", orgScoped("role", "update", deps.IAMHandler.UpdateRole))
	orgs.GET("/:orgId/permissions", orgScoped("permission", "read", deps.IAMHandler.ListPermissions))
	orgs.GET("/:orgId/sessions", orgScoped("session", "read", deps.IAMHandler.ListSessions))
	orgs.DELETE("/:orgId/sessions/:sessionId", orgScoped("session", "revoke", deps.IAMHandler.RevokeSession))
	orgs.GET("/:orgId/audit-logs", orgScoped("audit", "read", deps.IAMHandler.ListAuditLogs))
}

func registerPluginRoutes(v1 web.Router, deps RouterDeps) {
	plugins := v1.Group("/plugins")
	plugins.Use(middleware.Auth(deps.IAMAuth))
	plugins.Use(OperationRecorder(deps.SystemHandler))
	plugins.GET("", middleware.RequirePermission(deps.IAMAuthz, "plugin", "read", deps.PluginHandler.List))
	plugins.GET("/:pluginId", middleware.RequirePermission(deps.IAMAuthz, "plugin", "read", deps.PluginHandler.Get))
	plugins.GET("/:pluginId/health", middleware.RequirePermission(deps.IAMAuthz, "plugin", "read", deps.PluginHandler.Health))
	plugins.ANY("/:pluginId/proxy/*path", middleware.RequirePermission(deps.IAMAuthz, "plugin", "proxy", deps.PluginHandler.Proxy))
}

func registerSystemRoutes(v1 web.Router, deps RouterDeps) {
	system := v1.Group("/system")
	system.Use(middleware.Auth(deps.IAMAuth))
	system.Use(OperationRecorder(deps.SystemHandler))
	system.GET("/menus", deps.SystemHandler.ListMenus)
	system.GET("/config", middleware.RequirePermission(deps.IAMAuthz, "config", "read", deps.SystemHandler.ListConfig))
	system.GET("/server-info", middleware.RequirePermission(deps.IAMAuthz, "server", "read", deps.SystemHandler.GetServerInfo))
	system.GET("/apis", middleware.RequirePermission(deps.IAMAuthz, "permission", "read", deps.SystemHandler.ListAPIs))
	system.POST("/apis/sync", middleware.RequirePermission(deps.IAMAuthz, "permission", "read", deps.SystemHandler.SyncAPIs))
	system.POST("/apis/permissions/sync", middleware.RequirePermission(deps.IAMAuthz, "permission", "sync", deps.SystemHandler.SyncPermissions))
	system.GET("/operation-records", middleware.RequirePermission(deps.IAMAuthz, "operation", "read", deps.SystemHandler.ListOperationRecords))
	system.DELETE("/operation-records", middleware.RequirePermission(deps.IAMAuthz, "operation", "delete", deps.SystemHandler.DeleteOperationRecords))
	system.GET("/parameters", middleware.RequirePermission(deps.IAMAuthz, "parameter", "read", deps.SystemHandler.ListParameters))
	system.POST("/parameters", middleware.RequirePermission(deps.IAMAuthz, "parameter", "create", deps.SystemHandler.CreateParameter))
	system.DELETE("/parameters", middleware.RequirePermission(deps.IAMAuthz, "parameter", "delete", deps.SystemHandler.DeleteParameters))
	system.GET("/parameters/value", middleware.RequirePermission(deps.IAMAuthz, "parameter", "read", deps.SystemHandler.GetParameterByKey))
	system.GET("/parameters/:parameterId", middleware.RequirePermission(deps.IAMAuthz, "parameter", "read", deps.SystemHandler.GetParameter))
	system.PATCH("/parameters/:parameterId", middleware.RequirePermission(deps.IAMAuthz, "parameter", "update", deps.SystemHandler.UpdateParameter))
	system.DELETE("/parameters/:parameterId", middleware.RequirePermission(deps.IAMAuthz, "parameter", "delete", deps.SystemHandler.DeleteParameter))
	system.GET("/dictionaries", middleware.RequirePermission(deps.IAMAuthz, "dictionary", "read", deps.SystemHandler.ListDictionaries))
	system.POST("/dictionaries", middleware.RequirePermission(deps.IAMAuthz, "dictionary", "create", deps.SystemHandler.CreateDictionary))
	system.PATCH("/dictionaries/:dictionaryId", middleware.RequirePermission(deps.IAMAuthz, "dictionary", "update", deps.SystemHandler.UpdateDictionary))
	system.DELETE("/dictionaries/:dictionaryId", middleware.RequirePermission(deps.IAMAuthz, "dictionary", "delete", deps.SystemHandler.DeleteDictionary))
	system.POST("/dictionaries/:dictionaryId/items", middleware.RequirePermission(deps.IAMAuthz, "dictionary", "update", deps.SystemHandler.CreateDictionaryItem))
	system.PATCH("/dictionary-items/:itemId", middleware.RequirePermission(deps.IAMAuthz, "dictionary", "update", deps.SystemHandler.UpdateDictionaryItem))
	system.DELETE("/dictionary-items/:itemId", middleware.RequirePermission(deps.IAMAuthz, "dictionary", "delete", deps.SystemHandler.DeleteDictionaryItem))
}

type operationRecorder interface {
	RecordOperation(context.Context, systemservice.OperationRecordInput) error
}

func OperationRecorder(recorder operationRecorder) web.HandlerFunc {
	return func(c web.Context) {
		if recorder == nil || !strings.HasPrefix(c.Path(), "/api/v1/") {
			c.Next()
			return
		}
		body := readRequestBody(c.Request())
		start := time.Now()
		c.Next()

		status := c.Status()
		if status == 0 {
			status = http.StatusOK
		}
		principal, _ := middleware.GetPrincipal(c)
		input := systemservice.OperationRecordInput{
			Body:      body,
			IPAddress: utils.ClientIPRealIP(c),
			LatencyMs: time.Since(start).Milliseconds(),
			Method:    c.Method(),
			Path:      c.Path(),
			Status:    status,
			TraceID:   middleware.GetTraceID(c),
			UserAgent: c.GetHeader("User-Agent"),
			UserID:    principal.UserID,
			Username:  principal.Username,
		}
		if status >= http.StatusBadRequest {
			input.ErrorMessage = http.StatusText(status)
		}
		_ = recorder.RecordOperation(context.Background(), input)
	}
}

func readRequestBody(req *http.Request) string {
	if req == nil || req.Body == nil {
		return ""
	}
	raw, err := io.ReadAll(req.Body)
	if err != nil {
		req.Body = io.NopCloser(bytes.NewReader(nil))
		return ""
	}
	req.Body = io.NopCloser(bytes.NewReader(raw))
	return string(raw)
}

func catalogAPIRoutes(routes []web.RouteInfo) []systemmodel.APIEntry {
	entries := make([]systemmodel.APIEntry, 0, len(routes))
	for _, route := range routes {
		if !strings.HasPrefix(route.Path, "/api/v1/") {
			continue
		}
		permission := apiRoutePermission(route.Method, route.Path)
		entries = append(entries, systemmodel.APIEntry{
			Code:        strings.ToLower(route.Method + " " + route.Path),
			Group:       apiRouteGroup(route.Path),
			Method:      route.Method,
			Path:        route.Path,
			Description: apiRouteDescription(route.Method, route.Path, permission),
			Permission:  permission,
			Order:       apiRouteMethodOrder(route.Method),
		})
	}
	return entries
}

func apiRouteGroup(path string) string {
	path = strings.TrimPrefix(path, "/api/v1/")
	segment, _, _ := strings.Cut(path, "/")
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return "other"
	}
	return segment
}

func apiRouteDescription(method string, path string, permission string) string {
	if strings.HasPrefix(path, "/api/v1/auth/") || strings.HasPrefix(path, "/api/v1/invitations/") {
		return "认证流程接口"
	}
	if permission != "" {
		return "权限保护接口：" + permission
	}
	if strings.HasPrefix(path, "/api/v1/system/menus") {
		return "当前用户可见菜单"
	}
	return method + " " + path
}

func apiRoutePermission(method string, path string) string {
	switch {
	case strings.HasPrefix(path, "/api/v1/plugins/") && strings.Contains(path, "/proxy/"):
		return "plugin:proxy"
	case strings.HasPrefix(path, "/api/v1/plugins"):
		return "plugin:read"
	case path == "/api/v1/system/config":
		return "config:read"
	case path == "/api/v1/system/server-info":
		return "server:read"
	case path == "/api/v1/system/apis":
		return "permission:read"
	case path == "/api/v1/system/apis/sync":
		return "permission:read"
	case path == "/api/v1/system/apis/permissions/sync":
		return "permission:sync"
	case path == "/api/v1/system/operation-records" && method == http.MethodDelete:
		return "operation:delete"
	case path == "/api/v1/system/operation-records":
		return "operation:read"
	case path == "/api/v1/system/parameters" && method == http.MethodGet:
		return "parameter:read"
	case path == "/api/v1/system/parameters" && method == http.MethodPost:
		return "parameter:create"
	case path == "/api/v1/system/parameters" && method == http.MethodDelete:
		return "parameter:delete"
	case path == "/api/v1/system/parameters/value":
		return "parameter:read"
	case strings.HasPrefix(path, "/api/v1/system/parameters/") && method == http.MethodGet:
		return "parameter:read"
	case strings.HasPrefix(path, "/api/v1/system/parameters/") && method == http.MethodDelete:
		return "parameter:delete"
	case strings.HasPrefix(path, "/api/v1/system/parameters/"):
		return "parameter:update"
	case path == "/api/v1/system/dictionaries" && method == http.MethodGet:
		return "dictionary:read"
	case path == "/api/v1/system/dictionaries" && method == http.MethodPost:
		return "dictionary:create"
	case strings.HasPrefix(path, "/api/v1/system/dictionaries/") && method == http.MethodDelete:
		return "dictionary:delete"
	case strings.HasPrefix(path, "/api/v1/system/dictionaries/"):
		return "dictionary:update"
	case strings.HasPrefix(path, "/api/v1/system/dictionary-items/") && method == http.MethodDelete:
		return "dictionary:delete"
	case strings.HasPrefix(path, "/api/v1/system/dictionary-items/"):
		return "dictionary:update"
	case strings.Contains(path, "/users/invitations") || strings.Contains(path, "/invitations"):
		return "user:invite"
	case strings.Contains(path, "/users/"):
		return "user:update"
	case strings.HasSuffix(path, "/users"):
		return "user:read"
	case strings.Contains(path, "/roles/"):
		return "role:update"
	case strings.HasSuffix(path, "/roles") && method == http.MethodPost:
		return "role:create"
	case strings.HasSuffix(path, "/roles"):
		return "role:read"
	case strings.HasSuffix(path, "/permissions"):
		return "permission:read"
	case strings.Contains(path, "/sessions/"):
		return "session:revoke"
	case strings.HasSuffix(path, "/sessions"):
		return "session:read"
	case strings.HasSuffix(path, "/audit-logs"):
		return "audit:read"
	case strings.HasPrefix(path, "/api/v1/orgs/") && method == http.MethodPatch:
		return "org:update"
	case path == "/api/v1/orgs" && method == http.MethodPost:
		return "org:create"
	case path == "/api/v1/orgs":
		return "org:read"
	default:
		return ""
	}
}

func apiRouteMethodOrder(method string) int {
	switch method {
	case http.MethodGet:
		return 10
	case http.MethodPost:
		return 20
	case http.MethodPatch, http.MethodPut:
		return 30
	case http.MethodDelete:
		return 40
	default:
		return 50
	}
}

// health 返回轻量存活探针响应，只证明进程与路由栈仍可处理请求。
func health(c web.Context) {
	c.JSON(http.StatusOK, result.Success(map[string]any{"status": "ok"}))
}

// ready 执行数据库就绪检查，并把失败原因转化为 503 响应。
func ready(db database.Database) web.HandlerFunc {
	return func(c web.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, &result.Result[map[string]any]{
				Code:    apperrors.ErrDatabaseError,
				Message: "not ready",
				Data: map[string]any{
					"status": "not_ready",
					"checks": map[string]any{"database": "missing"},
				},
				ServerTime: time.Now().Unix(),
			})
			return
		}
		if err := db.Ping(c.RequestContext()); err != nil {
			c.JSON(http.StatusServiceUnavailable, &result.Result[map[string]any]{
				Code:    apperrors.ErrDatabaseError,
				Message: "not ready",
				Data: map[string]any{
					"status": "not_ready",
					"checks": map[string]any{"database": err.Error()},
				},
				ServerTime: time.Now().Unix(),
			})
			return
		}
		c.JSON(http.StatusOK, result.Success(map[string]any{
			"status": "ready",
			"checks": map[string]any{"database": "ok"},
		}))
	}
}

// ReadyCheck 构造就绪探针回调，通过数据库健康检查表达服务是否可以承接流量。
func ReadyCheck(db database.Database) func(context.Context) error {
	return func(ctx context.Context) error {
		if db == nil {
			return http.ErrServerClosed
		}
		return db.Ping(ctx)
	}
}
