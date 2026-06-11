package utils

// 本测试文件固定国际化工具和基础设施辅助函数的边界，防止注释补全和后续重构改变外部可观察行为。

import (
	"encoding/hex"
	"net"
	"strconv"
	"testing"
)

type fakeI18n struct {
	lang         string
	messageID    string
	templateData []map[string]interface{}
}

// T 实现国际化测试桩的同名方法，保证 utils 代理层只依赖接口契约。
func (f *fakeI18n) T(lang string, messageID string, templateData ...map[string]interface{}) string {
	f.lang = lang
	f.messageID = messageID
	f.templateData = append([]map[string]interface{}(nil), templateData...)
	return lang + ":" + messageID
}

// MustT 实现国际化测试桩的同名方法，保证 utils 代理层只依赖接口契约。
func (f *fakeI18n) MustT(lang string, messageID string, templateData ...map[string]interface{}) string {
	return f.T(lang, messageID, templateData...)
}

// IsSupported 实现国际化测试桩的同名方法，保证 utils 代理层只依赖接口契约。
func (f *fakeI18n) IsSupported(string) bool {
	return true
}

// GetDefaultLanguage 实现国际化测试桩的同名方法，保证 utils 代理层只依赖接口契约。
func (f *fakeI18n) GetDefaultLanguage() string {
	return "unused"
}

// LoadMessages 实现国际化测试桩的同名方法，保证 utils 代理层只依赖接口契约。
func (f *fakeI18n) LoadMessages(string) error {
	return nil
}

// TestSnowflakeGeneratesIDsAndRejectsInvalidNode 固定国际化工具和基础设施辅助函数的边界，确保后续注释补全或结构调整不改变该场景。
func TestSnowflakeGeneratesIDsAndRejectsInvalidNode(t *testing.T) {
	gen, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("NewSnowflake() error = %v", err)
	}

	first := gen.NextID()
	second := gen.NextID()
	if first <= 0 || second <= 0 {
		t.Fatalf("NextID() generated non-positive IDs: %d, %d", first, second)
	}
	if first == second {
		t.Fatalf("NextID() generated duplicate IDs: %d", first)
	}

	asString := gen.NextIDString()
	parsed, err := strconv.ParseInt(asString, 10, 64)
	if err != nil {
		t.Fatalf("NextIDString() = %q, parse error = %v", asString, err)
	}
	if parsed <= 0 {
		t.Fatalf("NextIDString() parsed to non-positive ID: %d", parsed)
	}

	defaultGen := DefaultSnowflake()
	if defaultGen.NextID() <= 0 {
		t.Fatal("DefaultSnowflake().NextID() generated a non-positive ID")
	}

	if _, err := NewSnowflake(1024); err == nil {
		t.Fatal("NewSnowflake(1024) error = nil, want invalid node error")
	}
}

// TestListenAddrValidation 固定国际化工具和基础设施辅助函数的边界，确保后续注释补全或结构调整不改变该场景。
func TestListenAddrValidation(t *testing.T) {
	valid := []string{
		":8080",
		"0.0.0.0:8080",
		"127.0.0.1:8080",
		"localhost:8080",
		"[::]:8080",
	}
	for _, addr := range valid {
		if err := IsValidListenAddr(addr); err != nil {
			t.Fatalf("IsValidListenAddr(%q) error = %v", addr, err)
		}
	}

	invalid := []string{
		"",
		"invalid",
		":0",
		":65536",
		"not-a-host:8080",
		"203.0.113.1:8080",
	}
	for _, addr := range invalid {
		if err := IsValidListenAddr(addr); err == nil {
			t.Fatalf("IsValidListenAddr(%q) error = nil, want error", addr)
		}
	}
}

// TestHTTPListenAddrValidation 固定国际化工具和基础设施辅助函数的边界，确保后续注释补全或结构调整不改变该场景。
func TestHTTPListenAddrValidation(t *testing.T) {
	if err := IsValidHTTPListenAddr("127.0.0.1:0"); err != nil {
		t.Fatalf("IsValidHTTPListenAddr() with ephemeral loopback error = %v", err)
	}

	if err := IsValidHTTPListenAddr(""); err == nil {
		t.Fatal("IsValidHTTPListenAddr(\"\") error = nil, want error")
	}

	if err := IsValidHTTPListenAddr("203.0.113.1:0"); err == nil {
		t.Fatal("IsValidHTTPListenAddr(\"203.0.113.1:0\") error = nil, want bind error")
	}
}

