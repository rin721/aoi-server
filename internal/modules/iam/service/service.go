package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/rei0721/go-scaffold/internal/modules/iam/model"
	"github.com/rei0721/go-scaffold/internal/modules/iam/repository"
	"github.com/rei0721/go-scaffold/pkg/authorization"
	passwordcrypto "github.com/rei0721/go-scaffold/pkg/crypto"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/mfa"
	"github.com/rei0721/go-scaffold/pkg/token"
	"github.com/rei0721/go-scaffold/pkg/utils"
)

var (
	ErrInvalidInput     = errors.New("invalid iam input")
	ErrUnauthorized     = errors.New("invalid credentials")
	ErrForbidden        = errors.New("permission denied")
	ErrNotFound         = errors.New("iam resource not found")
	ErrDuplicate        = errors.New("iam resource already exists")
	ErrMFARequired      = errors.New("mfa code required")
	ErrInvalidToken     = errors.New("invalid iam token")
	ErrAccountLocked    = errors.New("account locked")
	ErrAccountDisabled  = errors.New("account disabled")
	ErrSessionRevoked   = errors.New("session revoked")
	ErrInvitationClosed = errors.New("invitation is not available")
	ErrSignupDisabled   = errors.New("self signup is disabled")
	ErrSetupCompleted   = errors.New("setup already completed")
)

type Config struct {
	SelfSignupEnabled  bool
	MFAIssuer          string
	MFASecretKey       string
	LoginMaxFailures   int
	LoginLockDuration  time.Duration
	InvitationTTL      time.Duration
	PasswordResetTTL   time.Duration
	NotificationDriver string
	PublicBaseURL      string
	PasswordPolicy     PasswordPolicy
	Now                func() time.Time
}

type PasswordPolicy struct {
	MinLength     int
	RequireLower  bool
	RequireUpper  bool
	RequireNumber bool
	RequireSymbol bool
}

type Principal struct {
	UserID    int64  `json:"userId,string"`
	OrgID     int64  `json:"orgId,string"`
	SessionID int64  `json:"sessionId,string"`
	Username  string `json:"username"`
	Email     string `json:"email"`
}

type TokenPair struct {
	AccessToken      string    `json:"accessToken"`
	AccessExpiresAt  time.Time `json:"accessExpiresAt"`
	RefreshToken     string    `json:"refreshToken"`
	RefreshExpiresAt time.Time `json:"refreshExpiresAt"`
}

type LoginInput struct {
	Identifier string
	Password   string
	OrgCode    string
	MFACode    string
	UserAgent  string
	IPAddress  string
}

type SignupInput struct {
	OrgCode     string
	OrgName     string
	Username    string
	Email       string
	DisplayName string
	Password    string
	UserAgent   string
	IPAddress   string
}

type InitialAdminSetupInput struct {
	OrgCode     string
	OrgName     string
	Username    string
	Email       string
	DisplayName string
	Password    string
	UserAgent   string
	IPAddress   string
}

type RefreshInput struct {
	RefreshToken string
	UserAgent    string
	IPAddress    string
}

type BootstrapAdminInput struct {
	OrgCode     string
	OrgName     string
	Username    string
	Email       string
	DisplayName string
	Password    string
}

type InviteUserInput struct {
	Principal Principal
	Email     string
	RoleCode  string
	UserAgent string
	IPAddress string
}

type AcceptInvitationInput struct {
	Token       string
	Username    string
	DisplayName string
	Password    string
	UserAgent   string
	IPAddress   string
}

type ForgotPasswordInput struct {
	Email     string
	UserAgent string
	IPAddress string
}

type ResetPasswordInput struct {
	Token       string
	NewPassword string
	UserAgent   string
	IPAddress   string
}

type CreateRoleInput struct {
	Principal   Principal
	Code        string
	Name        string
	Description string
	Permissions []string
}

type UpdateUserInput struct {
	Principal Principal
	UserID    int64
	Status    *string
	Roles     []string
	HasRoles  bool
	UserAgent string
	IPAddress string
}

type UpdateRoleInput struct {
	Principal      Principal
	RoleID         int64
	Name           string
	Description    string
	Permissions    []string
	HasPermissions bool
	UserAgent      string
	IPAddress      string
}

type UpdateOrganizationInput struct {
	Principal Principal
	OrgID     int64
	Name      string
	UserAgent string
	IPAddress string
}

type AuditLogFilter = repository.AuditLogFilter

type OrganizationUser struct {
	User             model.User `json:"user"`
	MembershipStatus string     `json:"membershipStatus"`
	Roles            []string   `json:"roles"`
}

type NotificationDelivery struct {
	Token string `json:"token,omitempty"`
	URL   string `json:"url,omitempty"`
}

type SetupStatus struct {
	Required bool `json:"required"`
}

type Notifier interface {
	SendInvitation(context.Context, InvitationNotice) error
	SendPasswordReset(context.Context, PasswordResetNotice) error
}

type InvitationNotice struct {
	Email string
	Token string
	URL   string
}

type PasswordResetNotice struct {
	Email string
	Token string
	URL   string
}

type NoopNotifier struct{}

func (NoopNotifier) SendInvitation(context.Context, InvitationNotice) error       { return nil }
func (NoopNotifier) SendPasswordReset(context.Context, PasswordResetNotice) error { return nil }

type Service interface {
	BootstrapAdmin(context.Context, BootstrapAdminInput) (*Principal, error)
	SetupStatus(context.Context) (SetupStatus, error)
	InitialAdminSetup(context.Context, InitialAdminSetupInput) (TokenPair, error)
	Signup(context.Context, SignupInput) (TokenPair, error)
	Login(context.Context, LoginInput) (TokenPair, error)
	Refresh(context.Context, RefreshInput) (TokenPair, error)
	Logout(context.Context, Principal) error
	SwitchOrg(context.Context, Principal, int64, string, string) (TokenPair, error)
	AuthenticateToken(context.Context, string) (Principal, error)
	Authorize(context.Context, Principal, string, string) (bool, error)
	Me(context.Context, Principal) (*model.User, error)
	ListMyOrganizations(context.Context, Principal) ([]model.Organization, error)
	ListOrganizations(context.Context, Principal) ([]model.Organization, error)
	CreateOrganization(context.Context, Principal, string, string) (*model.Organization, error)
	UpdateOrganization(context.Context, UpdateOrganizationInput) (*model.Organization, error)
	InviteUser(context.Context, InviteUserInput) (NotificationDelivery, error)
	ListInvitations(context.Context, Principal) ([]model.Invitation, error)
	RevokeInvitation(context.Context, Principal, int64, string, string) error
	AcceptInvitation(context.Context, AcceptInvitationInput) (*Principal, error)
	ForgotPassword(context.Context, ForgotPasswordInput) (NotificationDelivery, error)
	ResetPassword(context.Context, ResetPasswordInput) error
	SetupMFA(context.Context, Principal) (string, string, error)
	VerifyMFA(context.Context, Principal, string) error
	ListUsers(context.Context, Principal) ([]OrganizationUser, error)
	UpdateUser(context.Context, UpdateUserInput) (*OrganizationUser, error)
	ListRoles(context.Context, Principal) ([]model.Role, error)
	CreateRole(context.Context, CreateRoleInput) (*model.Role, error)
	UpdateRole(context.Context, UpdateRoleInput) (*model.Role, error)
	ListPermissions(context.Context, Principal) ([]model.Permission, error)
	ListSessions(context.Context, Principal, int64) ([]model.Session, error)
	RevokeSession(context.Context, Principal, int64) error
	ListAuditLogs(context.Context, Principal, AuditLogFilter) ([]model.AuditLog, error)
	RecordAudit(context.Context, Principal, string, string, string, string, string, map[string]any) error
	LoadPolicies(context.Context) error
}

