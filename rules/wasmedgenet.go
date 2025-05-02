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
	logger := slog.With("rule", w.Name())

	// We're primarily interested in the compile tool
	if filepath.Base(ctx.Command) != "compile" {
		return nil
	}

	var err error
	switch ctx.Package {
	case "syscall":
		err = w.processSyscallPackage(ctx, logger)
	case "net":
		err = w.processNetPackage(ctx, logger)
	case "net/http":
		err = w.processHTTPPackage(ctx, logger)
	case "internal/poll":
		err = w.processPollPackage(ctx, logger)
	default:
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}

func (w *WASMEdgeNet) processSyscallPackage(ctx *ExecContext, logger *slog.Logger) error {
	var removedFiles = []string{
		"src/syscall/net_fake.go",
		"src/syscall/net_wasip1.go",
	}
	var addedFilesFromFS = []string{
		"syscall/net_fake.go.replaced",
		"syscall/net_wasip1.go.replaced",
	}

	_, err := w.removeFiles(ctx, removedFiles, logger)
	if err != nil {
		logger.Error("Error removing files", "error", err)
		return err
	}

	err = w.addFilesFromFS(ctx, addedFilesFromFS, logger)
	if err != nil {
		logger.Error("Error adding files from FS", "error", err)
		return err
	}

	return nil
}

func (w *WASMEdgeNet) processNetPackage(ctx *ExecContext, logger *slog.Logger) error {
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

	baseDir, err := w.removeFiles(ctx, removedFiles, logger)
	if err != nil {
		logger.Error("Error removing files", "error", err)
		return err
	}

	err = w.addFilesFromLocal(ctx, baseDir, addedFiles, logger)
	if err != nil {
		logger.Error("Error adding files from local", "error", err)
		return err
	}

	err = w.addFilesFromFS(ctx, addedFilesFromFS, logger)
	if err != nil {
		logger.Error("Error adding files from FS", "error", err)
		return err
	}

	return nil
}

func (w *WASMEdgeNet) processHTTPPackage(ctx *ExecContext, logger *slog.Logger) error {
	var removedFiles = []string{
		"src/net/http/transport_default_wasm.go",
	}
	var addedFiles = []string{
		"src/net/http/transport_default_other.go",
	}

	baseDir, err := w.removeFiles(ctx, removedFiles, logger)
	if err != nil {
		logger.Error("Error removing files", "error", err)
		return err
	}

	err = w.addFilesFromLocal(ctx, baseDir, addedFiles, logger)
	if err != nil {
		logger.Error("Error adding files from local", "error", err)
		return err
	}

	return nil
}

func (w *WASMEdgeNet) processPollPackage(ctx *ExecContext, logger *slog.Logger) error {
	var addedFilesFromFS = []string{
		"poll/sockopt.go.added",
	}

	err := w.addFilesFromFS(ctx, addedFilesFromFS, logger)
	if err != nil {
		logger.Error("Error adding files from FS", "error", err)
		return err
	}

	return nil
}

func (w *WASMEdgeNet) removeFiles(ctx *ExecContext, files []string, logger *slog.Logger) (baseDir string, err error) {
	for _, src := range files {
		for i, arg := range ctx.Args {
			b, found := strings.CutSuffix(arg, src)

			if !found {
				continue
			}

			// Remove the argument
			ctx.Args = append(ctx.Args[:i], ctx.Args[i+1:]...)
			logger.Info("Removed argument", "removed", arg)

			baseDir = b
		}
	}

	return baseDir, nil
}

func (w *WASMEdgeNet) addFilesFromLocal(ctx *ExecContext, baseDir string, files []string, logger *slog.Logger) error {
	for _, src := range files {
		addedPath := filepath.Join(baseDir, src)

		ctx.Args = append(ctx.Args, addedPath)
		logger.Debug("Added argument", "added", addedPath)
	}

	return nil
}

func (w *WASMEdgeNet) addFilesFromFS(ctx *ExecContext, files []string, logger *slog.Logger) error {
	for _, src := range files {
		// Read the file from the embedded filesystem
		content, err := fsForNet.ReadFile("wasmedgenet/" + src)
		if err != nil {
			logger.Error("Error reading embedded file", "file", src, "error", err)
			return err
		}
		// Prepare a temporary file with the content
		tmpFile, err := w.prepareTmpFile(string(content), logger)
		if err != nil {
			logger.Error("Error preparing temp file", "file", src, "error", err)
			return err
		}
		ctx.Args = append(ctx.Args, tmpFile)
		logger.Debug("Added argument", "added", tmpFile)
	}

	return nil
}

func (w *WASMEdgeNet) prepareTmpFile(content string, logger *slog.Logger) (string, error) {
	fp, err := os.CreateTemp("", "*.go")
	if err != nil {
		logger.Error("Error creating temp file", "error", err)
		return "", err
	}
	defer fp.Close()

	if _, err := fp.Write([]byte(content)); err != nil {
		logger.Error("Error writing to temp file", "error", err)
		return "", err
	}

	return fp.Name(), nil
}

func (w *WASMEdgeNet) Name() string {
	return "WASMEdgeNet"
}
