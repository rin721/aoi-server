package i18n

// 本测试文件固定国际化工具和基础设施辅助函数的边界，防止注释补全和后续重构改变外部可观察行为。

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewLoadsMessagesAndTranslates 固定国际化工具和基础设施辅助函数的边界，确保后续注释补全或结构调整不改变该场景。
func TestNewLoadsMessagesAndTranslates(t *testing.T) {
	dir := t.TempDir()
	writeMessageFile(t, filepath.Join(dir, "en-US.json"), `{
		"welcome": "Hello, {{.Name}}!",
		"plain": "Plain message"
	}`)
	writeMessageFile(t, filepath.Join(dir, "zh-CN.yaml"), "welcome: Ni hao, {{.Name}}!\n")

	translator, err := New(&Config{
		DefaultLanguage:    LanguageEnglish,
		SupportedLanguages: []string{LanguageEnglish, LanguageChinese},
		MessagesDir:        dir,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if got := translator.GetDefaultLanguage(); got != LanguageEnglish {
		t.Fatalf("GetDefaultLanguage() = %q, want %q", got, LanguageEnglish)
	}
	if !translator.IsSupported(LanguageChinese) {
		t.Fatalf("IsSupported(%q) = false, want true", LanguageChinese)
	}
	if got := translator.T(LanguageEnglish, "welcome", map[string]interface{}{"Name": "Ada"}); got != "Hello, Ada!" {
		t.Fatalf("T(en-US, welcome) = %q, want %q", got, "Hello, Ada!")
	}
	if got := translator.T("fr-FR", "welcome", map[string]interface{}{"Name": "Ada"}); got != "Hello, Ada!" {
		t.Fatalf("T(unsupported, welcome) = %q, want default language translation", got)
	}
	if got := translator.T(LanguageEnglish, "missing.key"); got != "missing.key" {
		t.Fatalf("T(missing.key) = %q, want message ID fallback", got)
	}
}

// TestMustTPanicsWhenTranslationIsMissing 固定国际化工具和基础设施辅助函数的边界，确保后续注释补全或结构调整不改变该场景。
func TestMustTPanicsWhenTranslationIsMissing(t *testing.T) {
	translator := Default()

	defer func() {
		if recover() == nil {
			t.Fatal("MustT() did not panic for missing translation")
		}
	}()

	translator.MustT(LanguageEnglish, "missing.key")
}

// TestLoadMessagesReturnsErrorsForMissingAndEmptyDirectory 固定国际化工具和基础设施辅助函数的边界，确保后续注释补全或结构调整不改变该场景。
func TestLoadMessagesReturnsErrorsForMissingAndEmptyDirectory(t *testing.T) {
	translator := Default()

	if err := translator.LoadMessages(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("LoadMessages(missing) error = nil, want error")
	}
	if err := translator.LoadMessages(t.TempDir()); err == nil {
		t.Fatal("LoadMessages(empty) error = nil, want error")
	}
}

// writeMessageFile 写入测试夹具文件，并把文件系统准备细节限制在测试辅助层。
func writeMessageFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
}