type service struct {
	db       database.Database
	repo     repository.Repository
	crypto   passwordcrypto.Crypto
	tokens   token.Manager
	authz    authorization.Enforcer
	ids      utils.IDGenerator
	cfg      Config
	notifier Notifier
}

func New(db database.Database, repo repository.Repository, crypto passwordcrypto.Crypto, tokens token.Manager, authz authorization.Enforcer, ids utils.IDGenerator, cfg Config, notifier Notifier) Service {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.LoginMaxFailures <= 0 {
		cfg.LoginMaxFailures = 5
	}
	if cfg.LoginLockDuration <= 0 {
		cfg.LoginLockDuration = 15 * time.Minute
	}
	if cfg.InvitationTTL <= 0 {
		cfg.InvitationTTL = 24 * time.Hour
	}
	if cfg.PasswordResetTTL <= 0 {
		cfg.PasswordResetTTL = 30 * time.Minute
	}
	if cfg.NotificationDriver == "" {
		cfg.NotificationDriver = "debug"
	}
	if cfg.PasswordPolicy.MinLength <= 0 {
		cfg.PasswordPolicy.MinLength = 8
	}
	if cfg.MFAIssuer == "" {
		cfg.MFAIssuer = "go-scaffold"
	}
	if notifier == nil {
		notifier = NoopNotifier{}
	}
	return &service{db: db, repo: repo, crypto: crypto, tokens: tokens, authz: authz, ids: ids, cfg: cfg, notifier: notifier}
}

func (s *service) BootstrapAdmin(ctx context.Context, input BootstrapAdminInput) (*Principal, error) {
	input.OrgCode = normalizeCode(input.OrgCode)
	input.Username = normalizeCode(input.Username)
	input.Email = normalizeEmail(input.Email)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.OrgCode == "" || input.Username == "" || input.Email == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}
	if input.OrgName == "" {
		input.OrgName = input.OrgCode
	}
	if input.DisplayName == "" {
		input.DisplayName = input.Username
	}

	var principal *Principal
	err := s.db.WithTx(ctx, func(txCtx context.Context, tx database.Executor) error {
		repo := s.repo.WithExecutor(tx)
		org, err := repo.FindOrganizationByCode(txCtx, input.OrgCode)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return err
			}
			now := s.now()
			org = &model.Organization{ID: s.ids.NextID(), Code: input.OrgCode, Name: input.OrgName, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now}
			if err := repo.CreateOrganization(txCtx, org); err != nil {
				return err
			}
		}

		user, err := repo.FindUserByIdentifier(txCtx, input.Email)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return err
			}
			hash, err := s.crypto.HashPassword(input.Password)
			if err != nil {
				return err
			}
			now := s.now()
			user = &model.User{ID: s.ids.NextID(), Username: input.Username, Email: input.Email, PasswordHash: hash, DisplayName: input.DisplayName, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now}
			if err := repo.CreateUser(txCtx, user); err != nil {
				return err
			}
		}
		if err := s.ensureMembership(txCtx, repo, org.ID, user.ID); err != nil {
			return err
		}
		if err := s.ensureBuiltins(txCtx, repo, org.ID); err != nil {
			return err
		}
		if err := s.addUserRole(txCtx, repo, user.ID, org.ID, model.RoleOwner); err != nil {
			return err
		}
		if err := s.audit(txCtx, repo, &org.ID, &user.ID, "iam.bootstrap_admin", "organization", strconv.FormatInt(org.ID, 10), "", "", nil); err != nil {
			return err
		}
		principal = &Principal{UserID: user.ID, OrgID: org.ID, Username: user.Username, Email: user.Email}
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = s.LoadPolicies(ctx)
	return principal, nil
}

func (s *service) SetupStatus(ctx context.Context) (SetupStatus, error) {
	count, err := s.repo.CountUsers(ctx)
	if err != nil {
		return SetupStatus{}, err
	}
	return SetupStatus{Required: count == 0}, nil
}

func (s *service) InitialAdminSetup(ctx context.Context, input InitialAdminSetupInput) (TokenPair, error) {
	input.OrgCode = normalizeCode(input.OrgCode)
	input.OrgName = strings.TrimSpace(input.OrgName)
	input.Username = normalizeCode(input.Username)
	input.Email = normalizeEmail(input.Email)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.OrgCode == "" || input.OrgName == "" || input.Username == "" || input.Email == "" || input.Password == "" {
		return TokenPair{}, ErrInvalidInput
	}
	if input.DisplayName == "" {
		input.DisplayName = input.Username
	}
	if err := s.validatePassword(input.Password); err != nil {
		return TokenPair{}, err
	}

	var pair TokenPair
	err := s.db.WithTx(ctx, func(txCtx context.Context, tx database.Executor) error {
		repo := s.repo.WithExecutor(tx)
		count, err := repo.CountUsers(txCtx)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrSetupCompleted
		}
		if _, err := repo.FindOrganizationByCode(txCtx, input.OrgCode); err == nil {
			return ErrDuplicate
		} else if !errors.Is(err, database.ErrNotFound) {
			return err
		}

		hash, err := s.crypto.HashPassword(input.Password)
		if err != nil {
			return err
		}
		now := s.now()
		org := &model.Organization{ID: s.ids.NextID(), Code: input.OrgCode, Name: input.OrgName, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now}
		if err := repo.CreateOrganization(txCtx, org); err != nil {
			return err
		}
		user := &model.User{ID: s.ids.NextID(), Username: input.Username, Email: input.Email, PasswordHash: hash, DisplayName: input.DisplayName, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now}
		if err := repo.CreateUser(txCtx, user); err != nil {
			return err
		}
		if err := s.ensureMembership(txCtx, repo, org.ID, user.ID); err != nil {
			return err
		}
		if err := s.ensureBuiltins(txCtx, repo, org.ID); err != nil {
			return err
		}
		if err := s.addUserRole(txCtx, repo, user.ID, org.ID, model.RoleOwner); err != nil {
			return err
		}
		issued, err := s.createSessionAndTokensWithRepo(txCtx, repo, user, org.ID, input.UserAgent, input.IPAddress)
		if err != nil {
			return err
		}
		pair = issued
		return s.audit(txCtx, repo, &org.ID, &user.ID, "iam.initial_setup", "organization", strconv.FormatInt(org.ID, 10), input.IPAddress, input.UserAgent, map[string]any{"orgCode": org.Code, "email": user.Email})
	})
	if err != nil {
		return TokenPair{}, err
	}
	_ = s.LoadPolicies(ctx)
	return pair, nil
}

