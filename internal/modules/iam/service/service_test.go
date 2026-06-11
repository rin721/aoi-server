package service

import (
	"context"
	"path/filepath"
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

	inviteToken, err := svc.InviteUser(ctx, InviteUserInput{Principal: principal, Email: "member@example.com", RoleCode: model.RoleMember})
	if err != nil {
		t.Fatalf("InviteUser() failed: %v", err)
	}
	member, err := svc.AcceptInvitation(ctx, AcceptInvitationInput{Token: inviteToken, Username: "member", Password: "password123"})
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

	resetToken, err := svc.ForgotPassword(ctx, ForgotPasswordInput{Email: "member@example.com"})
	if err != nil {
		t.Fatalf("ForgotPassword() failed: %v", err)
	}
	if err := svc.ResetPassword(ctx, ResetPasswordInput{Token: resetToken, NewPassword: "newpassword123"}); err != nil {
		t.Fatalf("ResetPassword() failed: %v", err)
	}
	if _, err := svc.Refresh(ctx, RefreshInput{RefreshToken: memberLogin.RefreshToken}); err != ErrSessionRevoked {
		t.Fatalf("old refresh after password reset = %v, want ErrSessionRevoked", err)
	}
	if _, err := svc.Login(ctx, LoginInput{Identifier: "member@example.com", Password: "newpassword123", OrgCode: "acme"}); err != nil {
		t.Fatalf("member login after reset failed: %v", err)
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

func newTestService(t *testing.T) (Service, func()) {
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
		MFAIssuer:         "go-scaffold-test",
		MFASecretKey:      "01234567890123456789012345678901",
		LoginMaxFailures:  3,
		LoginLockDuration: time.Minute,
		InvitationTTL:     time.Hour,
		PasswordResetTTL:  time.Hour,
	}, NoopNotifier{})
	return svc, func() { _ = db.Close() }
}
