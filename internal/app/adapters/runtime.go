package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/rei0721/go-scaffold/internal/ports"
	"github.com/rei0721/go-scaffold/pkg/authorization"
	"github.com/rei0721/go-scaffold/pkg/hostmetrics"
	"github.com/rei0721/go-scaffold/pkg/mfa"
	"github.com/rei0721/go-scaffold/pkg/rpcserver"
	"github.com/rei0721/go-scaffold/pkg/token"
	"github.com/rei0721/go-scaffold/pkg/web"
)

type HTTPRouter struct {
	inner web.Router
}

func NewHTTPRouter(router web.Router) ports.HTTPRouter {
	if router == nil {
		return nil
	}
	return HTTPRouter{inner: router}
}

func NewHTTPEngine(engine *web.Engine) *HTTPEngine {
	if engine == nil {
		return nil
	}
	return &HTTPEngine{Engine: engine}
}

type HTTPEngine struct {
	Engine *web.Engine
}

func (e *HTTPEngine) Use(handlers ...ports.HTTPHandlerFunc) {
	e.Engine.Use(wrapHTTPHandlers(handlers)...)
}

func (e *HTTPEngine) GET(path string, handler ports.HTTPHandlerFunc) {
	e.Engine.GET(path, wrapHTTPHandler(handler))
}

func (e *HTTPEngine) POST(path string, handler ports.HTTPHandlerFunc) {
	e.Engine.POST(path, wrapHTTPHandler(handler))
}

func (e *HTTPEngine) PATCH(path string, handler ports.HTTPHandlerFunc) {
	e.Engine.PATCH(path, wrapHTTPHandler(handler))
}

func (e *HTTPEngine) PUT(path string, handler ports.HTTPHandlerFunc) {
	e.Engine.PUT(path, wrapHTTPHandler(handler))
}

func (e *HTTPEngine) DELETE(path string, handler ports.HTTPHandlerFunc) {
	e.Engine.DELETE(path, wrapHTTPHandler(handler))
}

func (e *HTTPEngine) ANY(path string, handler ports.HTTPHandlerFunc) {
	e.Engine.ANY(path, wrapHTTPHandler(handler))
}

func (e *HTTPEngine) Group(path string) ports.HTTPRouter {
	return NewHTTPRouter(e.Engine.Group(path))
}

func (e *HTTPEngine) Routes() []ports.RouteInfo {
	return routeInfos(e.Engine.Routes())
}

func (e *HTTPEngine) MountStaticSPA(cfg ports.StaticSPAConfig) error {
	err := e.Engine.MountStaticSPA(web.StaticSPAConfig{
		MountPath: cfg.MountPath,
		DistDir:   cfg.DistDir,
	})
	if errors.Is(err, web.ErrStaticSPAIndexMissing) {
		return ports.ErrStaticSPAIndexMissing
	}
	return err
}

func (r HTTPRouter) Use(handlers ...ports.HTTPHandlerFunc) {
	r.inner.Use(wrapHTTPHandlers(handlers)...)
}

func (r HTTPRouter) GET(path string, handler ports.HTTPHandlerFunc) {
	r.inner.GET(path, wrapHTTPHandler(handler))
}

func (r HTTPRouter) POST(path string, handler ports.HTTPHandlerFunc) {
	r.inner.POST(path, wrapHTTPHandler(handler))
}

func (r HTTPRouter) PATCH(path string, handler ports.HTTPHandlerFunc) {
	r.inner.PATCH(path, wrapHTTPHandler(handler))
}

func (r HTTPRouter) PUT(path string, handler ports.HTTPHandlerFunc) {
	r.inner.PUT(path, wrapHTTPHandler(handler))
}

func (r HTTPRouter) DELETE(path string, handler ports.HTTPHandlerFunc) {
	r.inner.DELETE(path, wrapHTTPHandler(handler))
}

func (r HTTPRouter) ANY(path string, handler ports.HTTPHandlerFunc) {
	r.inner.ANY(path, wrapHTTPHandler(handler))
}