func (s *service) Signup(ctx context.Context, input SignupInput) (TokenPair, error) {
	if !s.cfg.SelfSignupEnabled {
		return TokenPair{}, ErrSignupDisabled
	}
	input.OrgCode = normalizeCode(input.OrgCode)
	input.OrgName = strings.TrimSpace(input.OrgName)
	input.Username = normalizeCode(input.Username)
	input.Email = normalizeEmail(input.Email)
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	if input.OrgCode == "" || input.OrgName == "" || input.Username == "" || input.Email == "" || input.Password == "" {
		return TokenPair{}, ErrInvalidInput
	}
	if input.DisplayName == "" {
		input.DisplayName = input.Username
	}
	if err := s.validatePassword(input.Password); err != nil {
		return TokenPair{}, err
	}

	var pair TokenPair
	err := s.db.WithTx(ctx, func(txCtx context.Context, tx database.Executor) error {
		repo := s.repo.WithExecutor(tx)
		if _, err := repo.FindOrganizationByCode(txCtx, input.OrgCode); err == nil {
			return ErrDuplicate
		} else if !errors.Is(err, database.ErrNotFound) {
			return err
		}
		if _, err := repo.FindUserByIdentifier(txCtx, input.Username); err == nil {
			return ErrDuplicate
		} else if !errors.Is(err, database.ErrNotFound) {
			return err
		}
		if input.Email != input.Username {
			if _, err := repo.FindUserByIdentifier(txCtx, input.Email); err == nil {
				return ErrDuplicate
			} else if !errors.Is(err, database.ErrNotFound) {
				return err
			}
		}

		hash, err := s.crypto.HashPassword(input.Password)
		if err != nil {
			return err
		}
		now := s.now()
		org := &model.Organization{ID: s.ids.NextID(), Code: input.OrgCode, Name: input.OrgName, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now}
		if err := repo.CreateOrganization(txCtx, org); err != nil {
			return err
		}
		user := &model.User{ID: s.ids.NextID(), Username: input.Username, Email: input.Email, PasswordHash: hash, DisplayName: input.DisplayName, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now}
		if err := repo.CreateUser(txCtx, user); err != nil {
			return err
		}
		if err := s.ensureMembership(txCtx, repo, org.ID, user.ID); err != nil {
			return err
		}
		if err := s.ensureBuiltins(txCtx, repo, org.ID); err != nil {
			return err
		}
		if err := s.addUserRole(txCtx, repo, user.ID, org.ID, model.RoleOwner); err != nil {
			return err
		}
		issued, err := s.createSessionAndTokensWithRepo(txCtx, repo, user, org.ID, input.UserAgent, input.IPAddress)
		if err != nil {
			return err
		}
		pair = issued
		return s.audit(txCtx, repo, &org.ID, &user.ID, "auth.signup", "organization", strconv.FormatInt(org.ID, 10), input.IPAddress, input.UserAgent, map[string]any{"orgCode": org.Code, "email": user.Email})
	})
	if err != nil {
		return TokenPair{}, err
	}
	_ = s.LoadPolicies(ctx)
	return pair, nil
}

func (s *service) Login(ctx context.Context, input LoginInput) (TokenPair, error) {
	identifier := strings.TrimSpace(strings.ToLower(input.Identifier))
	if identifier == "" || input.Password == "" {
		return TokenPair{}, ErrUnauthorized
	}
	user, err := s.repo.FindUserByIdentifier(ctx, identifier)
	if err != nil {
		return TokenPair{}, ErrUnauthorized
	}
	if err := s.ensureUserCanLogin(user); err != nil {
		return TokenPair{}, err
	}
	if err := s.crypto.VerifyPassword(user.PasswordHash, input.Password); err != nil {
		_ = s.recordFailedLogin(ctx, user)
		return TokenPair{}, ErrUnauthorized
	}
	if user.MFAEnabled {
		if input.MFACode == "" {
			return TokenPair{}, ErrMFARequired
		}
		if err := s.verifyUserMFA(ctx, user.ID, input.MFACode); err != nil {
			return TokenPair{}, ErrUnauthorized
		}
	}
	org, err := s.loginOrg(ctx, user.ID, input.OrgCode)
	if err != nil {
		return TokenPair{}, err
	}
	pair, err := s.createSessionAndTokens(ctx, user, org.ID, input.UserAgent, input.IPAddress)
	if err != nil {
		return TokenPair{}, err
	}
	now := s.now()
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil
	user.LastLoginAt = &now
	_ = s.repo.SaveUser(ctx, user)
	_ = s.audit(ctx, s.repo, &org.ID, &user.ID, "auth.login", "session", "", input.IPAddress, input.UserAgent, nil)
	return pair, nil
}

func (s *service) Refresh(ctx context.Context, input RefreshInput) (TokenPair, error) {
	hash := s.tokens.HashRefreshToken(strings.TrimSpace(input.RefreshToken))
	session, err := s.repo.FindSessionByRefreshHash(ctx, hash)
	if err != nil {
		return TokenPair{}, ErrInvalidToken
	}
	if err := s.ensureSessionActive(session); err != nil {
		return TokenPair{}, err
	}
	user, err := s.repo.FindUserByID(ctx, session.UserID)
	if err != nil {
		return TokenPair{}, ErrInvalidToken
	}
	if err := s.ensureUserCanLogin(user); err != nil {
		return TokenPair{}, err
	}
	if _, err := s.repo.FindMembership(ctx, session.OrgID, session.UserID); err != nil {
		return TokenPair{}, ErrForbidden
	}
	pair, err := s.tokens.IssuePair(ctx, token.Subject{UserID: user.ID, OrgID: session.OrgID, SessionID: session.ID})
	if err != nil {
		return TokenPair{}, err
	}
	now := s.now()
	session.RefreshTokenHash = pair.RefreshTokenHash
	session.ExpiresAt = pair.RefreshExpiresAt
	session.LastUsedAt = &now
	session.UserAgent = input.UserAgent
	session.IPAddress = input.IPAddress
	if err := s.repo.SaveSession(ctx, session); err != nil {
		return TokenPair{}, err
	}
	return tokenPair(pair), nil
}

