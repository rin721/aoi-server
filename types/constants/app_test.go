package constants

// 本测试文件固定跨包公共类型的导入边界和响应契约，防止注释补全和后续重构改变外部可观察行为。

import "testing"

// TestApplicationConstants 固定跨包公共类型的导入边界和响应契约，确保后续注释补全或结构调整不改变该场景。
func TestApplicationConstants(t *testing.T) {
	if AppPrefix != "Rin" {
		t.Fatalf("AppPrefix = %q, want %q", AppPrefix, "Rin")
	}

	if AppServerCommandName != "server" {
		t.Fatalf("AppServerCommandName = %q, want %q", AppServerCommandName, "server")
	}

	if EnvConfigPathName != "RIN_CONFIG_PATH" {
		t.Fatalf("EnvConfigPathName = %q, want %q", EnvConfigPathName, "RIN_CONFIG_PATH")
	}
}
