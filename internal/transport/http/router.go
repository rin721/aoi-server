package httptransport

// 本文件定义 HTTP 传输层装配，把中间件顺序、健康检查和业务路由注册为 Gin Engine。

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/rei0721/go-scaffold/internal/middleware"
	demohandler "github.com/rei0721/go-scaffold/internal/modules/demo/handler"
	iamhandler "github.com/rei0721/go-scaffold/internal/modules/iam/handler"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/i18n"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/web"
	apperrors "github.com/rei0721/go-scaffold/types/errors"
	"github.com/rei0721/go-scaffold/types/result"
)

// RouterDeps 聚合 HTTP 路由装配所需依赖，允许测试或可选模块传入 nil 以裁剪路由。
type RouterDeps struct {
	Mode        string
	Logger      logger.Logger
	I18n        i18n.I18n
	Database    database.Database
	Middleware  middleware.MiddlewareConfig
	TodoHandler *demohandler.TodoHandler
	IAMHandler  *iamhandler.Handler
	IAMAuth     middleware.Authenticator
	IAMAuthz    middleware.Authorizer
	WebUI       WebUIDeps
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
	auth.POST("/login", deps.IAMHandler.Login)
	auth.POST("/refresh", deps.IAMHandler.Refresh)
	auth.POST("/password/forgot", deps.IAMHandler.ForgotPassword)
	auth.POST("/password/reset", deps.IAMHandler.ResetPassword)

	invitations := v1.Group("/invitations")
	invitations.POST("/:token/accept", deps.IAMHandler.AcceptInvitation)

	protected := v1.Group("")
	protected.Use(middleware.Auth(deps.IAMAuth))
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
	orgs.GET("/:orgId/users", orgScoped("user", "read", deps.IAMHandler.ListUsers))
	orgs.POST("/:orgId/users/invitations", orgScoped("user", "invite", deps.IAMHandler.InviteUser))
	orgs.GET("/:orgId/roles", orgScoped("role", "read", deps.IAMHandler.ListRoles))
	orgs.POST("/:orgId/roles", orgScoped("role", "create", deps.IAMHandler.CreateRole))
	orgs.GET("/:orgId/permissions", orgScoped("permission", "read", deps.IAMHandler.ListPermissions))
	orgs.GET("/:orgId/sessions", orgScoped("session", "read", deps.IAMHandler.ListSessions))
	orgs.DELETE("/:orgId/sessions/:sessionId", orgScoped("session", "revoke", deps.IAMHandler.RevokeSession))
	orgs.GET("/:orgId/audit-logs", orgScoped("audit", "read", deps.IAMHandler.ListAuditLogs))
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
