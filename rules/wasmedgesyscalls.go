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

type WASMEdgeSyscalls struct {
}

func (w *WASMEdgeSyscalls) Apply(ctx *ExecContext) error {
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

func (w *WASMEdgeSyscalls) processSyscallPackage(ctx *ExecContext) error {
	var replacedFiles = map[string]string{
		"src/syscall/net_fake.go":   "syscall/net_fake.go.replaced",
		"src/syscall/net_wasip1.go": "syscall/net_wasip1.go.replaced",
	}

	for src, dst := range replacedFiles {
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
	return nil
}

func (w *WASMEdgeSyscalls) processNetPackage(ctx *ExecContext) error {
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

	var baseDir string
	for _, src := range removedFiles {
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

	for _, src := range addedFiles {
		addedPath := filepath.Join(baseDir, src)

		ctx.Args = append(ctx.Args, addedPath)
		slog.Info("Added argument", "added", addedPath)
	}

	for _, src := range addedFilesFromFS {
		// Read the file from the embedded filesystem
		content, err := fs.ReadFile("wasmedgesyscalls/" + src)
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

func (w *WASMEdgeSyscalls) processPollPackage(ctx *ExecContext) error {
	var addedFilesFromFS = []string{
		"poll/sockopt.go.added",
	}

	for _, src := range addedFilesFromFS {
		// Read the file from the embedded filesystem
		content, err := fs.ReadFile("wasmedgesyscalls/" + src)
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