func (r HTTPRouter) Group(path string) ports.HTTPRouter {
	return NewHTTPRouter(r.inner.Group(path))
}

func CORS(cfg ports.CORSConfig) ports.HTTPHandlerFunc {
	return func(c ports.HTTPContext) {
		web.CORS(web.CORSConfig{
			Enabled:          cfg.Enabled,
			AllowOrigins:     cfg.AllowOrigins,
			AllowMethods:     cfg.AllowMethods,
			AllowHeaders:     cfg.AllowHeaders,
			ExposeHeaders:    cfg.ExposeHeaders,
			AllowCredentials: cfg.AllowCredentials,
			MaxAge:           cfg.MaxAge,
		})(unwrapHTTPContext(c))
	}
}

func Recovery() ports.HTTPHandlerFunc {
	return func(c ports.HTTPContext) {
		web.Recovery()(unwrapHTTPContext(c))
	}
}

func wrapHTTPHandlers(handlers []ports.HTTPHandlerFunc) []web.HandlerFunc {
	wrapped := make([]web.HandlerFunc, 0, len(handlers))
	for _, handler := range handlers {
		wrapped = append(wrapped, wrapHTTPHandler(handler))
	}
	return wrapped
}

func wrapHTTPHandler(handler ports.HTTPHandlerFunc) web.HandlerFunc {
	return func(c web.Context) {
		handler(c)
	}
}

func unwrapHTTPContext(c ports.HTTPContext) web.Context {
	ctx, ok := c.(web.Context)
	if !ok {
		return noopHTTPContext{HTTPContext: c}
	}
	return ctx
}

type noopHTTPContext struct {
	ports.HTTPContext
}

func routeInfos(routes []web.RouteInfo) []ports.RouteInfo {
	out := make([]ports.RouteInfo, 0, len(routes))
	for _, route := range routes {
		out = append(out, ports.RouteInfo{
			Method:  route.Method,
			Path:    route.Path,
			Handler: route.Handler,
		})
	}
	return out
}

type TokenManager struct {
	inner token.Manager
}

func NewTokenManager(manager token.Manager) ports.TokenManager {
	if manager == nil {
		return nil
	}
	return TokenManager{inner: manager}
}

func (m TokenManager) IssueAccess(ctx context.Context, subject ports.TokenSubject) (string, time.Time, error) {
	return m.inner.IssueAccess(ctx, tokenSubject(subject))
}

func (m TokenManager) IssueRefresh(ctx context.Context) (string, string, time.Time, error) {
	return m.inner.IssueRefresh(ctx)
}

func (m TokenManager) IssuePair(ctx context.Context, subject ports.TokenSubject) (ports.TokenPair, error) {
	pair, err := m.inner.IssuePair(ctx, tokenSubject(subject))
	return ports.TokenPair{
		AccessToken:      pair.AccessToken,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshToken:     pair.RefreshToken,
		RefreshTokenHash: pair.RefreshTokenHash,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}, err
}

func (m TokenManager) Parse(ctx context.Context, raw string, expectedType string) (*ports.TokenClaims, error) {
	claims, err := m.inner.Parse(ctx, raw, expectedType)
	if err != nil {
		return nil, err
	}
	return &ports.TokenClaims{
		UserID:    claims.UserID,
		OrgID:     claims.OrgID,
		SessionID: claims.SessionID,
		TokenType: claims.TokenType,
	}, nil
}

func (m TokenManager) HashRefreshToken(raw string) string {
	return m.inner.HashRefreshToken(raw)
}

func tokenSubject(subject ports.TokenSubject) token.Subject {
	return token.Subject{
		UserID:    subject.UserID,
		OrgID:     subject.OrgID,
		SessionID: subject.SessionID,
	}
}

type AuthorizerEnforcer struct {
	inner authorization.Enforcer
}

func NewAuthorizerEnforcer(enforcer authorization.Enforcer) ports.AuthorizerEnforcer {
	if enforcer == nil {
		return nil
	}
	return AuthorizerEnforcer{inner: enforcer}
}

