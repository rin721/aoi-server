package service

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rei0721/go-scaffold/internal/modules/iam/model"
	"github.com/rei0721/go-scaffold/internal/modules/iam/repository"
	"github.com/rei0721/go-scaffold/pkg/authorization"
	"github.com/rei0721/go-scaffold/pkg/crypto"
	"github.com/rei0721/go-scaffold/pkg/database"
	"github.com/rei0721/go-scaffold/pkg/mfa"
	"github.com/rei0721/go-scaffold/pkg/migrator"
	"github.com/rei0721/go-scaffold/pkg/token"
	"github.com/rei0721/go-scaffold/pkg/utils"
)

func TestIAMLifecycle(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestService(t)
	defer cleanup()

	admin, err := svc.BootstrapAdmin(ctx, BootstrapAdminInput{
		OrgCode:  "acme",
		OrgName:  "Acme",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("BootstrapAdmin() failed: %v", err)
	}
	if admin.OrgID == 0 || admin.UserID == 0 {
		t.Fatalf("unexpected principal: %#v", admin)
	}

	allowed, err := svc.Authorize(ctx, *admin, "audit", "read")
	if err != nil || !allowed {
		t.Fatalf("owner should read audit logs, allowed=%v err=%v", allowed, err)
	}

	login, err := svc.Login(ctx, LoginInput{Identifier: "admin@example.com", Password: "password123", OrgCode: "acme"})
	if err != nil {
		t.Fatalf("Login() failed: %v", err)
	}
	principal, err := svc.AuthenticateToken(ctx, login.AccessToken)
	if err != nil {
		t.Fatalf("AuthenticateToken() failed: %v", err)
	}
	if principal.UserID != admin.UserID || principal.OrgID != admin.OrgID {
		t.Fatalf("unexpected authenticated principal: %#v", principal)
	}

	refreshed, err := svc.Refresh(ctx, RefreshInput{RefreshToken: login.RefreshToken})
	if err != nil {
		t.Fatalf("Refresh() failed: %v", err)
	}
	if refreshed.AccessToken == "" || refreshed.RefreshToken == "" || refreshed.RefreshToken == login.RefreshToken {
		t.Fatalf("refresh rotation failed: %#v", refreshed)
	}

	inviteDelivery, err := svc.InviteUser(ctx, InviteUserInput{Principal: principal, Email: "member@example.com", RoleCode: model.RoleMember})
	if err != nil {
		t.Fatalf("InviteUser() failed: %v", err)
	}
	if inviteDelivery.Token == "" || inviteDelivery.URL == "" {
		t.Fatalf("expected debug invitation delivery, got %#v", inviteDelivery)
	}
	member, err := svc.AcceptInvitation(ctx, AcceptInvitationInput{Token: inviteDelivery.Token, Username: "member", Password: "password123"})
	if err != nil {
		t.Fatalf("AcceptInvitation() failed: %v", err)
	}
	memberAllowed, err := svc.Authorize(ctx, *member, "audit", "read")
	if err != nil {
		t.Fatalf("Authorize(member) failed: %v", err)
	}
	if memberAllowed {
		t.Fatal("member should not read audit logs")
	}
	memberLogin, err := svc.Login(ctx, LoginInput{Identifier: "member@example.com", Password: "password123", OrgCode: "acme"})
	if err != nil {
		t.Fatalf("member Login() before reset failed: %v", err)
	}

	resetDelivery, err := svc.ForgotPassword(ctx, ForgotPasswordInput{Email: "member@example.com"})
	if err != nil {
		t.Fatalf("ForgotPassword() failed: %v", err)
	}
	if resetDelivery.Token == "" || resetDelivery.URL == "" {
		t.Fatalf("expected debug password reset delivery, got %#v", resetDelivery)
	}
	if err := svc.ResetPassword(ctx, ResetPasswordInput{Token: resetDelivery.Token, NewPassword: "newpassword123"}); err != nil {
		t.Fatalf("ResetPassword() failed: %v", err)
	}
	if _, err := svc.Refresh(ctx, RefreshInput{RefreshToken: memberLogin.RefreshToken}); err != ErrSessionRevoked {
		t.Fatalf("old refresh after password reset = %v, want ErrSessionRevoked", err)
	}
	if _, err := svc.Login(ctx, LoginInput{Identifier: "member@example.com", Password: "newpassword123", OrgCode: "acme"}); err != nil {
		t.Fatalf("member login after reset failed: %v", err)
	}
}

func TestSelfSignupCreatesOwnerSession(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestService(t)
	defer cleanup()

	pair, err := svc.Signup(ctx, SignupInput{
		OrgCode:  "acme",
		OrgName:  "Acme",
		Username: "owner",
		Email:    "owner@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Signup() failed: %v", err)
	}
	principal, err := svc.AuthenticateToken(ctx, pair.AccessToken)
	if err != nil {
		t.Fatalf("AuthenticateToken(signup token) failed: %v", err)
	}
	allowed, err := svc.Authorize(ctx, principal, "audit", "read")
	if err != nil || !allowed {
		t.Fatalf("signup owner should read audit logs, allowed=%v err=%v", allowed, err)
	}
	orgs, err := svc.ListMyOrganizations(ctx, principal)
	if err != nil || len(orgs) != 1 || orgs[0].Code != "acme" {
		t.Fatalf("unexpected signup organizations: %#v err=%v", orgs, err)
	}
	if _, err := svc.CreateRole(ctx, CreateRoleInput{
		Principal:   principal,
		Code:        "operator",
		Name:        "Operator",
		Permissions: []string{"audit:read", "user:read"},
	}); err != nil {
		t.Fatalf("CreateRole() failed: %v", err)
	}
	roles, err := svc.ListRoles(ctx, principal)
	if err != nil {
		t.Fatalf("ListRoles() failed: %v", err)
	}
	var operator *model.Role
	for i := range roles {
		if roles[i].Code == "operator" {
			operator = &roles[i]
			break
		}
	}
	if operator == nil || !containsString(operator.Permissions, "audit:read") || !containsString(operator.Permissions, "user:read") {
		t.Fatalf("operator permissions not hydrated: %#v", operator)
	}

	if _, err := svc.Signup(ctx, SignupInput{OrgCode: "acme", OrgName: "Other", Username: "other", Email: "other@example.com", Password: "password123"}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate org signup error = %v, want ErrDuplicate", err)
	}
	if _, err := svc.Signup(ctx, SignupInput{OrgCode: "other", OrgName: "Other", Username: "owner", Email: "other@example.com", Password: "password123"}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate username signup error = %v, want ErrDuplicate", err)
	}
	if _, err := svc.Signup(ctx, SignupInput{OrgCode: "other", OrgName: "Other", Username: "other", Email: "owner@example.com", Password: "password123"}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("duplicate email signup error = %v, want ErrDuplicate", err)
	}
}

func TestListOrganizationsFiltersAndPaginates(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestService(t)
	defer cleanup()

	admin, err := svc.BootstrapAdmin(ctx, BootstrapAdminInput{
		OrgCode:  "core",
		OrgName:  "Core Org",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("BootstrapAdmin() failed: %v", err)
	}
	if _, err := svc.CreateOrganization(ctx, *admin, "alpha", "Alpha Team"); err != nil {
		t.Fatalf("CreateOrganization(alpha) failed: %v", err)
	}
	if _, err := svc.CreateOrganization(ctx, *admin, "beta", "Beta Team"); err != nil {
		t.Fatalf("CreateOrganization(beta) failed: %v", err)
	}
	if _, err := svc.CreateOrganization(ctx, *admin, "support", "Support Desk"); err != nil {
		t.Fatalf("CreateOrganization(support) failed: %v", err)
	}

	firstPage, err := svc.ListOrganizations(ctx, *admin, OrganizationListFilter{
		Keyword:  "team",
		OrderKey: "code",
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("ListOrganizations(page 1) failed: %v", err)
	}
	if firstPage.Total != 2 || firstPage.Page != 1 || firstPage.PageSize != 1 || len(firstPage.Items) != 1 || firstPage.Items[0].Code != "alpha" || firstPage.StorageStatus != "persisted" {
		t.Fatalf("unexpected first page: %#v", firstPage)
	}

	secondPage, err := svc.ListOrganizations(ctx, *admin, OrganizationListFilter{
		Keyword:  "team",
		OrderKey: "code",
		Page:     2,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("ListOrganizations(page 2) failed: %v", err)
	}
	if len(secondPage.Items) != 1 || secondPage.Items[0].Code != "beta" {
		t.Fatalf("unexpected second page: %#v", secondPage)
	}

	filtered, err := svc.ListOrganizations(ctx, *admin, OrganizationListFilter{
		Code:     "sup",
		Name:     "desk",
		Status:   model.StatusActive,
		OrderKey: "name",
		Desc:     true,
	})
	if err != nil {
		t.Fatalf("ListOrganizations(filtered) failed: %v", err)
	}
	if filtered.Total != 1 || len(filtered.Items) != 1 || filtered.Items[0].Code != "support" {
		t.Fatalf("unexpected filtered organizations: %#v", filtered)
	}
}

func TestInitialAdminSetupCreatesFirstOwnerAndClosesSetup(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestServiceWithSignup(t, false)
	defer cleanup()

	status, err := svc.SetupStatus(ctx)
	if err != nil {
		t.Fatalf("SetupStatus() failed: %v", err)
	}
	if !status.Required {
		t.Fatal("SetupStatus().Required = false, want true for empty IAM users")
	}

	pair, err := svc.InitialAdminSetup(ctx, InitialAdminSetupInput{
		OrgCode:  "acme",
		OrgName:  "Acme",
		Username: "admin",
		Email:    "admin@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("InitialAdminSetup() failed: %v", err)
	}
	principal, err := svc.AuthenticateToken(ctx, pair.AccessToken)
	if err != nil {
		t.Fatalf("AuthenticateToken(initial setup token) failed: %v", err)
	}
	for _, tc := range []struct {
		obj string
		act string
	}{
		{obj: "audit", act: "read"},
		{obj: "session", act: "read"},
		{obj: "org", act: "read"},
		{obj: "user", act: "read"},
	} {
		allowed, err := svc.Authorize(ctx, principal, tc.obj, tc.act)
		if err != nil || !allowed {
			t.Fatalf("initial owner should %s:%s, allowed=%v err=%v", tc.obj, tc.act, allowed, err)
		}
	}
	orgs, err := svc.ListMyOrganizations(ctx, principal)
	if err != nil || len(orgs) != 1 || orgs[0].Code != "acme" {
		t.Fatalf("unexpected setup organizations: %#v err=%v", orgs, err)
	}
	status, err = svc.SetupStatus(ctx)
	if err != nil {
		t.Fatalf("SetupStatus(after setup) failed: %v", err)
	}
	if status.Required {
		t.Fatal("SetupStatus().Required = true after initial setup, want false")
	}
	if _, err := svc.InitialAdminSetup(ctx, InitialAdminSetupInput{OrgCode: "other", OrgName: "Other", Username: "other", Email: "other@example.com", Password: "password123"}); !errors.Is(err, ErrSetupCompleted) {
		t.Fatalf("second InitialAdminSetup() error = %v, want ErrSetupCompleted", err)
	}
}

func TestSelfSignupDisabled(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestServiceWithSignup(t, false)
	defer cleanup()

	_, err := svc.Signup(ctx, SignupInput{OrgCode: "acme", OrgName: "Acme", Username: "owner", Email: "owner@example.com", Password: "password123"})
	if err != ErrSignupDisabled {
		t.Fatalf("Signup() error = %v, want ErrSignupDisabled", err)
	}
}

func TestCreateOrganizationAddsCurrentUserAsOwner(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestService(t)
	defer cleanup()

	admin, err := svc.BootstrapAdmin(ctx, BootstrapAdminInput{OrgCode: "acme", Username: "admin", Email: "admin@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("BootstrapAdmin() failed: %v", err)
	}
	login, err := svc.Login(ctx, LoginInput{Identifier: "admin@example.com", Password: "password123", OrgCode: "acme"})
	if err != nil {
		t.Fatalf("Login() failed: %v", err)
	}
	principal, err := svc.AuthenticateToken(ctx, login.AccessToken)
	if err != nil {
		t.Fatalf("AuthenticateToken() failed: %v", err)
	}
	org, err := svc.CreateOrganization(ctx, principal, "beta", "Beta")
	if err != nil {
		t.Fatalf("CreateOrganization() failed: %v", err)
	}
	switched, err := svc.SwitchOrg(ctx, principal, org.ID, "", "")
	if err != nil {
		t.Fatalf("SwitchOrg(created org) failed: %v", err)
	}
	newPrincipal, err := svc.AuthenticateToken(ctx, switched.AccessToken)
	if err != nil {
		t.Fatalf("AuthenticateToken(new org) failed: %v", err)
	}
	if newPrincipal.UserID != admin.UserID || newPrincipal.OrgID != org.ID {
		t.Fatalf("unexpected switched principal: %#v", newPrincipal)
	}
	allowed, err := svc.Authorize(ctx, newPrincipal, "role", "create")
	if err != nil || !allowed {
		t.Fatalf("created org owner should create roles, allowed=%v err=%v", allowed, err)
	}
}

func TestListUsersFiltersAndPaginates(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestService(t)
	defer cleanup()

	admin, err := svc.BootstrapAdmin(ctx, BootstrapAdminInput{OrgCode: "acme", Username: "admin", Email: "admin@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("BootstrapAdmin() failed: %v", err)
	}
	for _, input := range []struct {
		email    string
		username string
		roleCode string
	}{
		{email: "alice@example.com", username: "alice", roleCode: model.RoleMember},
		{email: "bob@example.com", username: "bob", roleCode: model.RoleAdmin},
	} {
		invite, err := svc.InviteUser(ctx, InviteUserInput{Principal: *admin, Email: input.email, RoleCode: input.roleCode})
		if err != nil {
			t.Fatalf("InviteUser(%s) failed: %v", input.email, err)
		}
		if _, err := svc.AcceptInvitation(ctx, AcceptInvitationInput{Token: invite.Token, Username: input.username, Password: "password123"}); err != nil {
			t.Fatalf("AcceptInvitation(%s) failed: %v", input.email, err)
		}
	}

	all, err := svc.ListUsers(ctx, *admin, UserListFilter{Page: 1, PageSize: 2, Desc: true})
	if err != nil {
		t.Fatalf("ListUsers() failed: %v", err)
	}
	if all.Total != 3 || len(all.Items) != 2 || all.Page != 1 || all.PageSize != 2 {
		t.Fatalf("unexpected first page: %#v", all)
	}

	memberPage, err := svc.ListUsers(ctx, *admin, UserListFilter{RoleCode: model.RoleMember})
	if err != nil {
		t.Fatalf("ListUsers(member) failed: %v", err)
	}
	if memberPage.Total != 1 || memberPage.Items[0].User.Username != "alice" {
		t.Fatalf("unexpected member filter page: %#v", memberPage)
	}

	keywordPage, err := svc.ListUsers(ctx, *admin, UserListFilter{Keyword: "bob"})
	if err != nil {
		t.Fatalf("ListUsers(keyword) failed: %v", err)
	}
	if keywordPage.Total != 1 || keywordPage.Items[0].User.Email != "bob@example.com" {
		t.Fatalf("unexpected keyword filter page: %#v", keywordPage)
	}

	if _, err := svc.UpdateUser(ctx, UpdateUserInput{Principal: *admin, UserID: keywordPage.Items[0].User.ID, Status: ptrString(model.StatusDisabled)}); err != nil {
		t.Fatalf("UpdateUser(disable bob) failed: %v", err)
	}
	disabledPage, err := svc.ListUsers(ctx, *admin, UserListFilter{Status: model.StatusDisabled})
	if err != nil {
		t.Fatalf("ListUsers(disabled) failed: %v", err)
	}
	if disabledPage.Total != 1 || disabledPage.Items[0].User.Username != "bob" {
		t.Fatalf("unexpected disabled filter page: %#v", disabledPage)
	}
}

func TestListSessionsFiltersPaginatesAndScopesOrganization(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestService(t)
	defer cleanup()

	admin, err := svc.BootstrapAdmin(ctx, BootstrapAdminInput{OrgCode: "acme", Username: "admin", Email: "admin@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("BootstrapAdmin() failed: %v", err)
	}
	adminLogin, err := svc.Login(ctx, LoginInput{Identifier: "admin@example.com", Password: "password123", OrgCode: "acme", IPAddress: "127.0.0.1", UserAgent: "Edge"})
	if err != nil {
		t.Fatalf("Login(admin) failed: %v", err)
	}
	principal, err := svc.AuthenticateToken(ctx, adminLogin.AccessToken)
	if err != nil {
		t.Fatalf("AuthenticateToken(admin) failed: %v", err)
	}
	invite, err := svc.InviteUser(ctx, InviteUserInput{Principal: principal, Email: "member@example.com", RoleCode: model.RoleMember})
	if err != nil {
		t.Fatalf("InviteUser() failed: %v", err)
	}
	if _, err := svc.AcceptInvitation(ctx, AcceptInvitationInput{Token: invite.Token, Username: "member", Password: "password123"}); err != nil {
		t.Fatalf("AcceptInvitation() failed: %v", err)
	}
	memberLogin, err := svc.Login(ctx, LoginInput{Identifier: "member@example.com", Password: "password123", OrgCode: "acme", IPAddress: "10.0.0.2", UserAgent: "Firefox"})
	if err != nil {
		t.Fatalf("Login(member) failed: %v", err)
	}
	memberPrincipal, err := svc.AuthenticateToken(ctx, memberLogin.AccessToken)
	if err != nil {
		t.Fatalf("AuthenticateToken(member) failed: %v", err)
	}
	beta, err := svc.CreateOrganization(ctx, principal, "beta", "Beta")
	if err != nil {
		t.Fatalf("CreateOrganization(beta) failed: %v", err)
	}
	if _, err := svc.SwitchOrg(ctx, principal, beta.ID, "Safari", "172.16.0.1"); err != nil {
		t.Fatalf("SwitchOrg(beta) failed: %v", err)
	}

	ownPage, err := svc.ListSessions(ctx, principal, SessionListFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListSessions(own) failed: %v", err)
	}
	if ownPage.Total != 1 || len(ownPage.Items) != 1 || ownPage.Items[0].UserID != admin.UserID || ownPage.Items[0].OrgID != principal.OrgID {
		t.Fatalf("unexpected own sessions: %#v", ownPage)
	}

	orgPage, err := svc.ListSessions(ctx, principal, SessionListFilter{Scope: "org", Keyword: "fire", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListSessions(org keyword) failed: %v", err)
	}
	if orgPage.Total != 1 || len(orgPage.Items) != 1 || orgPage.Items[0].UserID != memberPrincipal.UserID || orgPage.Items[0].OrgID != principal.OrgID {
		t.Fatalf("unexpected org keyword sessions: %#v", orgPage)
	}

	adminOrgSessions, err := svc.ListSessions(ctx, principal, SessionListFilter{UserID: admin.UserID, Scope: "org", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListSessions(admin user) failed: %v", err)
	}
	if adminOrgSessions.Total != 1 || len(adminOrgSessions.Items) != 1 || adminOrgSessions.Items[0].OrgID != principal.OrgID {
		t.Fatalf("expected beta session to be filtered out: %#v", adminOrgSessions)
	}

	if err := svc.RevokeSession(ctx, principal, adminOrgSessions.Items[0].ID); err != nil {
		t.Fatalf("RevokeSession() failed: %v", err)
	}
	revokedPage, err := svc.ListSessions(ctx, principal, SessionListFilter{Scope: "org", Status: "revoked"})
	if err != nil {
		t.Fatalf("ListSessions(revoked) failed: %v", err)
	}
	if revokedPage.Total != 1 || len(revokedPage.Items) != 1 || revokedPage.Items[0].ID != adminOrgSessions.Items[0].ID {
		t.Fatalf("unexpected revoked sessions: %#v", revokedPage)
	}
}

func TestMFASetupAndLogin(t *testing.T) {
	ctx := context.Background()
	svc, cleanup := newTestService(t)
	defer cleanup()
	admin, err := svc.BootstrapAdmin(ctx, BootstrapAdminInput{OrgCode: "acme", Username: "admin", Email: "admin@example.com", Password: "password123"})
	if err != nil {
		t.Fatalf("BootstrapAdmin() failed: %v", err)
	}
	secret, _, err := svc.SetupMFA(ctx, *admin)
	if err != nil {
		t.Fatalf("SetupMFA() failed: %v", err)
	}
	oldCode, err := mfa.GenerateTOTPCode(secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateTOTPCode() for old secret failed: %v", err)
	}
	secret, _, err = svc.SetupMFA(ctx, *admin)
	if err != nil {
		t.Fatalf("second SetupMFA() failed: %v", err)
	}
	if err := svc.VerifyMFA(ctx, *admin, oldCode); err == nil {
		t.Fatal("VerifyMFA() should reject code from replaced setup secret")
	}
	code, err := mfa.GenerateTOTPCode(secret, time.Now())
	if err != nil {
		t.Fatalf("GenerateCode() failed: %v", err)
	}
	if err := svc.VerifyMFA(ctx, *admin, code); err != nil {
		t.Fatalf("VerifyMFA() failed: %v", err)
	}
	if _, err := svc.Login(ctx, LoginInput{Identifier: "admin@example.com", Password: "password123", OrgCode: "acme"}); err != ErrMFARequired {
		t.Fatalf("expected ErrMFARequired, got %v", err)
	}
	code, _ = mfa.GenerateTOTPCode(secret, time.Now())
	if _, err := svc.Login(ctx, LoginInput{Identifier: "admin@example.com", Password: "password123", OrgCode: "acme", MFACode: code}); err != nil {
		t.Fatalf("MFA login failed: %v", err)
	}
}

func TestLoginCaptchaWhenEnabled(t *testing.T) {
	svc, cleanup := newTestService(t)
	defer cleanup()
	impl := svc.(*service)
	impl.cfg.CaptchaEnabled = true
	impl.cfg.CaptchaTTL = time.Minute
	ctx := context.Background()

	if _, err := svc.BootstrapAdmin(ctx, BootstrapAdminInput{OrgCode: "acme", Username: "admin", Email: "admin@example.com", Password: "password123"}); err != nil {
		t.Fatalf("BootstrapAdmin() failed: %v", err)
	}

	challenge, err := svc.Captcha(ctx)
	if err != nil {
		t.Fatalf("Captcha() error = %v", err)
	}
	if !challenge.Enabled || challenge.CaptchaID == "" || !strings.HasPrefix(challenge.Image, "data:image/svg+xml;base64,") {
		t.Fatalf("unexpected captcha challenge: %#v", challenge)
	}
	if _, err := svc.Login(ctx, LoginInput{Identifier: "admin@example.com", Password: "password123", OrgCode: "acme"}); !errors.Is(err, ErrCaptchaRequired) {
		t.Fatalf("expected captcha required error, got %v", err)
	}
	if _, err := svc.Login(ctx, LoginInput{CaptchaID: challenge.CaptchaID, CaptchaCode: "bad", Identifier: "admin@example.com", Password: "password123", OrgCode: "acme"}); !errors.Is(err, ErrCaptchaInvalid) {
		t.Fatalf("expected captcha invalid error, got %v", err)
	}

	challenge, err = svc.Captcha(ctx)
	if err != nil {
		t.Fatalf("Captcha() second error = %v", err)
	}
	impl.captchaMu.Lock()
	answer := impl.captchaChallenges[challenge.CaptchaID].answer
	impl.captchaMu.Unlock()
	if answer == "" {
		t.Fatal("expected captcha answer to be stored")
	}
	if _, err := svc.Login(ctx, LoginInput{CaptchaID: challenge.CaptchaID, CaptchaCode: answer, Identifier: "admin@example.com", Password: "password123", OrgCode: "acme"}); err != nil {
		t.Fatalf("Login() with captcha failed: %v", err)
	}
}

func newTestService(t *testing.T) (Service, func()) {
	return newTestServiceWithSignup(t, true)
}

func newTestServiceWithSignup(t *testing.T, selfSignupEnabled bool) (Service, func()) {
	t.Helper()
	db, err := database.New(&database.Config{Driver: database.DriverSQLite, DBName: filepath.Join(t.TempDir(), "iam.db")})
	if err != nil {
		t.Fatalf("database.New() failed: %v", err)
	}
	runner, err := migrator.New(db, migrator.Config{Driver: string(database.DriverSQLite), Dir: filepath.Join("..", "..", "..", "migrations")})
	if err != nil {
		t.Fatalf("migrator.New() failed: %v", err)
	}
	if err := runner.Up(context.Background()); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	passwords, err := crypto.NewBcrypt()
	if err != nil {
		t.Fatalf("crypto.NewBcrypt() failed: %v", err)
	}
	tokens, err := token.New(token.Config{
		Issuer:        "test",
		Audience:      []string{"test"},
		SigningKey:    "01234567890123456789012345678901",
		AccessTTL:     time.Hour,
		RefreshTTL:    time.Hour,
		RefreshPepper: "refresh-pepper",
	})
	if err != nil {
		t.Fatalf("token.New() failed: %v", err)
	}
	authz, err := authorization.New()
	if err != nil {
		t.Fatalf("authorization.New() failed: %v", err)
	}
	ids, err := utils.NewSnowflake(7)
	if err != nil {
		t.Fatalf("utils.NewSnowflake() failed: %v", err)
	}
	repo := repository.New(db)
	svc := New(db, repo, passwords, tokens, authz, ids, Config{
		SelfSignupEnabled:  selfSignupEnabled,
		MFAIssuer:          "go-scaffold-test",
		MFASecretKey:       "01234567890123456789012345678901",
		LoginMaxFailures:   3,
		LoginLockDuration:  time.Minute,
		InvitationTTL:      time.Hour,
		PasswordResetTTL:   time.Hour,
		NotificationDriver: "debug",
		PublicBaseURL:      "/admin",
	}, NoopNotifier{})
	return svc, func() { _ = db.Close() }
}

func ptrString(value string) *string {
	return &value
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
