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

	cmd := exec.Command(wasibuilderPath, "go", "build", "-o", "bin/httpget.wasm", "testdata/httpget/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run command: %v", err)
	}

	b, err := exec.Command("wasirun", "bin/httpget.wasm").CombinedOutput()
	t.Log(string(b))
	if err != nil {
		t.Fatalf("failed to run command: %v", err)
	}
}
