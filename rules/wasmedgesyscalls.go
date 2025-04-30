package rules

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

//go:embed wasmedgesyscalls/*
var fs embed.FS

var wasmEdgeSyscallsReplaceFiles = map[string]string{
	"src/syscall/net_fake.go":   "net_fake.go.replaced",
	"src/syscall/net_wasip1.go": "net_wasip1.go.replaced",
}

type WASMEdgeSyscalls struct {
}

func (w *WASMEdgeSyscalls) Apply(ctx *ExecContext) error {
	// We're primarily interested in the compile tool
	if filepath.Base(ctx.Command) != "compile" {
		return nil
	}

	// If we couldn't determine the package path, return unmodified args
	if ctx.Package != "syscall" {
		return nil
	}

	for src, dst := range wasmEdgeSyscallsReplaceFiles {
		// Read the file from the embedded filesystem
		content, err := fs.ReadFile("wasmedgesyscalls/" + dst)
		if err != nil {
			slog.Error("Error reading embedded file", "file", dst, "error", err)
			return err
		}
		// Prepare a temporary file with the content
		tmpFile, err := w.prepareTmpFile(string(content))
		if err != nil {
			slog.Error("Error preparing temp file", "file", dst, "error", err)
			return err
		}

		for i, arg := range ctx.Args {
			if strings.HasSuffix(arg, src) {
				// Replace the argument with the temporary file path
				ctx.Args[i] = tmpFile
				slog.Info("Replaced argument", "original", arg, "replacement", tmpFile)
			}
		}
	}
	slog.Info("package is...", "package", ctx.Package, "args", ctx.Args)

	return nil
}

func (w *WASMEdgeSyscalls) prepareTmpFile(content string) (string, error) {
	fp, err := os.CreateTemp("", "*.go")
	if err != nil {
		slog.Error("Error creating temp file", "error", err)
		return "", err
	}
	defer fp.Close()

	if _, err := fp.Write([]byte(content)); err != nil {
		slog.Error("Error writing to temp file", "error", err)
		return "", err
	}

	return fp.Name(), nil
}

func (w *WASMEdgeSyscalls) Name() string {
	return "WASMEdgeSyscalls"
}
