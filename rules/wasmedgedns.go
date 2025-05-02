package rules

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"
)

//go:embed wasmedgedns/*
var fsForDNS embed.FS

type WASMEdgeDNS struct {
}

func (w *WASMEdgeDNS) Apply(ctx *ExecContext) error {
	logger := slog.With("rule", w.Name())

	// We're primarily interested in the compile tool
	if filepath.Base(ctx.Command) != "compile" {
		return nil
	}

	var err error
	switch ctx.Package {
	case "net":
		err = w.processNetPackage(ctx, logger)
	default:
		return nil
	}

	if err != nil {
		return err
	}

	return nil
}
func (w *WASMEdgeDNS) processNetPackage(ctx *ExecContext, logger *slog.Logger) error {
	var addedFilesFromFS = []string{
		"getaddrinfo_wasip1.go.added",
		"sock_getaddrinfo_wasip1.go.added",
		"getaddrinfo_hook_wasip1.go.added",
	}

	err := w.addFilesFromFS(ctx, addedFilesFromFS, logger)
	if err != nil {
		return err
	}

	return nil
}

func (w *WASMEdgeDNS) addFilesFromFS(ctx *ExecContext, files []string, logger *slog.Logger) error {
	for _, src := range files {
		// Read the file from the embedded filesystem
		content, err := fsForDNS.ReadFile("wasmedgedns/" + src)
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

func (w *WASMEdgeDNS) prepareTmpFile(content string, logger *slog.Logger) (string, error) {
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

func (w *WASMEdgeDNS) Name() string {
	return "WASMEdgeDNS"
}