func (s *service) Logout(ctx context.Context, principal Principal) error {
	session, err := s.repo.FindSessionByID(ctx, principal.SessionID)
	if err != nil {
		return ErrInvalidToken
	}
	now := s.now()
	session.RevokedAt = &now
	return s.repo.SaveSession(ctx, session)
}

func (s *service) SwitchOrg(ctx context.Context, principal Principal, orgID int64, userAgent, ip string) (TokenPair, error) {
	if _, err := s.repo.FindMembership(ctx, orgID, principal.UserID); err != nil {
		return TokenPair{}, ErrForbidden
	}
	user, err := s.repo.FindUserByID(ctx, principal.UserID)
	if err != nil {
		return TokenPair{}, ErrInvalidToken
	}
	return s.createSessionAndTokens(ctx, user, orgID, userAgent, ip)
}

func (s *service) AuthenticateToken(ctx context.Context, raw string) (Principal, error) {
	claims, err := s.tokens.Parse(ctx, raw, token.TokenTypeAccess)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	session, err := s.repo.FindSessionByID(ctx, claims.SessionID)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	if err := s.ensureSessionActive(session); err != nil {
		return Principal{}, err
	}
	user, err := s.repo.FindUserByID(ctx, claims.UserID)
	if err != nil {
		return Principal{}, ErrInvalidToken
	}
	if err := s.ensureUserCanLogin(user); err != nil {
		return Principal{}, err
	}
	if _, err := s.repo.FindMembership(ctx, claims.OrgID, claims.UserID); err != nil {
		return Principal{}, ErrForbidden
	}
	return Principal{UserID: user.ID, OrgID: claims.OrgID, SessionID: session.ID, Username: user.Username, Email: user.Email}, nil
}

func (s *service) Authorize(ctx context.Context, p Principal, obj, act string) (bool, error) {
	return s.authz.Enforce(ctx, userSubject(p.UserID), strconv.FormatInt(p.OrgID, 10), obj, act)
}

func (s *service) Me(ctx context.Context, p Principal) (*model.User, error) {
	return s.repo.FindUserByID(ctx, p.UserID)
}

func (s *service) ListMyOrganizations(ctx context.Context, p Principal) ([]model.Organization, error) {
	memberships, err := s.repo.ListMembershipsByUser(ctx, p.UserID)
	if err != nil {
		return nil, err
	}
	orgs := make([]model.Organization, 0, len(memberships))
	for _, membership := range memberships {
		org, err := s.repo.FindOrganizationByID(ctx, membership.OrgID)
		if err == nil {
			orgs = append(orgs, *org)
		}
	}
	return orgs, nil
}

func (s *service) ListOrganizations(ctx context.Context, _ Principal) ([]model.Organization, error) {
	return s.repo.ListOrganizations(ctx)
}

func (s *service) CreateOrganization(ctx context.Context, p Principal, code, name string) (*model.Organization, error) {
	code = normalizeCode(code)
	name = strings.TrimSpace(name)
	if code == "" || name == "" {
		return nil, ErrInvalidInput
	}
	var org *model.Organization
	err := s.db.WithTx(ctx, func(txCtx context.Context, tx database.Executor) error {
		repo := s.repo.WithExecutor(tx)
		if _, err := repo.FindOrganizationByCode(txCtx, code); err == nil {
			return ErrDuplicate
		} else if !errors.Is(err, database.ErrNotFound) {
			return err
		}
		now := s.now()
		org = &model.Organization{ID: s.ids.NextID(), Code: code, Name: name, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now}
		if err := repo.CreateOrganization(txCtx, org); err != nil {
			return err
		}
		if err := s.ensureMembership(txCtx, repo, org.ID, p.UserID); err != nil {
			return err
		}
		if err := s.ensureBuiltins(txCtx, repo, org.ID); err != nil {
			return err
		}
		if err := s.addUserRole(txCtx, repo, p.UserID, org.ID, model.RoleOwner); err != nil {
			return err
		}
		return s.audit(txCtx, repo, &org.ID, &p.UserID, "org.create", "organization", strconv.FormatInt(org.ID, 10), "", "", nil)
	})
	if err != nil {
		return nil, err
	}
	_ = s.LoadPolicies(ctx)
	return org, nil
}

func (s *service) UpdateOrganization(ctx context.Context, input UpdateOrganizationInput) (*model.Organization, error) {
	if input.OrgID != input.Principal.OrgID {
		return nil, ErrForbidden
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, ErrInvalidInput
	}
	org, err := s.repo.FindOrganizationByID(ctx, input.OrgID)
	if err != nil {
		return nil, ErrNotFound
	}
	org.Name = name
	if err := s.repo.SaveOrganization(ctx, org); err != nil {
		return nil, err
	}
	_ = s.audit(ctx, s.repo, &input.Principal.OrgID, &input.Principal.UserID, "org.update", "organization", strconv.FormatInt(org.ID, 10), input.IPAddress, input.UserAgent, map[string]any{"name": name})
	return org, nil
}

func (s *service) InviteUser(ctx context.Context, input InviteUserInput) (NotificationDelivery, error) {
	email := normalizeEmail(input.Email)
	roleCode := normalizeCode(input.RoleCode)
	if email == "" || roleCode == "" {
		return NotificationDelivery{}, ErrInvalidInput
	}
	if _, err := s.repo.FindRole(ctx, input.Principal.OrgID, roleCode); err != nil {
		return NotificationDelivery{}, ErrNotFound
	}
	raw, hash, err := s.oneTimeToken()
	if err != nil {
		return NotificationDelivery{}, err
	}
	now := s.now()
	invitation := &model.Invitation{
		ID: s.ids.NextID(), OrgID: input.Principal.OrgID, Email: email, RoleCode: roleCode, TokenHash: hash,
		Status: model.StatusPending, InvitedBy: input.Principal.UserID, ExpiresAt: now.Add(s.cfg.InvitationTTL), CreatedAt: now, UpdatedAt: now,
	}
	if err := s.repo.CreateInvitation(ctx, invitation); err != nil {
		return NotificationDelivery{}, err
	}
	adminPath := "invitations/" + url.PathEscape(raw)
	_ = s.notifier.SendInvitation(ctx, InvitationNotice{Email: email, Token: raw, URL: s.notificationURL(adminPath)})
	_ = s.audit(ctx, s.repo, &input.Principal.OrgID, &input.Principal.UserID, "user.invite", "invitation", strconv.FormatInt(invitation.ID, 10), input.IPAddress, input.UserAgent, map[string]any{"email": email})
	return s.debugDelivery(raw, adminPath), nil
}

