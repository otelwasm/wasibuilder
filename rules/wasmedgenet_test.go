package rules

import (
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestProcessNetPackageAddsFilePosixWhenPresent(t *testing.T) {
	root := t.TempDir()
	filePosix := filepath.Join(root, "src/net/file_posix.go")
	if err := os.MkdirAll(filepath.Dir(filePosix), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filePosix, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	ctx := &ExecContext{
		Args:    []string{filepath.Join(root, "src/net/net_fake.go")},
		Package: "net",
	}
	if err := (&WASMEdgeNet{}).processNetPackage(ctx, slog.Default()); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(ctx.Args, filePosix) {
		t.Fatalf("file_posix.go was not added: %v", ctx.Args)
	}
}
