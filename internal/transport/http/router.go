package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/internal/middleware"
	demohandler "github.com/rei0721/go-scaffold/internal/modules/demo/handler"
	iamhandler "github.com/rei0721/go-scaffold/internal/modules/iam/handler"
	pluginhandler "github.com/rei0721/go-scaffold/internal/modules/plugins/handler"
	systemhandler "github.com/rei0721/go-scaffold/internal/modules/system/handler"
	systemmodel "github.com/rei0721/go-scaffold/internal/modules/system/model"
	systemservice "github.com/rei0721/go-scaffold/internal/modules/system/service"
	"github.com/rei0721/go-scaffold/internal/ports"
	appconstants "github.com/rei0721/go-scaffold/types/constants"
	apperrors "github.com/rei0721/go-scaffold/types/errors"
	"github.com/rei0721/go-scaffold/types/result"
)

// RouterDeps 聚合 HTTP 路由装配所需依赖。
type RouterDeps struct {
	Router           ports.HTTPRouter
	RouteLister      ports.RouteLister
	StaticSPA        ports.StaticSPAMounter
	Logger           ports.Logger
	I18n             ports.I18n
	Database         ports.Database
	TraceIDGenerator ports.IDGenerator
	Middleware       middleware.MiddlewareConfig
	TodoHandler      *demohandler.TodoHandler
	CustomerHandler  *demohandler.CustomerHandler
	IAMHandler       *iamhandler.Handler
	PluginHandler    *pluginhandler.Handler
	SystemHandler    *systemhandler.Handler
	IAMAuth          middleware.Authenticator
	IAMAuthz         middleware.Authorizer
	WebUI            WebUIDeps
}

// WebUIDeps 描述管理台静态产物挂载所需配置。
type WebUIDeps struct {
	Enabled   bool
	MountPath string
	DistDir   string
}

// NewRouter 把中间件和业务路由注册到传入的 router。
func NewRouter(deps RouterDeps) ports.HTTPRouter {
	r := deps.Router
	if r == nil {
		return nil
	}

	if deps.I18n != nil {
		r.Use(middleware.I18n(deps.I18n))
	}
	r.Use(middleware.TraceID(deps.Middleware.TraceID, deps.TraceIDGenerator))
	r.Use(middleware.CORSMiddleware(deps.Middleware.CORS))
	if deps.Logger != nil {
		r.Use(middleware.Logger(deps.Middleware.Logger, deps.Logger))
		r.Use(middleware.Recovery(deps.Middleware.Recovery, deps.Logger))
	} else {
		r.Use(middleware.Recovery(deps.Middleware.Recovery, nil))
	}

	r.GET(appconstants.HTTPHealthPath, health)
	r.GET(appconstants.HTTPReadyPath, ready(deps.Database))

	v1 := r.Group(appconstants.APIBasePath)
	demo := v1.Group("/demo")
	if deps.TodoHandler != nil {
		todos := demo.Group("/todos")
		todos.POST("", deps.TodoHandler.Create)
		todos.GET("", deps.TodoHandler.List)
		todos.GET("/:id", deps.TodoHandler.Get)
		todos.PUT("/:id", deps.TodoHandler.Update)
		todos.DELETE("/:id", deps.TodoHandler.Delete)
	}
	if deps.CustomerHandler != nil {
		customers := demo.Group("/customers")
		customers.Use(middleware.Auth(deps.IAMAuth))
		customers.Use(OperationRecorder(deps.SystemHandler))
		customers.POST("", middleware.RequirePermission(deps.IAMAuthz, "customer", "create", deps.CustomerHandler.Create))
		customers.GET("", middleware.RequirePermission(deps.IAMAuthz, "customer", "read", deps.CustomerHandler.List))
		customers.GET("/:id", middleware.RequirePermission(deps.IAMAuthz, "customer", "read", deps.CustomerHandler.Get))
		customers.PATCH("/:id", middleware.RequirePermission(deps.IAMAuthz, "customer", "update", deps.CustomerHandler.Update))
		customers.DELETE("/:id", middleware.RequirePermission(deps.IAMAuthz, "customer", "delete", deps.CustomerHandler.Delete))
	}
	if deps.IAMHandler != nil {
		registerIAMRoutes(v1, deps)
	}
	if deps.PluginHandler != nil {
		registerPluginRoutes(v1, deps)
	}
	if deps.SystemHandler != nil {
		registerSystemRoutes(v1, deps)
		if deps.RouteLister != nil {
			deps.SystemHandler.RegisterAPIs(catalogAPIRoutes(deps.RouteLister.Routes()))
		}
	}
	registerWebUI(deps)

	return r
}