func (s *service) ListInvitations(ctx context.Context, p Principal) ([]model.Invitation, error) {
	return s.repo.ListInvitationsByOrg(ctx, p.OrgID)
}

func (s *service) RevokeInvitation(ctx context.Context, p Principal, invitationID int64, userAgent, ip string) error {
	invitation, err := s.repo.FindInvitationByID(ctx, invitationID)
	if err != nil {
		return ErrNotFound
	}
	if invitation.OrgID != p.OrgID {
		return ErrForbidden
	}
	if invitation.Status != model.StatusPending {
		return ErrInvitationClosed
	}
	invitation.Status = model.StatusRevoked
	if err := s.repo.SaveInvitation(ctx, invitation); err != nil {
		return err
	}
	return s.audit(ctx, s.repo, &p.OrgID, &p.UserID, "invite.revoke", "invitation", strconv.FormatInt(invitation.ID, 10), ip, userAgent, map[string]any{"email": invitation.Email})
}

func (s *service) AcceptInvitation(ctx context.Context, input AcceptInvitationInput) (*Principal, error) {
	hash := s.tokens.HashRefreshToken(strings.TrimSpace(input.Token))
	invitation, err := s.repo.FindInvitationByTokenHash(ctx, hash)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if invitation.Status != model.StatusPending || invitation.ExpiresAt.Before(s.now()) {
		return nil, ErrInvitationClosed
	}
	username := normalizeCode(input.Username)
	email := normalizeEmail(invitation.Email)
	displayName := strings.TrimSpace(input.DisplayName)
	if username == "" || input.Password == "" {
		return nil, ErrInvalidInput
	}
	if err := s.validatePassword(input.Password); err != nil {
		return nil, err
	}
	if displayName == "" {
		displayName = username
	}
	var principal *Principal
	err = s.db.WithTx(ctx, func(txCtx context.Context, tx database.Executor) error {
		repo := s.repo.WithExecutor(tx)
		user, err := repo.FindUserByIdentifier(txCtx, email)
		if err != nil {
			if !errors.Is(err, database.ErrNotFound) {
				return err
			}
			hash, err := s.crypto.HashPassword(input.Password)
			if err != nil {
				return err
			}
			now := s.now()
			user = &model.User{ID: s.ids.NextID(), Username: username, Email: email, PasswordHash: hash, DisplayName: displayName, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now}
			if err := repo.CreateUser(txCtx, user); err != nil {
				return err
			}
		}
		if err := s.ensureMembership(txCtx, repo, invitation.OrgID, user.ID); err != nil {
			return err
		}
		if err := s.addUserRole(txCtx, repo, user.ID, invitation.OrgID, invitation.RoleCode); err != nil {
			return err
		}
		now := s.now()
		invitation.Status = model.StatusUsed
		invitation.AcceptedBy = &user.ID
		invitation.UpdatedAt = now
		if err := repo.SaveInvitation(txCtx, invitation); err != nil {
			return err
		}
		principal = &Principal{UserID: user.ID, OrgID: invitation.OrgID, Username: user.Username, Email: user.Email}
		return s.audit(txCtx, repo, &invitation.OrgID, &user.ID, "invitation.accept", "invitation", strconv.FormatInt(invitation.ID, 10), input.IPAddress, input.UserAgent, nil)
	})
	if err != nil {
		return nil, err
	}
	_ = s.LoadPolicies(ctx)
	return principal, nil
}

