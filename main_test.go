package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAll(t *testing.T) {
	err := exec.Command("go", "build", "-o", "bin/wasibuilder", ".").Run()
	if err != nil {
		t.Fatalf("failed to build wasibuilder: %v", err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	wasibuilderPath := filepath.Join(wd, "bin/wasibuilder")

	cmd := exec.Command("go", "build", "-a", "-o", "bin/httpget.wasm", "-toolexec", wasibuilderPath, "testdata/httpget/main.go")
	cmd.Env = append(
		os.Environ(),
		"GOOS=wasip1",
		"GOARCH=wasm",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run command: %v", err)
	}
}