func registerWebUI(deps RouterDeps) {
	if !deps.WebUI.Enabled {
		return
	}
	mounter := deps.StaticSPA
	if mounter == nil {
		if candidate, ok := deps.Router.(ports.StaticSPAMounter); ok {
			mounter = candidate
		}
	}
	if mounter == nil {
		if deps.Logger != nil {
			deps.Logger.Warn("admin webui mount skipped", "mount_path", deps.WebUI.MountPath, "dist_dir", deps.WebUI.DistDir, "error", "static spa mounter missing")
		}
		return
	}
	err := mounter.MountStaticSPA(ports.StaticSPAConfig{
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
	if errors.Is(err, ports.ErrStaticSPAIndexMissing) {
		deps.Logger.Warn("admin webui static files missing", "mount_path", deps.WebUI.MountPath, "dist_dir", deps.WebUI.DistDir)
		return
	}
	deps.Logger.Warn("admin webui mount skipped", "mount_path", deps.WebUI.MountPath, "dist_dir", deps.WebUI.DistDir, "error", err)
}

func registerIAMRoutes(v1 ports.HTTPRouter, deps RouterDeps) {
	auth := v1.Group("/auth")
	auth.Use(middleware.RateLimit(middleware.RateLimitConfig{Enabled: true, Limit: 20, Window: time.Minute}))
	auth.GET("/setup/status", deps.IAMHandler.SetupStatus)
	auth.POST("/setup/initial-admin", deps.IAMHandler.InitialAdminSetup)
	auth.POST("/signup", deps.IAMHandler.Signup)
	auth.GET("/captcha", deps.IAMHandler.Captcha)
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
	orgScoped := func(obj, act string, next ports.HTTPHandlerFunc) ports.HTTPHandlerFunc {
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
	orgs.GET("/:orgId/api-tokens", orgScoped("api_token", "read", deps.IAMHandler.ListAPITokens))
	orgs.POST("/:orgId/api-tokens", orgScoped("api_token", "create", deps.IAMHandler.CreateAPIToken))
	orgs.DELETE("/:orgId/api-tokens/:tokenId", orgScoped("api_token", "revoke", deps.IAMHandler.RevokeAPIToken))
	orgs.GET("/:orgId/roles", orgScoped("role", "read", deps.IAMHandler.ListRoles))
	orgs.POST("/:orgId/roles", orgScoped("role", "create", deps.IAMHandler.CreateRole))
	orgs.PATCH("/:orgId/roles/:roleId", orgScoped("role", "update", deps.IAMHandler.UpdateRole))
	orgs.GET("/:orgId/permissions", orgScoped("permission", "read", deps.IAMHandler.ListPermissions))
	orgs.GET("/:orgId/sessions", orgScoped("session", "read", deps.IAMHandler.ListSessions))
	orgs.DELETE("/:orgId/sessions/:sessionId", orgScoped("session", "revoke", deps.IAMHandler.RevokeSession))
	orgs.GET("/:orgId/audit-logs", orgScoped("audit", "read", deps.IAMHandler.ListAuditLogs))
}

func registerPluginRoutes(v1 ports.HTTPRouter, deps RouterDeps) {
	plugins := v1.Group("/plugins")
	plugins.Use(middleware.Auth(deps.IAMAuth))
	plugins.Use(OperationRecorder(deps.SystemHandler))
	plugins.GET("", middleware.RequirePermission(deps.IAMAuthz, "plugin", "read", deps.PluginHandler.List))
	plugins.GET("/:pluginId", middleware.RequirePermission(deps.IAMAuthz, "plugin", "read", deps.PluginHandler.Get))
	plugins.GET("/:pluginId/health", middleware.RequirePermission(deps.IAMAuthz, "plugin", "read", deps.PluginHandler.Health))
	plugins.ANY("/:pluginId/proxy/*path", middleware.RequirePermission(deps.IAMAuthz, "plugin", "proxy", deps.PluginHandler.Proxy))
}

func registerSystemRoutes(v1 ports.HTTPRouter, deps RouterDeps) {
	system := v1.Group("/system")
	system.Use(middleware.Auth(deps.IAMAuth))
	system.Use(OperationRecorder(deps.SystemHandler))
	system.GET("/menus", deps.SystemHandler.ListMenus)
	system.GET("/config", middleware.RequirePermission(deps.IAMAuthz, "config", "read", deps.SystemHandler.ListConfig))
	system.PATCH("/config", middleware.RequirePermission(deps.IAMAuthz, "config", "update", deps.SystemHandler.UpdateConfig))
	system.GET("/server-info", middleware.RequirePermission(deps.IAMAuthz, "server", "read", deps.SystemHandler.GetServerInfo))
	system.GET("/apis", middleware.RequirePermission(deps.IAMAuthz, "permission", "read", deps.SystemHandler.ListAPIs))
	system.POST("/apis/sync", middleware.RequirePermission(deps.IAMAuthz, "permission", "read", deps.SystemHandler.SyncAPIs))
	system.POST("/apis/permissions/sync", middleware.RequirePermission(deps.IAMAuthz, "permission", "sync", deps.SystemHandler.SyncPermissions))
	system.GET("/operation-records", middleware.RequirePermission(deps.IAMAuthz, "operation", "read", deps.SystemHandler.ListOperationRecords))
	system.DELETE("/operation-records", middleware.RequirePermission(deps.IAMAuthz, "operation", "delete", deps.SystemHandler.DeleteOperationRecords))
	system.GET("/versions", middleware.RequirePermission(deps.IAMAuthz, "version", "read", deps.SystemHandler.ListVersions))
	system.POST("/versions/export", middleware.RequirePermission(deps.IAMAuthz, "version", "create", deps.SystemHandler.ExportVersion))
	system.POST("/versions/import", middleware.RequirePermission(deps.IAMAuthz, "version", "import", deps.SystemHandler.ImportVersion))
	system.DELETE("/versions", middleware.RequirePermission(deps.IAMAuthz, "version", "delete", deps.SystemHandler.DeleteVersions))
	system.GET("/versions/sources", middleware.RequirePermission(deps.IAMAuthz, "version", "read", deps.SystemHandler.ListVersionSources))
	system.GET("/versions/:versionId", middleware.RequirePermission(deps.IAMAuthz, "version", "read", deps.SystemHandler.GetVersion))
	system.GET("/versions/:versionId/download", middleware.RequirePermission(deps.IAMAuthz, "version", "download", deps.SystemHandler.DownloadVersion))
	system.DELETE("/versions/:versionId", middleware.RequirePermission(deps.IAMAuthz, "version", "delete", deps.SystemHandler.DeleteVersion))
	system.GET("/media/categories", middleware.RequirePermission(deps.IAMAuthz, "media", "read", deps.SystemHandler.ListMediaCategories))
	system.POST("/media/categories", middleware.RequirePermission(deps.IAMAuthz, "media", "update", deps.SystemHandler.UpsertMediaCategory))
	system.DELETE("/media/categories/:categoryId", middleware.RequirePermission(deps.IAMAuthz, "media", "update", deps.SystemHandler.DeleteMediaCategory))
	system.GET("/media/assets", middleware.RequirePermission(deps.IAMAuthz, "media", "read", deps.SystemHandler.ListMediaAssets))
	system.POST("/media/assets/upload", middleware.RequirePermission(deps.IAMAuthz, "media", "upload", deps.SystemHandler.UploadMediaAsset))
	system.POST("/media/assets/resumable/check", middleware.RequirePermission(deps.IAMAuthz, "media", "upload", deps.SystemHandler.CheckMediaResumableUpload))
	system.POST("/media/assets/resumable/chunks", middleware.RequirePermission(deps.IAMAuthz, "media", "upload", deps.SystemHandler.UploadMediaChunk))
	system.POST("/media/assets/resumable/complete", middleware.RequirePermission(deps.IAMAuthz, "media", "upload", deps.SystemHandler.CompleteMediaResumableUpload))
	system.POST("/media/assets/resumable/abort", middleware.RequirePermission(deps.IAMAuthz, "media", "upload", deps.SystemHandler.AbortMediaResumableUpload))
	system.POST("/media/assets/import-url", middleware.RequirePermission(deps.IAMAuthz, "media", "import", deps.SystemHandler.ImportMediaURLs))
	system.PATCH("/media/assets/:assetId", middleware.RequirePermission(deps.IAMAuthz, "media", "update", deps.SystemHandler.UpdateMediaAsset))
	system.GET("/media/assets/:assetId/download", middleware.RequirePermission(deps.IAMAuthz, "media", "download", deps.SystemHandler.DownloadMediaAsset))
	system.DELETE("/media/assets/:assetId", middleware.RequirePermission(deps.IAMAuthz, "media", "delete", deps.SystemHandler.DeleteMediaAsset))
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

func OperationRecorder(recorder operationRecorder) ports.HTTPHandlerFunc {
	return func(c ports.HTTPContext) {
		if isNilOperationRecorder(recorder) || !appconstants.IsAPIPath(c.Path()) {
			c.Next()
			return
		}
		body := sanitizeOperationRequestBody(c.Method(), c.Path(), readRequestBody(c.Request()))
		start := time.Now()
		c.Next()

		status := c.Status()
		if status == 0 {
			status = http.StatusOK
		}
		principal, _ := middleware.GetPrincipal(c)
		input := systemservice.OperationRecordInput{
			Body:      body,
			IPAddress: middleware.ClientIPRealIP(c),
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

func isNilOperationRecorder(recorder operationRecorder) bool {
	if recorder == nil {
		return true
	}
	value := reflect.ValueOf(recorder)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
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

func sanitizeOperationRequestBody(method string, path string, body string) string {
	if method != http.MethodPatch || path != appconstants.APIPath("system", "config") {
		return body
	}
	var payload struct {
		Items []struct {
			Key string `json:"key"`
		} `json:"items"`
		Persist bool `json:"persist"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return `{"items":"[redacted]"}`
	}
	out := struct {
		Items   []map[string]string `json:"items"`
		Persist bool                `json:"persist"`
	}{
		Items:   make([]map[string]string, 0, len(payload.Items)),
		Persist: payload.Persist,
	}
	for _, item := range payload.Items {
		out.Items = append(out.Items, map[string]string{
			"key":   item.Key,
			"value": "[redacted]",
		})
	}
	raw, err := json.Marshal(out)
	if err != nil {
		return `{"items":"[redacted]"}`
	}
	return string(raw)
}

func catalogAPIRoutes(routes []ports.RouteInfo) []systemmodel.APIEntry {
	entries := make([]systemmodel.APIEntry, 0, len(routes))
	for _, route := range routes {
		if !appconstants.IsAPIPath(route.Path) {
			continue
		}
		permission := apiRoutePermission(route.Method, route.Path)
		entries = append(entries, systemmodel.APIEntry{
			Access:      apiRouteAccess(route.Path, permission),
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

func apiRouteAccess(path string, permission string) string {
	if permission != "" {
		return systemmodel.APIAccessPermission
	}
	if publicAPIRoute(path) {
		return systemmodel.APIAccessPublic
	}
	return systemmodel.APIAccessAuthenticated
}

func publicAPIRoute(path string) bool {
	switch path {
	case appconstants.APIPath("auth", "setup", "status"),
		appconstants.APIPath("auth", "setup", "initial-admin"),
		appconstants.APIPath("auth", "signup"),
		appconstants.APIPath("auth", "captcha"),
		appconstants.APIPath("auth", "login"),
		appconstants.APIPath("auth", "refresh"),
		appconstants.APIPath("auth", "password", "forgot"),
		appconstants.APIPath("auth", "password", "reset"):
		return true
	}
	return strings.HasPrefix(path, appconstants.APIPath("invitations")+"/")
}

func apiRouteGroup(path string) string {
	path = appconstants.TrimAPIPathPrefix(path)
	segment, _, _ := strings.Cut(path, "/")
	segment = strings.TrimSpace(segment)
	if segment == "" {
		return "other"
	}
	return segment
}

func apiRouteDescription(method string, path string, permission string) string {
	if strings.HasPrefix(path, appconstants.APIPath("auth")+"/") || strings.HasPrefix(path, appconstants.APIPath("invitations")+"/") {
		return "认证流程接口"
	}
	if permission != "" {
		return "权限保护接口：" + permission
	}
	if strings.HasPrefix(path, appconstants.APIPath("system", "menus")) {
		return "当前用户可见菜单"
	}
	return method + " " + path
}

func apiRoutePermission(method string, path string) string {
	pluginsPath := appconstants.APIPath("plugins")
	demoCustomersPath := appconstants.APIPath("demo", "customers")
	orgsPath := appconstants.APIPath("orgs")
	systemConfigPath := appconstants.APIPath("system", "config")
	systemServerInfoPath := appconstants.APIPath("system", "server-info")
	systemAPIsPath := appconstants.APIPath("system", "apis")
	systemAPIsSyncPath := appconstants.APIPath("system", "apis", "sync")
	systemAPIPermissionsSyncPath := appconstants.APIPath("system", "apis", "permissions", "sync")
	systemOperationRecordsPath := appconstants.APIPath("system", "operation-records")
	systemVersionsPath := appconstants.APIPath("system", "versions")
	systemMediaCategoriesPath := appconstants.APIPath("system", "media", "categories")
	systemMediaAssetsPath := appconstants.SystemMediaAssetsAPIPath
	systemParametersPath := appconstants.APIPath("system", "parameters")
	systemDictionariesPath := appconstants.APIPath("system", "dictionaries")
	systemDictionaryItemsPath := appconstants.APIPath("system", "dictionary-items")

	switch {
	case strings.HasPrefix(path, pluginsPath+"/") && strings.Contains(path, "/proxy/"):
		return "plugin:proxy"
	case path == pluginsPath || strings.HasPrefix(path, pluginsPath+"/"):
		return "plugin:read"
	case path == demoCustomersPath && method == http.MethodGet:
		return "customer:read"
	case path == demoCustomersPath && method == http.MethodPost:
		return "customer:create"
	case strings.HasPrefix(path, demoCustomersPath+"/") && method == http.MethodDelete:
		return "customer:delete"
	case strings.HasPrefix(path, demoCustomersPath+"/") && method == http.MethodGet:
		return "customer:read"
	case strings.HasPrefix(path, demoCustomersPath+"/"):
		return "customer:update"
	case path == systemConfigPath && method == http.MethodPatch:
		return "config:update"
	case path == systemConfigPath:
		return "config:read"
	case path == systemServerInfoPath:
		return "server:read"
	case path == systemAPIsPath:
		return "permission:read"
	case path == systemAPIsSyncPath:
		return "permission:read"
	case path == systemAPIPermissionsSyncPath:
		return "permission:sync"
	case path == systemOperationRecordsPath && method == http.MethodDelete:
		return "operation:delete"
	case path == systemOperationRecordsPath:
		return "operation:read"
	case path == systemVersionsPath && method == http.MethodGet:
		return "version:read"
	case path == systemVersionsPath && method == http.MethodDelete:
		return "version:delete"
	case path == appconstants.APIPath("system", "versions", "export"):
		return "version:create"
	case path == appconstants.APIPath("system", "versions", "import"):
		return "version:import"
	case path == appconstants.APIPath("system", "versions", "sources"):
		return "version:read"
	case strings.HasPrefix(path, systemVersionsPath+"/") && strings.HasSuffix(path, "/download"):
		return "version:download"
	case strings.HasPrefix(path, systemVersionsPath+"/") && method == http.MethodDelete:
		return "version:delete"
	case strings.HasPrefix(path, systemVersionsPath+"/"):
		return "version:read"
	case path == systemMediaCategoriesPath && method == http.MethodGet:
		return "media:read"
	case path == systemMediaCategoriesPath && method == http.MethodPost:
		return "media:update"
	case strings.HasPrefix(path, systemMediaCategoriesPath+"/"):
		return "media:update"
	case path == systemMediaAssetsPath && method == http.MethodGet:
		return "media:read"
	case path == appconstants.APIPath("system", "media", "assets", "upload"):
		return "media:upload"
	case strings.HasPrefix(path, appconstants.APIPath("system", "media", "assets", "resumable")+"/"):
		return "media:upload"
	case path == appconstants.APIPath("system", "media", "assets", "import-url"):
		return "media:import"
	case strings.HasPrefix(path, systemMediaAssetsPath+"/") && strings.HasSuffix(path, "/download"):
		return "media:download"
	case strings.HasPrefix(path, systemMediaAssetsPath+"/") && method == http.MethodDelete:
		return "media:delete"
	case strings.HasPrefix(path, systemMediaAssetsPath+"/"):
		return "media:update"
	case path == systemParametersPath && method == http.MethodGet:
		return "parameter:read"
	case path == systemParametersPath && method == http.MethodPost:
		return "parameter:create"
	case path == systemParametersPath && method == http.MethodDelete:
		return "parameter:delete"
	case path == appconstants.APIPath("system", "parameters", "value"):
		return "parameter:read"
	case strings.HasPrefix(path, systemParametersPath+"/") && method == http.MethodGet:
		return "parameter:read"
	case strings.HasPrefix(path, systemParametersPath+"/") && method == http.MethodDelete:
		return "parameter:delete"
	case strings.HasPrefix(path, systemParametersPath+"/"):
		return "parameter:update"
	case path == systemDictionariesPath && method == http.MethodGet:
		return "dictionary:read"
	case path == systemDictionariesPath && method == http.MethodPost:
		return "dictionary:create"
	case strings.HasPrefix(path, systemDictionariesPath+"/") && method == http.MethodDelete:
		return "dictionary:delete"
	case strings.HasPrefix(path, systemDictionariesPath+"/"):
		return "dictionary:update"
	case strings.HasPrefix(path, systemDictionaryItemsPath+"/") && method == http.MethodDelete:
		return "dictionary:delete"
	case strings.HasPrefix(path, systemDictionaryItemsPath+"/"):
		return "dictionary:update"
	case strings.Contains(path, "/api-tokens/"):
		return "api_token:revoke"
	case strings.HasSuffix(path, "/api-tokens") && method == http.MethodPost:
		return "api_token:create"
	case strings.HasSuffix(path, "/api-tokens"):
		return "api_token:read"
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
	case strings.HasPrefix(path, orgsPath+"/") && method == http.MethodPatch:
		return "org:update"
	case path == orgsPath && method == http.MethodPost:
		return "org:create"
	case path == orgsPath:
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

// health 返回轻量存活探针响应。
func health(c ports.HTTPContext) {
	c.JSON(http.StatusOK, result.Success(map[string]any{"status": "ok"}))
}

// ready 执行数据库就绪检查。
func ready(db ports.Database) ports.HTTPHandlerFunc {
	return func(c ports.HTTPContext) {
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

// ReadyCheck 构造就绪探针回调。
func ReadyCheck(db ports.Database) func(context.Context) error {
	return func(ctx context.Context) error {
		if db == nil {
			return http.ErrServerClosed
		}
		return db.Ping(ctx)
	}
}
