package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/rei0721/go-scaffold/internal/middleware"
	"github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/pkg/logger"
	"github.com/rei0721/go-scaffold/pkg/web"
	"github.com/rei0721/go-scaffold/types/result"
)

type Handler struct {
	service      service.Service
	setupService setupService
	logger       logger.Logger
}

type setupService interface {
	SetupStatus(context.Context) (service.SetupStatus, error)
	InitialAdminSetup(context.Context, service.InitialAdminSetupInput) (service.TokenPair, error)
}

func New(service service.Service, logger logger.Logger) *Handler {
	return &Handler{service: service, setupService: iamSetupService{service: service}, logger: logger}
}

// UseSetupService 替换首次初始化专用后端，普通 IAM API 仍保持原服务实例。
func (h *Handler) UseSetupService(setup setupService) {
	if setup != nil {
		h.setupService = setup
	}
}

type iamSetupService struct {
	service service.Service
}

func (s iamSetupService) SetupStatus(ctx context.Context) (service.SetupStatus, error) {
	return s.service.SetupStatus(ctx)
}

func (s iamSetupService) InitialAdminSetup(ctx context.Context, input service.InitialAdminSetupInput) (service.TokenPair, error) {
	return s.service.InitialAdminSetup(ctx, input)
}

type loginRequest struct {
	CaptchaCode string `json:"captchaCode"`
	CaptchaID   string `json:"captchaId"`
	Identifier  string `json:"identifier" binding:"required"`
	Password    string `json:"password" binding:"required"`
	OrgCode     string `json:"orgCode"`
	MFACode     string `json:"mfaCode"`
}

type signupRequest struct {
	OrgCode     string `json:"orgCode" binding:"required"`
	OrgName     string `json:"orgName" binding:"required"`
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required"`
	DisplayName string `json:"displayName"`
	Password    string `json:"password" binding:"required"`
}

type initialAdminSetupRequest = signupRequest

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type switchOrgRequest struct {
	OrgID int64String `json:"orgId" binding:"required"`
}

type int64String int64

func (v *int64String) UnmarshalJSON(raw []byte) error {
	text := string(raw)
	if unquoted, err := strconv.Unquote(text); err == nil {
		text = unquoted
	}
	id, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return err
	}
	*v = int64String(id)
	return nil
}

type createOrgRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type updateOrgRequest struct {
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

type updateUserRequest struct {
	Status *string   `json:"status"`
	Roles  *[]string `json:"roles"`
}

type createAPITokenRequest struct {
	UserID   int64String `json:"userId" binding:"required"`
	RoleCode string      `json:"roleCode" binding:"required"`
	Days     int         `json:"days"`
	Remark   string      `json:"remark"`
}

type updateRoleRequest struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions *[]string `json:"permissions"`
}

func (h *Handler) Signup(c web.Context) {
	var req signupRequest
	if !bind(c, &req) {
		return
	}
	pair, err := h.service.Signup(c.RequestContext(), service.SignupInput{
		OrgCode:     req.OrgCode,
		OrgName:     req.OrgName,
		Username:    req.Username,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Password:    req.Password,
		UserAgent:   c.GetHeader("User-Agent"),
		IPAddress:   c.ClientIP(),
	})
	h.write(c, pair, err)
}

func (h *Handler) SetupStatus(c web.Context) {
	status, err := h.setupService.SetupStatus(c.RequestContext())
	h.write(c, status, err)
}

func (h *Handler) InitialAdminSetup(c web.Context) {
	var req initialAdminSetupRequest
	if !bind(c, &req) {
		return
	}
	pair, err := h.setupService.InitialAdminSetup(c.RequestContext(), service.InitialAdminSetupInput{
		OrgCode:     req.OrgCode,
		OrgName:     req.OrgName,
		Username:    req.Username,
		Email:       req.Email,
		DisplayName: req.DisplayName,
		Password:    req.Password,
		UserAgent:   c.GetHeader("User-Agent"),
		IPAddress:   c.ClientIP(),
	})
	if err != nil {
		h.writeSetupError(c, err)
		return
	}
	result.OK(c, pair)
}

func (h *Handler) Login(c web.Context) {
	var req loginRequest
	if !bind(c, &req) {
		return
	}
	pair, err := h.service.Login(c.RequestContext(), service.LoginInput{
		CaptchaCode: req.CaptchaCode,
		CaptchaID:   req.CaptchaID,
		Identifier:  req.Identifier,
		Password:    req.Password,
		OrgCode:     req.OrgCode,
		MFACode:     req.MFACode,
		UserAgent:   c.GetHeader("User-Agent"),
		IPAddress:   c.ClientIP(),
	})
	h.write(c, pair, err)
}

