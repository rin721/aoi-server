package authorization

import (
	"context"
	"testing"
)

func TestDomainRBAC(t *testing.T) {
	enforcer, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	ctx := context.Background()

	if _, err := enforcer.AddPolicy(ctx, "role:admin", "100", "user", "read|invite"); err != nil {
		t.Fatalf("AddPolicy() failed: %v", err)
	}
	if _, err := enforcer.AddRoleForUser(ctx, "user:1", "role:admin", "100"); err != nil {
		t.Fatalf("AddRoleForUser() failed: %v", err)
	}

	allowed, err := enforcer.Enforce(ctx, "user:1", "100", "user", "read")
	if err != nil {
		t.Fatalf("Enforce(read) failed: %v", err)
	}
	if !allowed {
		t.Fatal("expected user:1 to read users in org 100")
	}

	allowed, err = enforcer.Enforce(ctx, "user:1", "200", "user", "read")
	if err != nil {
		t.Fatalf("Enforce(other org) failed: %v", err)
	}
	if allowed {
		t.Fatal("domain isolation failed")
	}
}

func TestLoadRules(t *testing.T) {
	enforcer, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	err = enforcer.LoadRules(context.Background(), []Rule{
		{PType: "p", Values: []string{"role:owner", "1", "*", "*"}},
		{PType: "g", Values: []string{"user:9", "role:owner", "1"}},
	})
	if err != nil {
		t.Fatalf("LoadRules() failed: %v", err)
	}
	allowed, err := enforcer.Enforce(context.Background(), "user:9", "1", "audit", "read")
	if err != nil {
		t.Fatalf("Enforce() failed: %v", err)
	}
	if !allowed {
		t.Fatal("owner wildcard policy should allow audit read")
	}
}