func (s *service) ForgotPassword(ctx context.Context, input ForgotPasswordInput) (NotificationDelivery, error) {
	user, err := s.repo.FindUserByIdentifier(ctx, normalizeEmail(input.Email))
	if err != nil {
		return NotificationDelivery{}, nil
	}
	raw, hash, err := s.oneTimeToken()
	if err != nil {
		return NotificationDelivery{}, err
	}
	now := s.now()
	reset := &model.PasswordReset{ID: s.ids.NextID(), UserID: user.ID, TokenHash: hash, Status: model.StatusPending, ExpiresAt: now.Add(s.cfg.PasswordResetTTL), CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreatePasswordReset(ctx, reset); err != nil {
		return NotificationDelivery{}, err
	}
	adminPath := "password/reset?token=" + url.QueryEscape(raw)
	_ = s.notifier.SendPasswordReset(ctx, PasswordResetNotice{Email: user.Email, Token: raw, URL: s.notificationURL(adminPath)})
	_ = s.audit(ctx, s.repo, nil, &user.ID, "password.forgot", "password_reset", strconv.FormatInt(reset.ID, 10), input.IPAddress, input.UserAgent, nil)
	return s.debugDelivery(raw, adminPath), nil
}

func (s *service) ResetPassword(ctx context.Context, input ResetPasswordInput) error {
	reset, err := s.repo.FindPasswordResetByTokenHash(ctx, s.tokens.HashRefreshToken(strings.TrimSpace(input.Token)))
	if err != nil {
		return ErrInvalidToken
	}
	if reset.Status != model.StatusPending || reset.ExpiresAt.Before(s.now()) {
		return ErrInvalidToken
	}
	if err := s.validatePassword(input.NewPassword); err != nil {
		return err
	}
	user, err := s.repo.FindUserByID(ctx, reset.UserID)
	if err != nil {
		return ErrInvalidToken
	}
	hash, err := s.crypto.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hash
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil
	if err := s.repo.SaveUser(ctx, user); err != nil {
		return err
	}
	now := s.now()
	reset.Status = model.StatusUsed
	reset.UsedAt = &now
	if err := s.repo.SavePasswordReset(ctx, reset); err != nil {
		return err
	}
	sessions, err := s.repo.ListSessionsByUser(ctx, user.ID)
	if err != nil {
		return err
	}
	for i := range sessions {
		if sessions[i].RevokedAt != nil {
			continue
		}
		sessions[i].RevokedAt = &now
		if err := s.repo.SaveSession(ctx, &sessions[i]); err != nil {
			return err
		}
	}
	return s.audit(ctx, s.repo, nil, &user.ID, "password.reset", "password_reset", strconv.FormatInt(reset.ID, 10), input.IPAddress, input.UserAgent, nil)
}

func (s *service) SetupMFA(ctx context.Context, p Principal) (string, string, error) {
	user, err := s.repo.FindUserByID(ctx, p.UserID)
	if err != nil {
		return "", "", ErrInvalidToken
	}
	key, err := mfa.GenerateTOTP(s.cfg.MFAIssuer, user.Email)
	if err != nil {
		return "", "", err
	}
	encrypted, err := s.encryptSecret(key.Secret)
	if err != nil {
		return "", "", err
	}
	now := s.now()
	factor, err := s.repo.FindActiveMFAFactor(ctx, user.ID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return "", "", err
	}
	if err == nil {
		factor.Secret = encrypted
		factor.ConfirmedAt = nil
		factor.UpdatedAt = now
		if err := s.repo.SaveMFAFactor(ctx, factor); err != nil {
			return "", "", err
		}
		if err := s.audit(ctx, s.repo, &p.OrgID, &p.UserID, "mfa.setup", "mfa_factor", strconv.FormatInt(factor.ID, 10), "", "", nil); err != nil {
			return "", "", err
		}
		return key.Secret, key.URL, nil
	}
	factor = &model.MFAFactor{ID: s.ids.NextID(), UserID: user.ID, Type: "totp", Secret: encrypted, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateMFAFactor(ctx, factor); err != nil {
		return "", "", err
	}
	if err := s.audit(ctx, s.repo, &p.OrgID, &p.UserID, "mfa.setup", "mfa_factor", strconv.FormatInt(factor.ID, 10), "", "", nil); err != nil {
		return "", "", err
	}
	return key.Secret, key.URL, nil
}

func (s *service) VerifyMFA(ctx context.Context, p Principal, code string) error {
	if err := s.verifyUserMFA(ctx, p.UserID, code); err != nil {
		return err
	}
	user, err := s.repo.FindUserByID(ctx, p.UserID)
	if err != nil {
		return err
	}
	now := s.now()
	user.MFAEnabled = true
	if err := s.repo.SaveUser(ctx, user); err != nil {
		return err
	}
	factor, err := s.repo.FindActiveMFAFactor(ctx, p.UserID)
	if err == nil {
		factor.ConfirmedAt = &now
		_ = s.repo.SaveMFAFactor(ctx, factor)
	}
	return s.audit(ctx, s.repo, &p.OrgID, &p.UserID, "mfa.verify", "mfa_factor", "", "", "", nil)
}

func (s *service) ListUsers(ctx context.Context, p Principal) ([]OrganizationUser, error) {
	memberships, err := s.repo.ListMembershipsByOrg(ctx, p.OrgID)
	if err != nil {
		return nil, err
	}
	out := make([]OrganizationUser, 0, len(memberships))
	for _, membership := range memberships {
		user, err := s.repo.FindUserByID(ctx, membership.UserID)
		if err != nil {
			continue
		}
		roles, _ := s.authz.GetRolesForUser(ctx, userSubject(user.ID), strconv.FormatInt(p.OrgID, 10))
		out = append(out, OrganizationUser{User: *user, MembershipStatus: membership.Status, Roles: roles})
	}
	return out, nil
}

func (s *service) UpdateUser(ctx context.Context, input UpdateUserInput) (*OrganizationUser, error) {
	membership, err := s.repo.FindMembershipAnyStatus(ctx, input.Principal.OrgID, input.UserID)
	if err != nil {
		return nil, ErrNotFound
	}
	user, err := s.repo.FindUserByID(ctx, input.UserID)
	if err != nil {
		return nil, ErrNotFound
	}
	if input.Status != nil {
		status := normalizeCode(*input.Status)
		switch status {
		case model.StatusActive, model.StatusDisabled:
			membership.Status = status
			if err := s.repo.SaveMembership(ctx, membership); err != nil {
				return nil, err
			}
			_ = s.audit(ctx, s.repo, &input.Principal.OrgID, &input.Principal.UserID, "user.update", "membership", strconv.FormatInt(membership.ID, 10), input.IPAddress, input.UserAgent, map[string]any{"userId": input.UserID, "status": status})
		default:
			return nil, ErrInvalidInput
		}
	}
	if input.HasRoles {
		for _, roleCode := range input.Roles {
			if _, err := s.repo.FindRole(ctx, input.Principal.OrgID, normalizeCode(roleCode)); err != nil {
				return nil, ErrNotFound
			}
		}
		if err := s.repo.DeleteCasbinRules(ctx, "g", userSubject(input.UserID), "", strconv.FormatInt(input.Principal.OrgID, 10)); err != nil {
			return nil, err
		}
		for _, roleCode := range input.Roles {
			if err := s.addUserRole(ctx, s.repo, input.UserID, input.Principal.OrgID, roleCode); err != nil {
				return nil, err
			}
		}
		_ = s.LoadPolicies(ctx)
		_ = s.audit(ctx, s.repo, &input.Principal.OrgID, &input.Principal.UserID, "user.roles.update", "user", strconv.FormatInt(input.UserID, 10), input.IPAddress, input.UserAgent, map[string]any{"roles": input.Roles})
	}
	roles, _ := s.authz.GetRolesForUser(ctx, userSubject(user.ID), strconv.FormatInt(input.Principal.OrgID, 10))
	return &OrganizationUser{User: *user, MembershipStatus: membership.Status, Roles: roles}, nil
}

func (s *service) ListRoles(ctx context.Context, p Principal) ([]model.Role, error) {
	roles, err := s.repo.ListRoles(ctx, p.OrgID)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		_ = s.hydrateRolePermissions(ctx, &roles[i])
	}
	return roles, nil
}

func (s *service) CreateRole(ctx context.Context, input CreateRoleInput) (*model.Role, error) {
	code := normalizeCode(input.Code)
	name := strings.TrimSpace(input.Name)
	if code == "" || name == "" {
		return nil, ErrInvalidInput
	}
	now := s.now()
	role := &model.Role{ID: s.ids.NextID(), OrgID: input.Principal.OrgID, Code: code, Name: name, Description: strings.TrimSpace(input.Description), CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	for _, permission := range input.Permissions {
		obj, act := permissionObjectAction(permission)
		if obj == "" || act == "" {
			continue
		}
		if err := s.addPolicy(ctx, s.repo, input.Principal.OrgID, code, obj, act); err != nil {
			return nil, err
		}
	}
	_ = s.LoadPolicies(ctx)
	_ = s.hydrateRolePermissions(ctx, role)
	return role, nil
}

func (s *service) UpdateRole(ctx context.Context, input UpdateRoleInput) (*model.Role, error) {
	role, err := s.repo.FindRoleByID(ctx, input.RoleID)
	if err != nil {
		return nil, ErrNotFound
	}
	if role.OrgID != input.Principal.OrgID {
		return nil, ErrForbidden
	}
	if role.System {
		return nil, ErrForbidden
	}
	name := strings.TrimSpace(input.Name)
	if name != "" {
		role.Name = name
	}
	role.Description = strings.TrimSpace(input.Description)
	if err := s.repo.SaveRole(ctx, role); err != nil {
		return nil, err
	}
	if input.HasPermissions {
		if err := s.repo.DeleteCasbinRules(ctx, "p", roleSubject(role.Code), strconv.FormatInt(input.Principal.OrgID, 10)); err != nil {
			return nil, err
		}
		for _, permission := range input.Permissions {
			obj, act := permissionObjectAction(permission)
			if obj == "" || act == "" {
				continue
			}
			if err := s.addPolicy(ctx, s.repo, input.Principal.OrgID, role.Code, obj, act); err != nil {
				return nil, err
			}
		}
		_ = s.LoadPolicies(ctx)
	}
	_ = s.hydrateRolePermissions(ctx, role)
	_ = s.audit(ctx, s.repo, &input.Principal.OrgID, &input.Principal.UserID, "role.update", "role", strconv.FormatInt(role.ID, 10), input.IPAddress, input.UserAgent, map[string]any{"permissions": input.Permissions})
	return role, nil
}

func (s *service) ListPermissions(ctx context.Context, _ Principal) ([]model.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *service) ListSessions(ctx context.Context, p Principal, userID int64) ([]model.Session, error) {
	if userID == 0 {
		userID = p.UserID
	}
	return s.repo.ListSessionsByUser(ctx, userID)
}

func (s *service) RevokeSession(ctx context.Context, p Principal, sessionID int64) error {
	session, err := s.repo.FindSessionByID(ctx, sessionID)
	if err != nil {
		return ErrNotFound
	}
	if session.OrgID != p.OrgID {
		return ErrForbidden
	}
	now := s.now()
	session.RevokedAt = &now
	return s.repo.SaveSession(ctx, session)
}

func (s *service) ListAuditLogs(ctx context.Context, p Principal, filter AuditLogFilter) ([]model.AuditLog, error) {
	return s.repo.ListAuditLogs(ctx, p.OrgID, filter)
}

func (s *service) RecordAudit(ctx context.Context, p Principal, action, resource, resourceID, ip, userAgent string, metadata map[string]any) error {
	return s.audit(ctx, s.repo, &p.OrgID, &p.UserID, action, resource, resourceID, ip, userAgent, metadata)
}

func (s *service) LoadPolicies(ctx context.Context) error {
	rules, err := s.repo.ListCasbinRules(ctx)
	if err != nil {
		return err
	}
	return s.authz.LoadRules(ctx, rules)
}

func (s *service) createSessionAndTokens(ctx context.Context, user *model.User, orgID int64, userAgent, ip string) (TokenPair, error) {
	return s.createSessionAndTokensWithRepo(ctx, s.repo, user, orgID, userAgent, ip)
}

func (s *service) createSessionAndTokensWithRepo(ctx context.Context, repo repository.Repository, user *model.User, orgID int64, userAgent, ip string) (TokenPair, error) {
	now := s.now()
	sessionID := s.ids.NextID()
	pair, err := s.tokens.IssuePair(ctx, token.Subject{UserID: user.ID, OrgID: orgID, SessionID: sessionID})
	if err != nil {
		return TokenPair{}, err
	}
	session := &model.Session{ID: sessionID, UserID: user.ID, OrgID: orgID, RefreshTokenHash: pair.RefreshTokenHash, UserAgent: userAgent, IPAddress: ip, ExpiresAt: pair.RefreshExpiresAt, CreatedAt: now, UpdatedAt: now}
	if err := repo.CreateSession(ctx, session); err != nil {
		return TokenPair{}, err
	}
	return tokenPair(pair), nil
}

func (s *service) loginOrg(ctx context.Context, userID int64, code string) (*model.Organization, error) {
	if code != "" {
		org, err := s.repo.FindOrganizationByCode(ctx, normalizeCode(code))
		if err != nil {
			return nil, ErrNotFound
		}
		if _, err := s.repo.FindMembership(ctx, org.ID, userID); err != nil {
			return nil, ErrForbidden
		}
		return org, nil
	}
	memberships, err := s.repo.ListMembershipsByUser(ctx, userID)
	if err != nil || len(memberships) == 0 {
		return nil, ErrForbidden
	}
	return s.repo.FindOrganizationByID(ctx, memberships[0].OrgID)
}

func (s *service) ensureUserCanLogin(user *model.User) error {
	if user.Status != model.StatusActive {
		return ErrAccountDisabled
	}
	if user.LockedUntil != nil && user.LockedUntil.After(s.now()) {
		return ErrAccountLocked
	}
	return nil
}

func (s *service) ensureSessionActive(session *model.Session) error {
	if session.RevokedAt != nil {
		return ErrSessionRevoked
	}
	if session.ExpiresAt.Before(s.now()) {
		return ErrInvalidToken
	}
	return nil
}

func (s *service) recordFailedLogin(ctx context.Context, user *model.User) error {
	user.FailedLoginAttempts++
	if user.FailedLoginAttempts >= s.cfg.LoginMaxFailures {
		lockedUntil := s.now().Add(s.cfg.LoginLockDuration)
		user.LockedUntil = &lockedUntil
	}
	return s.repo.SaveUser(ctx, user)
}

func (s *service) verifyUserMFA(ctx context.Context, userID int64, code string) error {
	factor, err := s.repo.FindActiveMFAFactor(ctx, userID)
	if err != nil {
		return ErrUnauthorized
	}
	secret, err := s.decryptSecret(factor.Secret)
	if err != nil {
		return err
	}
	if !mfa.ValidateTOTP(code, secret) {
		return ErrUnauthorized
	}
	return nil
}

func (s *service) ensureMembership(ctx context.Context, repo repository.Repository, orgID, userID int64) error {
	if _, err := repo.FindMembership(ctx, orgID, userID); err == nil {
		return nil
	} else if !errors.Is(err, database.ErrNotFound) {
		return err
	}
	now := s.now()
	return repo.CreateMembership(ctx, &model.Membership{ID: s.ids.NextID(), OrgID: orgID, UserID: userID, Status: model.StatusActive, CreatedAt: now, UpdatedAt: now})
}

func (s *service) ensureBuiltins(ctx context.Context, repo repository.Repository, orgID int64) error {
	for _, permission := range builtinPermissions {
		if _, err := repo.FindPermission(ctx, permission.Code); err == nil {
			continue
		} else if !errors.Is(err, database.ErrNotFound) {
			return err
		}
		now := s.now()
		if err := repo.CreatePermission(ctx, &model.Permission{ID: s.ids.NextID(), Code: permission.Code, Name: permission.Name, Description: permission.Description, CreatedAt: now, UpdatedAt: now}); err != nil {
			return err
		}
	}
	for _, role := range []struct {
		code string
		name string
	}{
		{model.RoleOwner, "Owner"},
		{model.RoleAdmin, "Admin"},
		{model.RoleMember, "Member"},
	} {
		if _, err := repo.FindRole(ctx, orgID, role.code); err == nil {
			continue
		} else if !errors.Is(err, database.ErrNotFound) {
			return err
		}
		now := s.now()
		if err := repo.CreateRole(ctx, &model.Role{ID: s.ids.NextID(), OrgID: orgID, Code: role.code, Name: role.name, Description: role.name, System: true, CreatedAt: now, UpdatedAt: now}); err != nil {
			return err
		}
	}
	if err := s.addPolicy(ctx, repo, orgID, model.RoleOwner, "*", "*"); err != nil {
		return err
	}
	for _, permission := range builtinPermissions {
		obj, act := permissionObjectAction(permission.Code)
		if obj != "" && act != "" {
			if err := s.addPolicy(ctx, repo, orgID, model.RoleAdmin, obj, act); err != nil {
				return err
			}
		}
	}
	return s.addPolicy(ctx, repo, orgID, model.RoleMember, "me", "read")
}

func (s *service) addUserRole(ctx context.Context, repo repository.Repository, userID, orgID int64, roleCode string) error {
	now := s.now()
	rule := &model.CasbinRule{ID: s.ids.NextID(), PType: "g", V0: userSubject(userID), V1: roleSubject(roleCode), V2: strconv.FormatInt(orgID, 10), CreatedAt: now}
	return repo.AddCasbinRule(ctx, rule)
}

func (s *service) addPolicy(ctx context.Context, repo repository.Repository, orgID int64, roleCode, obj, act string) error {
	now := s.now()
	rule := &model.CasbinRule{ID: s.ids.NextID(), PType: "p", V0: roleSubject(roleCode), V1: strconv.FormatInt(orgID, 10), V2: obj, V3: act, CreatedAt: now}
	return repo.AddCasbinRule(ctx, rule)
}

func (s *service) audit(ctx context.Context, repo repository.Repository, orgID, userID *int64, action, resource, resourceID, ip, userAgent string, metadata map[string]any) error {
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, _ := json.Marshal(metadata)
	return repo.CreateAuditLog(ctx, &model.AuditLog{ID: s.ids.NextID(), OrgID: orgID, UserID: userID, Action: action, Resource: resource, ResourceID: resourceID, IPAddress: ip, UserAgent: userAgent, Metadata: string(raw), CreatedAt: s.now()})
}

func (s *service) oneTimeToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return "", "", err
	}
	value := base64.RawURLEncoding.EncodeToString(raw)
	return value, s.tokens.HashRefreshToken(value), nil
}

