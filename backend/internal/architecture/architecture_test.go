package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestFeatureAndMockPaymentDependencyBoundary(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve backend root: %v", err)
	}

	requireNoImports(
		t,
		filepath.Join(root, "internal", "feature"),
		"pinshaoshao/internal/mockpayment",
	)
	requireNoImports(
		t,
		filepath.Join(root, "internal", "mockpayment"),
		"pinshaoshao/internal/feature",
	)
}

func requireNoImports(t *testing.T, root string, forbiddenPrefix string) {
	t.Helper()

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") ||
			strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imported := range file.Imports {
			value, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				return err
			}
			if value == forbiddenPrefix || strings.HasPrefix(value, forbiddenPrefix+"/") {
				relative, relErr := filepath.Rel(filepath.Dir(root), path)
				if relErr != nil {
					relative = path
				}
				t.Errorf("%s imports forbidden package %s", relative, value)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect imports under %s: %v", root, err)
	}
}
