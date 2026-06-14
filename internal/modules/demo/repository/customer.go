package repository

import (
	"context"
	"strings"
	"time"

	"github.com/rei0721/go-scaffold/internal/modules/demo/model"
	database "github.com/rei0721/go-scaffold/internal/ports"
)

// CustomerRepository 定义客户资源示例的持久化端口。
type CustomerRepository interface {
	WithExecutor(db database.Executor) CustomerRepository
	Create(ctx context.Context, customer *model.Customer) error
	ListVisible(ctx context.Context, filter CustomerFilter) ([]model.Customer, int64, error)
	FindVisibleByID(ctx context.Context, filter CustomerFilter, id uint) (*model.Customer, error)
	Update(ctx context.Context, customer *model.Customer) error
	Delete(ctx context.Context, id uint) error
}

// CustomerFilter 描述客户资源列表和单条查询的可见范围。
type CustomerFilter struct {
	OrgID       int64
	UserID      int64
	RoleCode    string
	Keyword     string
	Limit       int
	Offset      int
	WithPaging  bool
	WithKeyword bool
}

type customerRepository struct {
	db database.Executor
}

// NewCustomerRepository 创建客户资源仓储。
func NewCustomerRepository(db database.Executor) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) WithExecutor(db database.Executor) CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, customer *model.Customer) error {
	return r.db.Create(ctx, customer)
}

func (r *customerRepository) ListVisible(ctx context.Context, filter CustomerFilter) ([]model.Customer, int64, error) {
	opts := customerQueryOptions(filter)
	total, err := r.db.Count(ctx, &model.Customer{}, opts...)
	if err != nil {
		return nil, 0, err
	}
	if filter.WithPaging {
		opts = append(opts, database.Limit(filter.Limit), database.Offset(filter.Offset))
	}
	opts = append(opts, database.Order("created_at DESC, id DESC"))

	var customers []model.Customer
	if err := r.db.Find(ctx, &customers, opts...); err != nil {
		return nil, 0, err
	}
	return customers, total, nil
}

func (r *customerRepository) FindVisibleByID(ctx context.Context, filter CustomerFilter, id uint) (*model.Customer, error) {
	opts := customerQueryOptions(filter)
	opts = append(opts, database.Where("id = ?", id))

	var customer model.Customer
	if err := r.db.First(ctx, &customer, opts...); err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *customerRepository) Update(ctx context.Context, customer *model.Customer) error {
	return r.db.Save(ctx, customer)
}

func (r *customerRepository) Delete(ctx context.Context, id uint) error {
	now := time.Now().UTC()
	_, err := r.db.Update(ctx, &model.Customer{}, map[string]any{
		"deleted_at": now,
		"updated_at": now,
	}, database.Where("id = ?", id), aliveCustomers())
	return err
}

func customerQueryOptions(filter CustomerFilter) []database.QueryOption {
	opts := []database.QueryOption{aliveCustomers(), customerVisibility(filter.OrgID, filter.UserID, filter.RoleCode)}
	keyword := strings.TrimSpace(filter.Keyword)
	if filter.WithKeyword && keyword != "" {
		like := "%" + keyword + "%"
		opts = append(opts, database.Where("(customer_name LIKE ? OR customer_phone_data LIKE ? OR owner_username LIKE ?)", like, like, like))
	}
	return opts
}

func customerVisibility(orgID int64, userID int64, roleCode string) database.QueryOption {
	roleCode = strings.TrimSpace(roleCode)
	if roleCode == "" {
		return database.Where("org_id = ? AND owner_user_id = ?", orgID, userID)
	}
	return database.Where("org_id = ? AND (owner_user_id = ? OR owner_role_code = ?)", orgID, userID, roleCode)
}

func aliveCustomers() database.QueryOption {
	return database.Where("deleted_at IS NULL")
}
