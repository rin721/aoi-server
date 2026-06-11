package internal_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

const modulePath = "github.com/rei0721/go-scaffold"

func TestInternalPackagesDoNotImportThirdPartyInfrastructure(t *testing.T) {
	files, err := goFilesUnder(".")
	if err != nil {
		t.Fatalf("collect internal go files: %v", err)
	}

	for _, file := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s imports: %v", file, err)
		}

		for _, spec := range parsed.Imports {
			path := strings.Trim(spec.Path.Value, `"`)
			if isThirdPartyImport(path) {
				t.Fatalf("internal package must use pkg anti-corruption wrappers instead of third-party import %q from %s", path, file)
			}
		}
	}
}

func goFilesUnder(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func isThirdPartyImport(path string) bool {
	if strings.HasPrefix(path, modulePath) {
		return false
	}
	first := path
	if idx := strings.Index(first, "/"); idx >= 0 {
		first = first[:idx]
	}
	return strings.Contains(first, ".")
}
