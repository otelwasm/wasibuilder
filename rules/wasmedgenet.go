package rules

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

//go:embed wasmedgenet/*
var fsForNet embed.FS

type WASMEdgeNet struct {
}

func (w *WASMEdgeNet) Apply(ctx *ExecContext) error {
	// We're primarily interested in the compile tool
	if filepath.Base(ctx.Command) != "compile" {
		return nil
	}

	var err error
	switch ctx.Package {
	case "syscall":
		err = w.processSyscallPackage(ctx)
	case "net":
		err = w.processNetPackage(ctx)
	case "net/http":
		err = w.processHTTPPackage(ctx)
	case "internal/poll":
		err = w.processPollPackage(ctx)
	default:
		return nil
	}

	if err != nil {
		return err
	}

	slog.Info("package is...", "package", ctx.Package, "args", ctx.Args)

	return nil
}

func (w *WASMEdgeNet) processSyscallPackage(ctx *ExecContext) error {
	var removedFiles = []string{
		"src/syscall/net_fake.go",
		"src/syscall/net_wasip1.go",
	}
	var addedFilesFromFS = []string{
		"syscall/net_fake.go.replaced",
		"syscall/net_wasip1.go.replaced",
	}

	_, err := w.removeFiles(ctx, removedFiles)
	if err != nil {
		slog.Error("Error removing files", "error", err)
		return err
	}

	err = w.addFilesFromFS(ctx, addedFilesFromFS)
	if err != nil {
		slog.Error("Error adding files from FS", "error", err)
		return err
	}

	return nil
}

func (w *WASMEdgeNet) processNetPackage(ctx *ExecContext) error {
	var removedFiles = []string{
		"src/net/net_fake.go",
		"src/net/fd_fake.go",
		"src/net/file_wasip1.go",
		"src/net/fd_wasip1.go",
		"src/net/sockopt_fake.go",
	}
	var addedFiles = []string{
		"src/net/fd_posix.go",
		"src/net/fd_unix.go",
		"src/net/sock_posix.go",
		"src/net/file_unix.go",
	}
	var addedFilesFromFS = []string{
		"net/fake.go.added",
		"net/sockopt_wasip1.go.added",
	}

	baseDir, err := w.removeFiles(ctx, removedFiles)
	if err != nil {
		slog.Error("Error removing files", "error", err)
		return err
	}

	err = w.addFilesFromLocal(ctx, baseDir, addedFiles)
	if err != nil {
		slog.Error("Error adding files from local", "error", err)
		return err
	}

	err = w.addFilesFromFS(ctx, addedFilesFromFS)
	if err != nil {
		slog.Error("Error adding files from FS", "error", err)
		return err
	}

	return nil
}

func (w *WASMEdgeNet) processHTTPPackage(ctx *ExecContext) error {
	var removedFiles = []string{
		"src/net/http/transport_default_wasm.go",
	}
	var addedFiles = []string{
		"src/net/http/transport_default_other.go",
	}

	baseDir, err := w.removeFiles(ctx, removedFiles)
	if err != nil {
		slog.Error("Error removing files", "error", err)
		return err
	}

	err = w.addFilesFromLocal(ctx, baseDir, addedFiles)
	if err != nil {
		slog.Error("Error adding files from local", "error", err)
		return err
	}

	return nil
}

func (w *WASMEdgeNet) processPollPackage(ctx *ExecContext) error {
	var addedFilesFromFS = []string{
		"poll/sockopt.go.added",
	}

	err := w.addFilesFromFS(ctx, addedFilesFromFS)
	if err != nil {
		slog.Error("Error adding files from FS", "error", err)
		return err
	}

	return nil
}

func (w *WASMEdgeNet) removeFiles(ctx *ExecContext, files []string) (baseDir string, err error) {
	for _, src := range files {
		for i, arg := range ctx.Args {
			b, found := strings.CutSuffix(arg, src)

			if !found {
				continue
			}

			// Remove the argument
			ctx.Args = append(ctx.Args[:i], ctx.Args[i+1:]...)
			slog.Info("Removed argument", "removed", arg)

			baseDir = b
		}
	}

	return baseDir, nil
}

func (w *WASMEdgeNet) addFilesFromLocal(ctx *ExecContext, baseDir string, files []string) error {
	for _, src := range files {
		addedPath := filepath.Join(baseDir, src)

		ctx.Args = append(ctx.Args, addedPath)
		slog.Info("Added argument", "added", addedPath)
	}

	return nil
}

func (w *WASMEdgeNet) addFilesFromFS(ctx *ExecContext, files []string) error {
	for _, src := range files {
		// Read the file from the embedded filesystem
		content, err := fsForNet.ReadFile("wasmedgenet/" + src)
		if err != nil {
			slog.Error("Error reading embedded file", "file", src, "error", err)
			return err
		}
		// Prepare a temporary file with the content
		tmpFile, err := w.prepareTmpFile(string(content))
		if err != nil {
			slog.Error("Error preparing temp file", "file", src, "error", err)
			return err
		}
		ctx.Args = append(ctx.Args, tmpFile)
		slog.Info("Added argument", "added", tmpFile)
	}

	return nil
}

func (w *WASMEdgeNet) prepareTmpFile(content string) (string, error) {
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

func (w *WASMEdgeNet) Name() string {
	return "WASMEdgeNet"
}
