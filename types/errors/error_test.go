package errors

// 本测试文件固定跨包公共类型的导入边界和响应契约，防止注释补全和后续重构改变外部可观察行为。

import (
	stderrors "errors"
	"strings"
	"testing"
)

// TestBizErrorContract 固定跨包公共类型的导入边界和响应契约，确保后续注释补全或结构调整不改变该场景。
func TestBizErrorContract(t *testing.T) {
	cause := stderrors.New("database offline")
	err := NewBizError(ErrDatabaseError, "database failed").WithCause(cause)

	if err.Code != ErrDatabaseError {
		t.Fatalf("Code = %d, want %d", err.Code, ErrDatabaseError)
	}
	if err.Message != "database failed" {
		t.Fatalf("Message = %q, want database failed", err.Message)
	}
	if !stderrors.Is(err, cause) {
		t.Fatal("BizError should unwrap the cause error")
	}
	if got := err.Error(); !strings.Contains(got, "[5001] database failed: database offline") {
		t.Fatalf("Error() = %q, want code, message, and cause", got)
	}
}

// TestErrorCodeRangesContract 固定跨包公共类型的导入边界和响应契约，确保后续注释补全或结构调整不改变该场景。
func TestErrorCodeRangesContract(t *testing.T) {
	tests := []struct {
		name string
		code int
		min  int
		max  int
	}{
		{"invalid params", ErrInvalidParams, 1000, 1999},
		{"business logic", ErrBusinessLogic, 2000, 2999},
		{"unauthorized", ErrUnauthorized, 3000, 3999},
		{"invalid token", ErrInvalidToken, 3000, 3999},
		{"token expired", ErrTokenExpired, 3000, 3999},
		{"permission denied", ErrPermissionDenied, 3000, 3999},
		{"resource not found", ErrResourceNotFound, 4000, 4999},
		{"internal server", ErrInternalServer, 5000, 5999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code < tt.min || tt.code > tt.max {
				t.Fatalf("code %d outside range %d-%d", tt.code, tt.min, tt.max)
			}
		})
	}
}
