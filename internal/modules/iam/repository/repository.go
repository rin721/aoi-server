package repository

import (
	"context"
	"errors"
	"time"

	"github.com/rei0721/go-scaffold/internal/modules/iam/model"
	"github.com/rei0721/go-scaffold/pkg/authorization"
	"github.com/rei0721/go-scaffold/pkg/database"
)

type Repository interface {
	WithExecutor(database.Executor) Repository
	CreateOrganization(context.Context, *model.Organization) error
	FindOrganizationByID(context.Context, int64) (*model.Organization, error)
	FindOrganizationByCode(context.Context, string) (*model.Organization, error)
	ListOrganizations(context.Context) ([]model.Organization, error)
	CreateUser(context.Context, *model.User) error
	FindUserByID(context.Context, int64) (*model.User, error)
	FindUserByIdentifier(context.Context, string) (*model.User, error)
	SaveUser(context.Context, *model.User) error
	CreateMembership(context.Context, *model.Membership) error
	FindMembership(context.Context, int64, int64) (*model.Membership, error)
	ListMembershipsByUser(context.Context, int64) ([]model.Membership, error)
	ListUsersByOrg(context.Context, int64) ([]model.User, error)
	CreateRole(context.Context, *model.Role) error
	FindRole(context.Context, int64, string) (*model.Role, error)
	ListRoles(context.Context, int64) ([]model.Role, error)
	CreatePermission(context.Context, *model.Permission) error
	FindPermission(context.Context, string) (*model.Permission, error)
	ListPermissions(context.Context) ([]model.Permission, error)
	CreateSession(context.Context, *model.Session) error
	FindSessionByID(context.Context, int64) (*model.Session, error)
	FindSessionByRefreshHash(context.Context, string) (*model.Session, error)
	ListSessionsByUser(context.Context, int64) ([]model.Session, error)
	SaveSession(context.Context, *model.Session) error
	CreateInvitation(context.Context, *model.Invitation) error
	FindInvitationByTokenHash(context.Context, string) (*model.Invitation, error)
	SaveInvitation(context.Context, *model.Invitation) error
	CreatePasswordReset(context.Context, *model.PasswordReset) error
	FindPasswordResetByTokenHash(context.Context, string) (*model.PasswordReset, error)
	SavePasswordReset(context.Context, *model.PasswordReset) error
	CreateMFAFactor(context.Context, *model.MFAFactor) error
	FindActiveMFAFactor(context.Context, int64) (*model.MFAFactor, error)
	SaveMFAFactor(context.Context, *model.MFAFactor) error
	CreateAuditLog(context.Context, *model.AuditLog) error
	ListAuditLogs(context.Context, int64, int) ([]model.AuditLog, error)
	AddCasbinRule(context.Context, *model.CasbinRule) error
	ListCasbinRules(context.Context) ([]authorization.Rule, error)
}

type repository struct {
	db database.Executor
}

func New(db database.Executor) Repository {
	return &repository{db: db}
}

func (r *repository) WithExecutor(db database.Executor) Repository {
	return &repository{db: db}
}

func (r *repository) CreateOrganization(ctx context.Context, org *model.Organization) error {
	return r.db.Create(ctx, org)
}

func (r *repository) FindOrganizationByID(ctx context.Context, id int64) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.First(ctx, &org, database.Where("id = ?", id), alive()); err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *repository) FindOrganizationByCode(ctx context.Context, code string) (*model.Organization, error) {
	var org model.Organization
	if err := r.db.First(ctx, &org, database.Where("code = ?", code), alive()); err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *repository) ListOrganizations(ctx context.Context) ([]model.Organization, error) {
	var orgs []model.Organization
	err := r.db.Find(ctx, &orgs, alive(), database.Order("id DESC"))
	return orgs, err
}

func (r *repository) CreateUser(ctx context.Context, user *model.User) error {
	return r.db.Create(ctx, user)
}

func (r *repository) FindUserByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	if err := r.db.First(ctx, &user, database.Where("id = ?", id), alive()); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) FindUserByIdentifier(ctx context.Context, identifier string) (*model.User, error) {
	var user model.User
	if err := r.db.First(ctx, &user, database.Where("(username = ? OR email = ?)", identifier, identifier), alive()); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *repository) SaveUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now().UTC()
	return r.db.Save(ctx, user)
}

func (r *repository) CreateMembership(ctx context.Context, membership *model.Membership) error {
	return r.db.Create(ctx, membership)
}

func (r *repository) FindMembership(ctx context.Context, orgID, userID int64) (*model.Membership, error) {
	var membership model.Membership
	err := r.db.First(ctx, &membership,
		database.Where("org_id = ? AND user_id = ?", orgID, userID),
		database.Where("status = ?", model.StatusActive),
		alive(),
	)
	if err != nil {
		return nil, err
	}
	return &membership, nil
}

func (r *repository) ListMembershipsByUser(ctx context.Context, userID int64) ([]model.Membership, error) {
	var memberships []model.Membership
	err := r.db.Find(ctx, &memberships,
		database.Where("user_id = ?", userID),
		database.Where("status = ?", model.StatusActive),
		alive(),
	)
	return memberships, err
}

func (r *repository) ListUsersByOrg(ctx context.Context, orgID int64) ([]model.User, error) {
	var users []model.User
	_, err := r.db.Raw(ctx, &users, `
SELECT u.*
FROM iam_users u
JOIN iam_memberships m ON m.user_id = u.id
WHERE m.org_id = ? AND m.status = ? AND m.deleted_at IS NULL AND u.deleted_at IS NULL
ORDER BY u.id DESC`, orgID, model.StatusActive)
	return users, err
}

func (r *repository) CreateRole(ctx context.Context, role *model.Role) error {
	return r.db.Create(ctx, role)
}