// TestGetAvailablePortHonorsRangeAndExclude 固定国际化工具和基础设施辅助函数的边界，确保后续注释补全或结构调整不改变该场景。
func TestGetAvailablePortHonorsRangeAndExclude(t *testing.T) {
	invalidRanges := [][2]int{
		{0, 1},
		{10002, 10001},
		{1, 65536},
	}
	for _, ports := range invalidRanges {
		if _, err := GetAvailablePort(ports[0], ports[1]); err == nil {
			t.Fatalf("GetAvailablePort(%d, %d) error = nil, want error", ports[0], ports[1])
		}
	}

	port := freeLoopbackPort(t)
	got, err := GetAvailablePort(port, port)
	if err != nil {
		t.Fatalf("GetAvailablePort(%d, %d) error = %v", port, port, err)
	}
	if got != port {
		t.Fatalf("GetAvailablePort(%d, %d) = %d, want %d", port, port, got, port)
	}

	if _, err := GetAvailablePort(port, port, port); err == nil {
		t.Fatalf("GetAvailablePort(%d, %d, %d) error = nil, want no available port", port, port, port)
	}
}

// TestGenerateDeviceIDIsStableAndSalted 固定国际化工具和基础设施辅助函数的边界，确保后续注释补全或结构调整不改变该场景。
func TestGenerateDeviceIDIsStableAndSalted(t *testing.T) {
	first := GenerateDeviceID("app-a")
	second := GenerateDeviceID("app-a")
	third := GenerateDeviceID("app-b")

	if first != second {
		t.Fatalf("GenerateDeviceID() is not stable for the same salt: %q != %q", first, second)
	}
	if first == third {
		t.Fatal("GenerateDeviceID() returned the same value for different salts")
	}
	if len(first) != 64 {
		t.Fatalf("GenerateDeviceID() length = %d, want 64", len(first))
	}
	if _, err := hex.DecodeString(first); err != nil {
		t.Fatalf("GenerateDeviceID() = %q, invalid hex: %v", first, err)
	}
}

// TestI18nUtilsUsesDefaultLanguageAndForwardsTemplates 固定国际化工具和基础设施辅助函数的边界，确保后续注释补全或结构调整不改变该场景。
func TestI18nUtilsUsesDefaultLanguageAndForwardsTemplates(t *testing.T) {
	backend := &fakeI18n{}
	utils := NewI18nUtils(backend, "zh-CN")

	template := map[string]interface{}{"Name": "Codex"}
	got := utils.T("welcome", template)
	if got != "zh-CN:welcome" {
		t.Fatalf("T() = %q, want %q", got, "zh-CN:welcome")
	}
	if backend.lang != "zh-CN" {
		t.Fatalf("backend lang = %q, want zh-CN", backend.lang)
	}
	if backend.messageID != "welcome" {
		t.Fatalf("backend messageID = %q, want welcome", backend.messageID)
	}
	if len(backend.templateData) != 1 {
		t.Fatalf("backend template count = %d, want 1", len(backend.templateData))
	}
	if backend.templateData[0]["Name"] != "Codex" {
		t.Fatalf("backend template Name = %v, want Codex", backend.templateData[0]["Name"])
	}
}

// freeLoopbackPort 是当前测试文件的辅助函数，用于复用夹具、断言或输入构造逻辑。
func freeLoopbackPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	port := portOf(t, listener.Addr())
	if err := listener.Close(); err != nil {
		t.Fatalf("listener.Close() error = %v", err)
	}
	return port
}

// portOf 是当前测试文件的辅助函数，用于复用夹具、断言或输入构造逻辑。
func portOf(t *testing.T, addr net.Addr) int {
	t.Helper()

	_, portString, err := net.SplitHostPort(addr.String())
	if err != nil {
		t.Fatalf("net.SplitHostPort(%q) error = %v", addr.String(), err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatalf("strconv.Atoi(%q) error = %v", portString, err)
	}
	return port
}