func (s *service) debugDelivery(token string, adminPath string) NotificationDelivery {
	if !s.debugNotificationsEnabled() {
		return NotificationDelivery{}
	}
	return NotificationDelivery{Token: token, URL: s.notificationURL(adminPath)}
}

func (s *service) notificationURL(adminPath string) string {
	adminPath = strings.TrimLeft(adminPath, "/")
	base := s.publicBaseURL()
	if base == "" {
		return "/admin/" + adminPath
	}
	return strings.TrimRight(base, "/") + "/" + adminPath
}

func (s *service) debugNotificationsEnabled() bool {
	switch normalizeCode(s.cfg.NotificationDriver) {
	case "", "debug", "noop", "local":
		return true
	default:
		return false
	}
}

func (s *service) publicBaseURL() string {
	return strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/")
}

func (s *service) validatePassword(value string) error {
	policy := s.cfg.PasswordPolicy
	if policy.MinLength <= 0 {
		policy.MinLength = 8
	}
	if len([]rune(value)) < policy.MinLength {
		return ErrInvalidInput
	}
	var hasLower, hasUpper, hasNumber, hasSymbol bool
	for _, r := range value {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}
	if policy.RequireLower && !hasLower {
		return ErrInvalidInput
	}
	if policy.RequireUpper && !hasUpper {
		return ErrInvalidInput
	}
	if policy.RequireNumber && !hasNumber {
		return ErrInvalidInput
	}
	if policy.RequireSymbol && !hasSymbol {
		return ErrInvalidInput
	}
	return nil
}