func (h *Handler) Captcha(c web.Context) {
	challenge, err := h.service.Captcha(c.RequestContext())
	h.write(c, challenge, err)
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
	pair, err := h.service.SwitchOrg(c.RequestContext(), principal, int64(req.OrgID), c.GetHeader("User-Agent"), c.ClientIP())
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
	filter, ok := parseOrganizationListFilter(c)
	if !ok {
		return
	}
	orgs, err := h.service.ListOrganizations(c.RequestContext(), principal, filter)
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

func (h *Handler) UpdateOrganization(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	orgID, ok := parseInt64Param(c, "orgId")
	if !ok {
		return
	}
	var req updateOrgRequest
	if !bind(c, &req) {
		return
	}
	org, err := h.service.UpdateOrganization(c.RequestContext(), service.UpdateOrganizationInput{
		Principal: principal,
		OrgID:     orgID,
		Name:      req.Name,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	})
	h.write(c, org, err)
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
	delivery, err := h.service.InviteUser(c.RequestContext(), service.InviteUserInput{
		Principal: principal,
		Email:     req.Email,
		RoleCode:  req.RoleCode,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	})
	h.writeCreated(c, delivery, err)
}

func (h *Handler) ListInvitations(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	invitations, err := h.service.ListInvitations(c.RequestContext(), principal)
	h.write(c, invitations, err)
}

func (h *Handler) RevokeInvitation(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	id, ok := parseInt64Param(c, "invitationId")
	if !ok {
		return
	}
	h.write(c, map[string]bool{"revoked": true}, h.service.RevokeInvitation(c.RequestContext(), principal, id, c.GetHeader("User-Agent"), c.ClientIP()))
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
	delivery, err := h.service.ForgotPassword(c.RequestContext(), service.ForgotPasswordInput{
		Email:     req.Email,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	})
	h.write(c, delivery, err)
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
	filter, ok := parseUserListFilter(c)
	if !ok {
		return
	}
	users, err := h.service.ListUsers(c.RequestContext(), principal, filter)
	h.write(c, users, err)
}

func (h *Handler) UpdateUser(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	userID, ok := parseInt64Param(c, "userId")
	if !ok {
		return
	}
	var req updateUserRequest
	if !bind(c, &req) {
		return
	}
	input := service.UpdateUserInput{
		Principal: principal,
		UserID:    userID,
		Status:    req.Status,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	}
	if req.Roles != nil {
		input.Roles = *req.Roles
		input.HasRoles = true
	}
	user, err := h.service.UpdateUser(c.RequestContext(), input)
	h.write(c, user, err)
}

func (h *Handler) ListAPITokens(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	filter, ok := parseAPITokenFilter(c)
	if !ok {
		return
	}
	page, err := h.service.ListAPITokens(c.RequestContext(), principal, filter)
	h.write(c, page, err)
}

func (h *Handler) CreateAPIToken(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	var req createAPITokenRequest
	if !bind(c, &req) {
		return
	}
	created, err := h.service.CreateAPIToken(c.RequestContext(), service.CreateAPITokenInput{
		Principal: principal,
		UserID:    int64(req.UserID),
		RoleCode:  req.RoleCode,
		Days:      req.Days,
		Remark:    req.Remark,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	})
	h.writeCreated(c, created, err)
}

func (h *Handler) RevokeAPIToken(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	id, ok := parseInt64Param(c, "tokenId")
	if !ok {
		return
	}
	h.write(c, map[string]bool{"revoked": true}, h.service.RevokeAPIToken(c.RequestContext(), service.RevokeAPITokenInput{
		Principal: principal,
		TokenID:   id,
		UserAgent: c.GetHeader("User-Agent"),
		IPAddress: c.ClientIP(),
	}))
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

func (h *Handler) UpdateRole(c web.Context) {
	principal, ok := requirePrincipal(c)
	if !ok {
		return
	}
	roleID, ok := parseInt64Param(c, "roleId")
	if !ok {
		return
	}
	var req updateRoleRequest
	if !bind(c, &req) {
		return
	}
	input := service.UpdateRoleInput{
		Principal:   principal,
		RoleID:      roleID,
		Name:        req.Name,
		Description: req.Description,
		UserAgent:   c.GetHeader("User-Agent"),
		IPAddress:   c.ClientIP(),
	}
	if req.Permissions != nil {
		input.Permissions = *req.Permissions
		input.HasPermissions = true
	}
	role, err := h.service.UpdateRole(c.RequestContext(), input)
	h.write(c, role, err)
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
	filter, ok := parseSessionListFilter(c)
	if !ok {
		return
	}
	sessions, err := h.service.ListSessions(c.RequestContext(), principal, filter)
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
	filter, ok := parseAuditLogFilter(c)
	if !ok {
		return
	}
	logs, err := h.service.ListAuditLogs(c.RequestContext(), principal, filter)
	h.write(c, logs, err)
}

func parseAuditLogFilter(c web.Context) (service.AuditLogFilter, bool) {
	query := c.Request().URL.Query()
	filter := service.AuditLogFilter{Action: query.Get("action")}
	if raw := query.Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid limit")
			return service.AuditLogFilter{}, false
		}
		filter.Limit = parsed
	}
	if raw := query.Get("userId"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			result.BadRequest(c, "invalid userId")
			return service.AuditLogFilter{}, false
		}
		filter.UserID = parsed
	}
	if raw := query.Get("cursor"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			result.BadRequest(c, "invalid cursor")
			return service.AuditLogFilter{}, false
		}
		filter.Cursor = parsed
	}
	if raw := query.Get("from"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			result.BadRequest(c, "invalid from")
			return service.AuditLogFilter{}, false
		}
		filter.From = parsed
	}
	if raw := query.Get("to"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			result.BadRequest(c, "invalid to")
			return service.AuditLogFilter{}, false
		}
		filter.To = parsed
	}
	return filter, true
}