func (r *repository) FindRole(ctx context.Context, orgID int64, code string) (*model.Role, error) {
	var role model.Role
	if err := r.db.First(ctx, &role, database.Where("org_id = ? AND code = ?", orgID, code), alive()); err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *repository) ListRoles(ctx context.Context, orgID int64) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Find(ctx, &roles, database.Where("org_id = ?", orgID), alive(), database.Order("code ASC"))
	return roles, err
}

func (r *repository) CreatePermission(ctx context.Context, permission *model.Permission) error {
	return r.db.Create(ctx, permission)
}

func (r *repository) FindPermission(ctx context.Context, code string) (*model.Permission, error) {
	var permission model.Permission
	if err := r.db.First(ctx, &permission, database.Where("code = ?", code)); err != nil {
		return nil, err
	}
	return &permission, nil
}

func (r *repository) ListPermissions(ctx context.Context) ([]model.Permission, error) {
	var permissions []model.Permission
	err := r.db.Find(ctx, &permissions, database.Order("code ASC"))
	return permissions, err
}

func (r *repository) CreateSession(ctx context.Context, session *model.Session) error {
	return r.db.Create(ctx, session)
}

func (r *repository) FindSessionByID(ctx context.Context, id int64) (*model.Session, error) {
	var session model.Session
	if err := r.db.First(ctx, &session, database.Where("id = ?", id)); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *repository) FindSessionByRefreshHash(ctx context.Context, hash string) (*model.Session, error) {
	var session model.Session
	if err := r.db.First(ctx, &session, database.Where("refresh_token_hash = ?", hash)); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *repository) ListSessionsByUser(ctx context.Context, userID int64) ([]model.Session, error) {
	var sessions []model.Session
	err := r.db.Find(ctx, &sessions, database.Where("user_id = ?", userID), database.Order("created_at DESC"))
	return sessions, err
}

func (r *repository) SaveSession(ctx context.Context, session *model.Session) error {
	session.UpdatedAt = time.Now().UTC()
	return r.db.Save(ctx, session)
}

func (r *repository) CreateInvitation(ctx context.Context, invitation *model.Invitation) error {
	return r.db.Create(ctx, invitation)
}

func (r *repository) FindInvitationByTokenHash(ctx context.Context, hash string) (*model.Invitation, error) {
	var invitation model.Invitation
	if err := r.db.First(ctx, &invitation, database.Where("token_hash = ?", hash)); err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *repository) SaveInvitation(ctx context.Context, invitation *model.Invitation) error {
	invitation.UpdatedAt = time.Now().UTC()
	return r.db.Save(ctx, invitation)
}

func (r *repository) CreatePasswordReset(ctx context.Context, reset *model.PasswordReset) error {
	return r.db.Create(ctx, reset)
}

func (r *repository) FindPasswordResetByTokenHash(ctx context.Context, hash string) (*model.PasswordReset, error) {
	var reset model.PasswordReset
	if err := r.db.First(ctx, &reset, database.Where("token_hash = ?", hash)); err != nil {
		return nil, err
	}
	return &reset, nil
}

func (r *repository) SavePasswordReset(ctx context.Context, reset *model.PasswordReset) error {
	reset.UpdatedAt = time.Now().UTC()
	return r.db.Save(ctx, reset)
}

func (r *repository) CreateMFAFactor(ctx context.Context, factor *model.MFAFactor) error {
	return r.db.Create(ctx, factor)
}

func (r *repository) FindActiveMFAFactor(ctx context.Context, userID int64) (*model.MFAFactor, error) {
	var factor model.MFAFactor
	err := r.db.First(ctx, &factor,
		database.Where("user_id = ?", userID),
		database.Where("type = ?", "totp"),
		database.Where("status = ?", model.StatusActive),
		database.Order("id DESC"),
	)
	if err != nil {
		return nil, err
	}
	return &factor, nil
}

func (r *repository) SaveMFAFactor(ctx context.Context, factor *model.MFAFactor) error {
	factor.UpdatedAt = time.Now().UTC()
	return r.db.Save(ctx, factor)
}

func (r *repository) CreateAuditLog(ctx context.Context, audit *model.AuditLog) error {
	return r.db.Create(ctx, audit)
}

func (r *repository) ListAuditLogs(ctx context.Context, orgID int64, limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var logs []model.AuditLog
	err := r.db.Find(ctx, &logs, database.Where("org_id = ?", orgID), database.Order("created_at DESC"), database.Limit(limit))
	return logs, err
}

func (r *repository) AddCasbinRule(ctx context.Context, rule *model.CasbinRule) error {
	var existing model.CasbinRule
	err := r.db.First(ctx, &existing,
		database.Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ? AND v3 = ? AND v4 = ? AND v5 = ?",
			rule.PType, rule.V0, rule.V1, rule.V2, rule.V3, rule.V4, rule.V5),
	)
	if err == nil {
		return nil
	}
	if !errors.Is(err, database.ErrNotFound) {
		return err
	}
	return r.db.Create(ctx, rule)
}

func (r *repository) ListCasbinRules(ctx context.Context) ([]authorization.Rule, error) {
	var rows []model.CasbinRule
	if err := r.db.Find(ctx, &rows, database.Order("id ASC")); err != nil {
		return nil, err
	}
	rules := make([]authorization.Rule, 0, len(rows))
	for _, row := range rows {
		values := []string{row.V0, row.V1, row.V2, row.V3, row.V4, row.V5}
		switch row.PType {
		case "p":
			values = values[:4]
		case "g":
			values = values[:3]
		}
		rules = append(rules, authorization.Rule{PType: row.PType, Values: values})
	}
	return rules, nil
}

func alive() database.QueryOption {
	return database.Where("deleted_at IS NULL")
}
