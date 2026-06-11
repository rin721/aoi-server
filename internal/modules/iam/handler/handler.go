package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/rei0721/go-scaffold/internal/middleware"
	"github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/web"
	"github.com/rei0721/go-scaffold/types/result"
)

type Handler struct {
	service service.Service
	logger  logger.Logger
}

func New(service service.Service, logger logger.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

type loginRequest struct {
	Identifier string `json:"identifier" binding:"required"`
	Password   string `json:"password" binding:"required"`
	OrgCode    string `json:"orgCode"`
	MFACode    string `json:"mfaCode"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type switchOrgRequest struct {
	OrgID int64 `json:"orgId" binding:"required"`
}

type createOrgRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type inviteUserRequest struct {
	Email    string `json:"email" binding:"required"`
	RoleCode string `json:"roleCode" binding:"required"`
}

type acceptInvitationRequest struct {
	Username    string `json:"username" binding:"required"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password" binding:"required"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

type verifyMFARequest struct {
	Code string `json:"code" binding:"required"`
}

type createRoleRequest struct {
	Code        string   `json:"code" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

func (h *Handler) Login(c web.Context) {
	var req loginRequest
	if !bind(c, &req) {
		return
	}
	pair, err := h.service.Login(c.RequestContext(), service.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
		OrgCode:    req.OrgCode,
		MFACode:    req.MFACode,
		UserAgent:  c.GetHeader("User-Agent"),
		IPAddress:  c.ClientIP(),
	})
	h.write(c, pair, err)
}

func (h *Handler) Refresh(c web.Context) {
	var req refreshRequest
	if !bind(c, &req) {
		return
	}
	pair, err := h.service.Refresh(c.RequestContext(), service.RefreshInput{
		RefreshToken: req.RefreshToken,
		UserAgent:    c.GetHeader("User-Agent"),
		IPAddress:    c.ClientIP(),
	})
	h.write(c, pair, err)
}

func (h *Handler) Logout(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	h.write(c, map[string]bool{"loggedOut": true}, h.service.Logout(c.RequestContext(), principal))
}

func (h *Handler) SwitchOrg(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req switchOrgRequest
	if !bind(c, &req) {
		return
	}
	pair, err := h.service.SwitchOrg(c.RequestContext(), principal, req.OrgID, c.GetHeader("User-Agent"), c.ClientIP())
	h.write(c, pair, err)
}

func (h *Handler) Me(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	user, err := h.service.Me(c.RequestContext(), principal)
	h.write(c, user, err)
}

func (h *Handler) MyOrganizations(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	orgs, err := h.service.ListMyOrganizations(c.RequestContext(), principal)
	h.write(c, orgs, err)
}

func (h *Handler) ListOrganizations(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	orgs, err := h.service.ListOrganizations(c.RequestContext(), principal)
	h.write(c, orgs, err)
}

func (h *Handler) CreateOrganization(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req createOrgRequest
	if !bind(c, &req) {
		return
	}
	org, err := h.service.CreateOrganization(c.RequestContext(), principal, req.Code, req.Name)
	h.writeCreated(c, org, err)
}

func (h *Handler) InviteUser(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req inviteUserRequest
	if !bind(c, &req) {
		return
	}
	token, err := h.service.InviteUser(c.RequestContext(), service.InviteUserInput{
		Principal: principal,
		Email:     req.Email,
		RoleCode:  req.RoleCode,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	})
	h.writeCreated(c, map[string]string{"token": token}, err)
}

func (h *Handler) AcceptInvitation(c web.Context) {
	var req acceptInvitationRequest
	if !bind(c, &req) {
		return
	}
	principal, err := h.service.AcceptInvitation(c.RequestContext(), service.AcceptInvitationInput{
		Token:       c.Param("token"),
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Password:    req.Password,
		UserAgent:   c.GetHeader("User-Agent"),
		IPAddress:   c.ClientIP(),
	})
	h.writeCreated(c, principal, err)
}

func (h *Handler) ForgotPassword(c web.Context) {
	var req forgotPasswordRequest
	if !bind(c, &req) {
		return
	}
	token, err := h.service.ForgotPassword(c.RequestContext(), service.ForgotPasswordInput{
		Email:     req.Email,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	})
	h.write(c, map[string]string{"token": token}, err)
}

func (h *Handler) ResetPassword(c web.Context) {
	var req resetPasswordRequest
	if !bind(c, &req) {
		return
	}
	err := h.service.ResetPassword(c.RequestContext(), service.ResetPasswordInput{
		Token:       req.Token,
		NewPassword: req.NewPassword,
		UserAgent:   c.GetHeader("User-Agent"),
		IPAddress:   c.ClientIP(),
	})
	h.write(c, map[string]bool{"reset": true}, err)
}

func (h *Handler) SetupMFA(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	secret, url, err := h.service.SetupMFA(c.RequestContext(), principal)
	h.write(c, map[string]string{"secret": secret, "otpauthUrl": url}, err)
}

func (h *Handler) VerifyMFA(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req verifyMFARequest
	if !bind(c, &req) {
		return
	}
	h.write(c, map[string]bool{"verified": true}, h.service.VerifyMFA(c.RequestContext(), principal, req.Code))
}

func (h *Handler) ListUsers(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	users, err := h.service.ListUsers(c.RequestContext(), principal)
	h.write(c, users, err)
}

func (h *Handler) ListRoles(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	roles, err := h.service.ListRoles(c.RequestContext(), principal)
	h.write(c, roles, err)
}

func (h *Handler) CreateRole(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req createRoleRequest
	if !bind(c, &req) {
		return
	}
	role, err := h.service.CreateRole(c.RequestContext(), service.CreateRoleInput{
		Principal:   principal,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Permissions: req.Permissions,
	})
	h.writeCreated(c, role, err)
}

func (h *Handler) ListPermissions(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	permissions, err := h.service.ListPermissions(c.RequestContext(), principal)
	h.write(c, permissions, err)
}

func (h *Handler) ListSessions(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	userID := int64(0)
	if raw := c.Request().URL.Query().Get("userId"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			result.BadRequest(c, "invalid userId")
			return
		}
		userID = parsed
	}
	sessions, err := h.service.ListSessions(c.RequestContext(), principal, userID)
	h.write(c, sessions, err)
}

func (h *Handler) RevokeSession(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	id, ok := parseInt64Param(c, "sessionId")
	if !ok {
		return
	}
	h.write(c, map[string]bool{"revoked": true}, h.service.RevokeSession(c.RequestContext(), principal, id))
}

func (h *Handler) ListAuditLogs(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	limit := 100
	if raw := c.Request().URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid limit")
			return
		}
		limit = parsed
	}
	logs, err := h.service.ListAuditLogs(c.RequestContext(), principal, limit)
	h.write(c, logs, err)
}

func bind(c web.Context, dest any) bool {
	if err := c.BindJSON(dest); err != nil {
		result.BadRequest(c, err.Error())
		return false
	}
	return true
}

func requirePrincipal(c web.Context) (service.Principal, bool) {
	principal, ok := middleware.GetPrincipal(c)
	if !ok {
		result.Unauthorized(c, "missing principal")
		return service.Principal{}, false
	}
	return principal, true
}

func parseInt64Param(c web.Context, name string) (int64, bool) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		result.BadRequest(c, "invalid "+name)
		return 0, false
	}
	return id, true
}

func (h *Handler) write(c web.Context, data any, err error) {
	if err != nil {
		h.writeError(c, err)
		return
	}
	result.OK(c, data)
}

func (h *Handler) writeCreated(c web.Context, data any, err error) {
	if err != nil {
		h.writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result.Success(data))
}

func (h *Handler) writeError(c web.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		result.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrUnauthorized), errors.Is(err, service.ErrMFARequired), errors.Is(err, service.ErrInvalidToken), errors.Is(err, service.ErrAccountLocked), errors.Is(err, service.ErrAccountDisabled), errors.Is(err, service.ErrSessionRevoked):
		result.Unauthorized(c, err.Error())
	case errors.Is(err, service.ErrForbidden):
		result.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrNotFound):
		result.NotFound(c, err.Error())
	case errors.Is(err, service.ErrDuplicate), errors.Is(err, service.ErrInvitationClosed):
		result.BadRequest(c, err.Error())
	default:
		if h.logger != nil {
			h.logger.Error("iam request failed", "error", err)
		}
		result.InternalError(c, "internal server error")
	}
}