func parseAPITokenFilter(c web.Context) (service.APITokenFilter, bool) {
	query := c.Request().URL.Query()
	filter := service.APITokenFilter{Status: query.Get("status")}
	if raw := query.Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid page")
			return service.APITokenFilter{}, false
		}
		filter.Page = parsed
	}
	if raw := query.Get("pageSize"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid pageSize")
			return service.APITokenFilter{}, false
		}
		filter.PageSize = parsed
	}
	if raw := query.Get("userId"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			result.BadRequest(c, "invalid userId")
			return service.APITokenFilter{}, false
		}
		filter.UserID = parsed
	}
	return filter, true
}

func parseUserListFilter(c web.Context) (service.UserListFilter, bool) {
	query := c.Request().URL.Query()
	filter := service.UserListFilter{
		Keyword:     query.Get("keyword"),
		Username:    query.Get("username"),
		DisplayName: firstNonEmpty(query.Get("displayName"), query.Get("nickName"), query.Get("nickname")),
		Email:       query.Get("email"),
		RoleCode:    query.Get("roleCode"),
		Status:      query.Get("status"),
		OrderKey:    query.Get("orderKey"),
		Desc:        query.Get("desc") == "true" || query.Get("desc") == "1",
	}
	if raw := query.Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid page")
			return service.UserListFilter{}, false
		}
		filter.Page = parsed
	}
	if raw := query.Get("pageSize"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid pageSize")
			return service.UserListFilter{}, false
		}
		filter.PageSize = parsed
	}
	return filter, true
}

func parseOrganizationListFilter(c web.Context) (service.OrganizationListFilter, bool) {
	query := c.Request().URL.Query()
	filter := service.OrganizationListFilter{
		Keyword:  query.Get("keyword"),
		Code:     query.Get("code"),
		Name:     query.Get("name"),
		Status:   query.Get("status"),
		OrderKey: query.Get("orderKey"),
		Desc:     query.Get("desc") == "true" || query.Get("desc") == "1",
	}
	if raw := query.Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid page")
			return service.OrganizationListFilter{}, false
		}
		filter.Page = parsed
	}
	if raw := query.Get("pageSize"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid pageSize")
			return service.OrganizationListFilter{}, false
		}
		filter.PageSize = parsed
	}
	return filter, true
}

func parseSessionListFilter(c web.Context) (service.SessionListFilter, bool) {
	query := c.Request().URL.Query()
	filter := service.SessionListFilter{
		Keyword:   query.Get("keyword"),
		IPAddress: firstNonEmpty(query.Get("ipAddress"), query.Get("ip")),
		Status:    query.Get("status"),
		Scope:     query.Get("scope"),
		OrderKey:  query.Get("orderKey"),
		Desc:      query.Get("desc") == "true" || query.Get("desc") == "1",
	}
	if raw := query.Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid page")
			return service.SessionListFilter{}, false
		}
		filter.Page = parsed
	}
	if raw := query.Get("pageSize"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			result.BadRequest(c, "invalid pageSize")
			return service.SessionListFilter{}, false
		}
		filter.PageSize = parsed
	}
	if raw := query.Get("userId"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			result.BadRequest(c, "invalid userId")
			return service.SessionListFilter{}, false
		}
		filter.UserID = parsed
	}
	return filter, true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
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

func (h *Handler) writeSetupError(c web.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput):
		result.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrSetupCompleted):
		result.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrDuplicate):
		result.BadRequest(c, err.Error())
	default:
		if h.logger != nil {
			h.logger.Error("iam setup failed", "error", err)
		}
		result.InternalError(c, "initial setup failed: "+err.Error())
	}
}

func (h *Handler) writeError(c web.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidInput), errors.Is(err, service.ErrCaptchaRequired), errors.Is(err, service.ErrCaptchaInvalid):
		result.BadRequest(c, err.Error())
	case errors.Is(err, service.ErrUnauthorized), errors.Is(err, service.ErrMFARequired), errors.Is(err, service.ErrInvalidToken), errors.Is(err, service.ErrAccountLocked), errors.Is(err, service.ErrAccountDisabled), errors.Is(err, service.ErrSessionRevoked):
		result.Unauthorized(c, err.Error())
	case errors.Is(err, service.ErrForbidden), errors.Is(err, service.ErrSignupDisabled), errors.Is(err, service.ErrSetupCompleted):
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
