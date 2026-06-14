package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/rei0721/go-scaffold/internal/middleware"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/internal/modules/plugins/service"
	"github.com/rei0721/go-scaffold/internal/ports"
	"github.com/rei0721/go-scaffold/types/result"
)

type Auditor interface {
	RecordAudit(context.Context, iamservice.Principal, string, string, string, string, string, map[string]any) error
}

type dataWriter interface {
	Data(status int, contentType string, data []byte)
}

type Handler struct {
	service    service.Service
	authorizer middleware.Authorizer
	auditor    Auditor
	logger     ports.Logger
}

func New(service service.Service, authorizer middleware.Authorizer, auditor Auditor, logger ports.Logger) *Handler {
	return &Handler{service: service, authorizer: authorizer, auditor: auditor, logger: logger}
}

func (h *Handler) List(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	plugins, err := h.service.List(c.RequestContext())
	if err != nil {
		h.writeError(c, err)
		return
	}
	for i := range plugins {
		plugins[i] = h.filterManifest(c.RequestContext(), principal, plugins[i])
	}
	result.OK(c, plugins)
}

func (h *Handler) Get(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	plugin, err := h.service.Get(c.RequestContext(), c.Param("pluginId"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, h.filterManifest(c.RequestContext(), principal, plugin))
}

func (h *Handler) Health(c ports.HTTPContext) {
	status, err := h.service.Health(c.RequestContext(), c.Param("pluginId"))
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, status)
}

func (h *Handler) Proxy(c ports.HTTPContext) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	proxyPath := c.Param("path")
	response, err := h.service.Proxy(c.RequestContext(), service.ProxyRequest{
		PluginID: c.Param("pluginId"),
		Method:   c.Method(),
		Path:     proxyPath,
		RawQuery: c.Request().URL.RawQuery,
		Headers:  c.Request().Header.Clone(),
		Body:     c.Request().Body,
		Identity: service.ProxyIdentity{
			UserID:  strconv.FormatInt(principal.UserID, 10),
			OrgID:   strconv.FormatInt(principal.OrgID, 10),
			TraceID: middleware.GetTraceID(c),
		},
	})
	h.auditProxy(c, principal, proxyPath, response.StatusCode, err)
	if err != nil {
		h.writeError(c, err)
		return
	}
	for key, values := range response.Headers {
		for _, value := range values {
			c.Header(key, value)
		}
	}
	contentType := response.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	writer, ok := c.(dataWriter)
	if !ok {
		result.InternalError(c, "plugin response writer unavailable")
		return
	}
	writer.Data(response.StatusCode, contentType, response.Body)
}

func (h *Handler) filterManifest(ctx context.Context, principal iamservice.Principal, plugin service.Manifest) service.Manifest {
	if h.authorizer == nil {
		plugin.Menus = nil
		return plugin
	}
	menus := make([]service.Menu, 0, len(plugin.Menus))
	for _, menu := range plugin.Menus {
		if menu.Permission == "" || h.allowed(ctx, principal, menu.Permission) {
			menus = append(menus, menu)
		}
	}
	plugin.Menus = menus
	return plugin
}

func (h *Handler) allowed(ctx context.Context, principal iamservice.Principal, permission string) bool {
	obj, act := permissionObjectAction(permission)
	if obj == "" || act == "" {
		return false
	}
	allowed, err := h.authorizer.Authorize(ctx, principal, obj, act)
	return err == nil && allowed
}

func (h *Handler) auditProxy(c ports.HTTPContext, principal iamservice.Principal, path string, status int, err error) {
	if h.auditor == nil {
		return
	}
	metadata := map[string]any{
		"pluginId": c.Param("pluginId"),
		"method":   c.Method(),
		"path":     path,
		"status":   status,
	}
	if err != nil {
		metadata["error"] = err.Error()
	}
	_ = h.auditor.RecordAudit(c.RequestContext(), principal, "plugin.proxy", "plugin", c.Param("pluginId"), c.ClientIP(), c.GetHeader("User-Agent"), metadata)
}

func (h *Handler) writeError(c ports.HTTPContext, err error) {
	switch {
	case errors.Is(err, service.ErrDisabled):
		result.NotFound(c, "plugins disabled")
	case errors.Is(err, service.ErrPluginNotFound):
		result.NotFound(c, "plugin not found")
	case errors.Is(err, service.ErrProxyForbidden), errors.Is(err, service.ErrSecretMissing):
		result.Forbidden(c, err.Error())
	default:
		if h.logger != nil {
			h.logger.Error("plugin request failed", "error", err)
		}
		result.Fail(c, http.StatusBadGateway, "plugin request failed")
	}
}

func requirePrincipal(c ports.HTTPContext) (iamservice.Principal, bool) {
	principal, ok := middleware.GetPrincipal(c)
	if !ok {
		result.Unauthorized(c, "missing principal")
		return iamservice.Principal{}, false
	}
	return principal, true
}

func permissionObjectAction(code string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(code), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", ""
	}
	return parts[0], parts[1]
}
