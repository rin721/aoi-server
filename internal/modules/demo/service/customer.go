package service

import (
	"context"
	"errors"
	"strings"

	"github.com/rei0721/go-scaffold/internal/modules/demo/model"
	"github.com/rei0721/go-scaffold/internal/modules/demo/repository"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/pkg/database"
)

const (
	defaultCustomerPageSize = 10
	maxCustomerPageSize     = 100
)

var (
	ErrCustomerPrincipalRequired = errors.New("customer principal is required")
	ErrCustomerNameRequired      = errors.New("customer name is required")
	ErrCustomerPhoneRequired     = errors.New("customer phone is required")
	ErrCustomerNotFound          = errors.New("customer not found")
)

// CustomerService 提供受权限保护的客户资源示例业务能力。
type CustomerService interface {
	Create(ctx context.Context, input CreateCustomerInput) (*model.Customer, error)
	List(ctx context.Context, input ListCustomerInput) (CustomerPage, error)
	Get(ctx context.Context, input CustomerIdentityInput) (*model.Customer, error)
	Update(ctx context.Context, input UpdateCustomerInput) (*model.Customer, error)
	Delete(ctx context.Context, input CustomerIdentityInput) error
}

// CreateCustomerInput 描述创建客户资源所需字段。
type CreateCustomerInput struct {
	Principal         iamservice.Principal
	CustomerName      string
	CustomerPhoneData string
}

// ListCustomerInput 描述客户资源分页和过滤条件。
type ListCustomerInput struct {
	Principal iamservice.Principal
	Keyword   string
	Page      int
	PageSize  int
}

// CustomerIdentityInput 描述需要定位单条客户资源的请求。
type CustomerIdentityInput struct {
	Principal iamservice.Principal
	ID        uint
}

// UpdateCustomerInput 描述客户资源局部更新字段。
type UpdateCustomerInput struct {
	Principal         iamservice.Principal
	ID                uint
	CustomerName      *string
	CustomerPhoneData *string
}

// CustomerPage 是客户资源列表的分页响应。
type CustomerPage struct {
	Items         []model.Customer `json:"items"`
	Page          int              `json:"page"`
	PageSize      int              `json:"pageSize"`
	StorageStatus string           `json:"storageStatus"`
	Total         int64            `json:"total"`
}

type customerService struct {
	db   database.Database
	repo repository.CustomerRepository
}

// NewCustomerService 创建客户资源示例服务。
func NewCustomerService(db database.Database, repo repository.CustomerRepository) CustomerService {
	return &customerService{db: db, repo: repo}
}

func (s *customerService) Create(ctx context.Context, input CreateCustomerInput) (*model.Customer, error) {
	if err := validateCustomerPrincipal(input.Principal); err != nil {
		return nil, err
	}
	name, phone, err := normalizeCustomerPayload(input.CustomerName, input.CustomerPhoneData)
	if err != nil {
		return nil, err
	}

	customer := &model.Customer{
		CustomerName:      name,
		CustomerPhoneData: phone,
		OwnerUserID:       input.Principal.UserID,
		OwnerUsername:     strings.TrimSpace(input.Principal.Username),
		OwnerRoleCode:     strings.TrimSpace(input.Principal.RoleCode),
		OrgID:             input.Principal.OrgID,
	}

	if err := s.db.WithTx(ctx, func(ctx context.Context, tx database.Executor) error {
		return s.repo.WithExecutor(tx).Create(ctx, customer)
	}); err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *customerService) List(ctx context.Context, input ListCustomerInput) (CustomerPage, error) {
	if err := validateCustomerPrincipal(input.Principal); err != nil {
		return CustomerPage{}, err
	}
	page, pageSize := normalizeCustomerPage(input.Page, input.PageSize)
	items, total, err := s.repo.ListVisible(ctx, repository.CustomerFilter{
		OrgID:       input.Principal.OrgID,
		UserID:      input.Principal.UserID,
		RoleCode:    input.Principal.RoleCode,
		Keyword:     input.Keyword,
		Limit:       pageSize,
		Offset:      (page - 1) * pageSize,
		WithPaging:  true,
		WithKeyword: true,
	})
	if err != nil {
		return CustomerPage{}, err
	}
	return CustomerPage{Items: items, Page: page, PageSize: pageSize, StorageStatus: "persisted", Total: total}, nil
}

func (s *customerService) Get(ctx context.Context, input CustomerIdentityInput) (*model.Customer, error) {
	if err := validateCustomerPrincipal(input.Principal); err != nil {
		return nil, err
	}
	customer, err := s.repo.FindVisibleByID(ctx, visibleCustomerFilter(input.Principal), input.ID)
	return customer, normalizeCustomerNotFound(err)
}

func (s *customerService) Update(ctx context.Context, input UpdateCustomerInput) (*model.Customer, error) {
	if err := validateCustomerPrincipal(input.Principal); err != nil {
		return nil, err
	}
	var customer *model.Customer
	err := s.db.WithTx(ctx, func(ctx context.Context, tx database.Executor) error {
		txRepo := s.repo.WithExecutor(tx)
		current, err := txRepo.FindVisibleByID(ctx, visibleCustomerFilter(input.Principal), input.ID)
		if err != nil {
			return normalizeCustomerNotFound(err)
		}
		if input.CustomerName != nil {
			name := strings.TrimSpace(*input.CustomerName)
			if name == "" {
				return ErrCustomerNameRequired
			}
			current.CustomerName = name
		}
		if input.CustomerPhoneData != nil {
			phone := strings.TrimSpace(*input.CustomerPhoneData)
			if phone == "" {
				return ErrCustomerPhoneRequired
			}
			current.CustomerPhoneData = phone
		}
		if err := txRepo.Update(ctx, current); err != nil {
			return err
		}
		customer = current
		return nil
	})
	if err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *customerService) Delete(ctx context.Context, input CustomerIdentityInput) error {
	if err := validateCustomerPrincipal(input.Principal); err != nil {
		return err
	}
	return s.db.WithTx(ctx, func(ctx context.Context, tx database.Executor) error {
		txRepo := s.repo.WithExecutor(tx)
		if _, err := txRepo.FindVisibleByID(ctx, visibleCustomerFilter(input.Principal), input.ID); err != nil {
			return normalizeCustomerNotFound(err)
		}
		return txRepo.Delete(ctx, input.ID)
	})
}

func validateCustomerPrincipal(principal iamservice.Principal) error {
	if principal.UserID <= 0 || principal.OrgID <= 0 {
		return ErrCustomerPrincipalRequired
	}
	return nil
}

func normalizeCustomerPayload(name string, phone string) (string, string, error) {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	if name == "" {
		return "", "", ErrCustomerNameRequired
	}
	if phone == "" {
		return "", "", ErrCustomerPhoneRequired
	}
	return name, phone, nil
}

func normalizeCustomerPage(page int, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = defaultCustomerPageSize
	}
	if pageSize > maxCustomerPageSize {
		pageSize = maxCustomerPageSize
	}
	return page, pageSize
}

func visibleCustomerFilter(principal iamservice.Principal) repository.CustomerFilter {
	return repository.CustomerFilter{
		OrgID:    principal.OrgID,
		UserID:   principal.UserID,
		RoleCode: principal.RoleCode,
	}
}

func normalizeCustomerNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, database.ErrNotFound) {
		return ErrCustomerNotFound
	}
	return err
}