func (s *service) encryptSecret(secret string) (string, error) {
	block, err := aes.NewCipher(secretKey(s.cfg.MFASecretKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(secret), nil)
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

func (s *service) decryptSecret(value string) (string, error) {
	raw, err := base64.RawStdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(secretKey(s.cfg.MFASecretKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", ErrInvalidToken
	}
	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

func (s *service) now() time.Time {
	return s.cfg.Now().UTC()
}

func tokenPair(pair token.Pair) TokenPair {
	return TokenPair{AccessToken: pair.AccessToken, AccessExpiresAt: pair.AccessExpiresAt, RefreshToken: pair.RefreshToken, RefreshExpiresAt: pair.RefreshExpiresAt}
}

func normalizeCode(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func userSubject(id int64) string {
	return "user:" + strconv.FormatInt(id, 10)
}

func roleSubject(code string) string {
	if strings.HasPrefix(code, "role:") {
		return code
	}
	return "role:" + normalizeCode(code)
}

func permissionObjectAction(code string) (string, string) {
	parts := strings.SplitN(strings.TrimSpace(code), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", ""
	}
	return parts[0], parts[1]
}

func (s *service) hydrateRolePermissions(ctx context.Context, role *model.Role) error {
	if role == nil {
		return nil
	}
	permissions, err := s.repo.ListRolePermissions(ctx, role.OrgID, roleSubject(role.Code))
	if err != nil {
		return err
	}
	role.Permissions = permissions
	return nil
}

func secretKey(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}

type permissionSeed struct {
	Code        string
	Name        string
	Description string
}

var builtinPermissions = []permissionSeed{
	{Code: "org:create", Name: "Create organizations", Description: "Create organizations"},
	{Code: "org:read", Name: "Read organizations", Description: "Read organizations"},
	{Code: "org:update", Name: "Update organizations", Description: "Update organization settings"},
	{Code: "user:read", Name: "Read users", Description: "Read organization users"},
	{Code: "user:invite", Name: "Invite users", Description: "Invite users into organization"},
	{Code: "user:update", Name: "Update users", Description: "Update organization users"},
	{Code: "user:disable", Name: "Disable users", Description: "Disable organization users"},
	{Code: "role:read", Name: "Read roles", Description: "Read roles"},
	{Code: "role:create", Name: "Create roles", Description: "Create roles"},
	{Code: "role:update", Name: "Update roles", Description: "Update roles"},
	{Code: "permission:read", Name: "Read permissions", Description: "Read permissions"},
	{Code: "session:read", Name: "Read sessions", Description: "Read sessions"},
	{Code: "session:revoke", Name: "Revoke sessions", Description: "Revoke sessions"},
	{Code: "audit:read", Name: "Read audit logs", Description: "Read audit logs"},
	{Code: "plugin:read", Name: "Read plugins", Description: "Read installed plugin manifests"},
	{Code: "plugin:proxy", Name: "Proxy plugins", Description: "Call installed plugin sidecar APIs"},
}