func (e AuthorizerEnforcer) Enforce(ctx context.Context, sub, org, obj, act string) (bool, error) {
	return e.inner.Enforce(ctx, sub, org, obj, act)
}

func (e AuthorizerEnforcer) AddPolicy(ctx context.Context, role, org, obj, act string) (bool, error) {
	return e.inner.AddPolicy(ctx, role, org, obj, act)
}

func (e AuthorizerEnforcer) AddRoleForUser(ctx context.Context, user, role, org string) (bool, error) {
	return e.inner.AddRoleForUser(ctx, user, role, org)
}

func (e AuthorizerEnforcer) DeleteRoleForUser(ctx context.Context, user, role, org string) (bool, error) {
	return e.inner.DeleteRoleForUser(ctx, user, role, org)
}

func (e AuthorizerEnforcer) GetRolesForUser(ctx context.Context, user, org string) ([]string, error) {
	return e.inner.GetRolesForUser(ctx, user, org)
}

func (e AuthorizerEnforcer) LoadRules(ctx context.Context, rules []ports.AuthorizationRule) error {
	out := make([]authorization.Rule, 0, len(rules))
	for _, rule := range rules {
		out = append(out, authorization.Rule{PType: rule.PType, Values: rule.Values})
	}
	return e.inner.LoadRules(ctx, out)
}

type TOTPProvider struct{}

func (TOTPProvider) GenerateTOTP(issuer, accountName string) (ports.TOTPKey, error) {
	key, err := mfa.GenerateTOTP(issuer, accountName)
	return ports.TOTPKey{Secret: key.Secret, URL: key.URL}, err
}

func (TOTPProvider) ValidateTOTP(code, secret string) bool {
	return mfa.ValidateTOTP(code, secret)
}

type HostMetricsCollector struct{}

func (HostMetricsCollector) Collect(ctx context.Context) ports.HostMetrics {
	metrics := hostmetrics.Collect(ctx)
	disks := make([]ports.DiskInfo, 0, len(metrics.Disk))
	for _, disk := range metrics.Disk {
		disks = append(disks, ports.DiskInfo{
			FSType:      disk.FSType,
			MountPoint:  disk.MountPoint,
			TotalGB:     disk.TotalGB,
			TotalMB:     disk.TotalMB,
			UsedGB:      disk.UsedGB,
			UsedMB:      disk.UsedMB,
			UsedPercent: disk.UsedPercent,
		})
	}
	return ports.HostMetrics{
		CPU: ports.CPUInfo{
			Cores:   metrics.CPU.Cores,
			Percent: append([]float64(nil), metrics.CPU.Percent...),
		},
		RAM: ports.RAMInfo{
			TotalMB:     metrics.RAM.TotalMB,
			UsedMB:      metrics.RAM.UsedMB,
			UsedPercent: metrics.RAM.UsedPercent,
		},
		Disk: disks,
	}
}

type RPCRegistry struct {
	inner *rpcserver.Registry
}

func NewRPCRegistry(registry *rpcserver.Registry) ports.RPCRegistry {
	if registry == nil {
		return nil
	}
	return RPCRegistry{inner: registry}
}

func (r RPCRegistry) Register(method string, handler ports.RPCHandlerFunc) error {
	return r.inner.Register(method, func(ctx context.Context, params json.RawMessage) (any, error) {
		result, err := handler(ctx, params)
		if err == nil {
			return result, nil
		}
		var rpcErr *ports.RPCError
		if errors.As(err, &rpcErr) {
			return nil, &rpcserver.RPCError{Code: rpcErr.Code, Message: rpcErr.Message, Data: rpcErr.Data}
		}
		return nil, err
	})
}

func (r RPCRegistry) Methods() []string {
	return r.inner.Methods()
}

func UnwrapRPCRegistry(registry ports.RPCRegistry) *rpcserver.Registry {
	if wrapped, ok := registry.(RPCRegistry); ok {
		return wrapped.inner
	}
	if wrapped, ok := registry.(*RPCRegistry); ok && wrapped != nil {
		return wrapped.inner
	}
	return nil
}
