package service_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/rei0721/go-scaffold/internal/app/dbapp"
	"github.com/rei0721/go-scaffold/internal/app/testsupport"
	"github.com/rei0721/go-scaffold/internal/modules/demo/repository"
	"github.com/rei0721/go-scaffold/internal/modules/demo/service"
	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/pkg/database"
)

func TestCustomerServiceVisibilityAndCRUD(t *testing.T) {
	customerService := newCustomerService(t)
	ctx := context.Background()
	webPrincipal := iamservice.Principal{UserID: 1, OrgID: 10, Username: "alice"}
	adminPrincipal := iamservice.Principal{UserID: 1, OrgID: 10, Username: "alice", RoleCode: "admin"}
	sameRolePrincipal := iamservice.Principal{UserID: 2, OrgID: 10, Username: "bob", RoleCode: "admin"}
	plainOtherPrincipal := iamservice.Principal{UserID: 2, OrgID: 10, Username: "bob"}
	otherOrgPrincipal := iamservice.Principal{UserID: 3, OrgID: 20, Username: "carol", RoleCode: "admin"}

	webOwned, err := customerService.Create(ctx, service.CreateCustomerInput{
		Principal:         webPrincipal,
		CustomerName:      "  Web 客户  ",
		CustomerPhoneData: "  13800000000  ",
	})
	if err != nil {
		t.Fatalf("create web customer: %v", err)
	}
	if webOwned.CustomerName != "Web 客户" || webOwned.CustomerPhoneData != "13800000000" {
		t.Fatalf("expected trimmed customer fields, got %#v", webOwned)
	}

	roleOwned, err := customerService.Create(ctx, service.CreateCustomerInput{
		Principal:         adminPrincipal,
		CustomerName:      "角色客户",
		CustomerPhoneData: "13900000000",
	})
	if err != nil {
		t.Fatalf("create role customer: %v", err)
	}

	ownerPage, err := customerService.List(ctx, service.ListCustomerInput{Principal: adminPrincipal, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list owner customers: %v", err)
	}
	if ownerPage.Total != 2 {
		t.Fatalf("owner total = %d, want 2", ownerPage.Total)
	}

	sameRolePage, err := customerService.List(ctx, service.ListCustomerInput{Principal: sameRolePrincipal, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list same-role customers: %v", err)
	}
	if sameRolePage.Total != 1 || sameRolePage.Items[0].ID != roleOwned.ID {
		t.Fatalf("same-role page = %#v, want only role-owned customer", sameRolePage)
	}

	plainOtherPage, err := customerService.List(ctx, service.ListCustomerInput{Principal: plainOtherPrincipal, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list plain other customers: %v", err)
	}
	if plainOtherPage.Total != 0 {
		t.Fatalf("plain other total = %d, want 0", plainOtherPage.Total)
	}

	otherOrgPage, err := customerService.List(ctx, service.ListCustomerInput{Principal: otherOrgPrincipal, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list other org customers: %v", err)
	}
	if otherOrgPage.Total != 0 {
		t.Fatalf("other org total = %d, want 0", otherOrgPage.Total)
	}

	updatedName := "更新客户"
	updated, err := customerService.Update(ctx, service.UpdateCustomerInput{
		Principal:    webPrincipal,
		ID:           webOwned.ID,
		CustomerName: &updatedName,
	})
	if err != nil {
		t.Fatalf("update customer: %v", err)
	}
	if updated.CustomerName != updatedName {
		t.Fatalf("updated name = %q, want %q", updated.CustomerName, updatedName)
	}

	if _, err := customerService.Get(ctx, service.CustomerIdentityInput{Principal: plainOtherPrincipal, ID: roleOwned.ID}); !errors.Is(err, service.ErrCustomerNotFound) {
		t.Fatalf("plain other get error = %v, want ErrCustomerNotFound", err)
	}
	if err := customerService.Delete(ctx, service.CustomerIdentityInput{Principal: adminPrincipal, ID: roleOwned.ID}); err != nil {
		t.Fatalf("delete role customer: %v", err)
	}
	if _, err := customerService.Get(ctx, service.CustomerIdentityInput{Principal: adminPrincipal, ID: roleOwned.ID}); !errors.Is(err, service.ErrCustomerNotFound) {
		t.Fatalf("deleted customer get error = %v, want ErrCustomerNotFound", err)
	}
}

func TestCustomerServiceValidation(t *testing.T) {
	customerService := newCustomerService(t)
	ctx := context.Background()
	principal := iamservice.Principal{UserID: 1, OrgID: 10, Username: "alice"}

	if _, err := customerService.Create(ctx, service.CreateCustomerInput{Principal: principal, CustomerPhoneData: "138"}); !errors.Is(err, service.ErrCustomerNameRequired) {
		t.Fatalf("blank name error = %v, want ErrCustomerNameRequired", err)
	}
	if _, err := customerService.Create(ctx, service.CreateCustomerInput{Principal: principal, CustomerName: "客户"}); !errors.Is(err, service.ErrCustomerPhoneRequired) {
		t.Fatalf("blank phone error = %v, want ErrCustomerPhoneRequired", err)
	}
	if _, err := customerService.Create(ctx, service.CreateCustomerInput{CustomerName: "客户", CustomerPhoneData: "138"}); !errors.Is(err, service.ErrCustomerPrincipalRequired) {
		t.Fatalf("missing principal error = %v, want ErrCustomerPrincipalRequired", err)
	}
}

func newCustomerService(t *testing.T) service.CustomerService {
	t.Helper()

	db, err := database.New(&database.Config{
		Driver: database.DriverSQLite,
		DBName: filepath.Join(t.TempDir(), "demo-customer.db"),
	})
	if err != nil {
		t.Fatalf("create sqlite database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close sqlite database: %v", err)
		}
	})

	if _, err := dbapp.ApplyDemoSchema(context.Background(), db, string(database.DriverSQLite)); err != nil {
		t.Fatalf("apply demo schema: %v", err)
	}

	moduleDB := testsupport.Database(db)
	return service.NewCustomerService(moduleDB, repository.NewCustomerRepository(moduleDB))
}
