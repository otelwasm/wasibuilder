package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/tsuzu/wasibuilder/rules"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s command [args...]\n", os.Args[0])
		os.Exit(1)
	}

	// The first argument is the compiler tool to run (like compile, link, etc.)
	toolName := os.Args[1]
	args := os.Args[2:]

	// Check if we need to modify args based on the package being compiled
	args = modifyArgsIfNeeded(toolName, args)

	// Execute the original tool with possibly modified arguments
	cmd := exec.Command(toolName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		slog.Error("Error executing command", "command", toolName, "args", args, "error", err)
		os.Exit(1)
	}
}

var allRules = []rules.Rule{
	&rules.WASMEdgeSyscalls{},
}

// modifyArgsIfNeeded checks if the current compilation is for a package we want to modify
// and returns potentially modified arguments
func modifyArgsIfNeeded(toolName string, args []string) []string {
	// Find the package path in the arguments
	var packagePath string
	packageIndex := -1
	for i, arg := range args {
		if arg == "-p" && i+1 < len(args) {
			packagePath = args[i+1]
			packageIndex = i + 1
			break
		}
	}

	eCtx := rules.ExecContext{
		Command:      toolName,
		Args:         args,
		Package:      packagePath,
		PackageIndex: packageIndex,
	}

	for _, rule := range allRules {
		if err := rule.Apply(&eCtx); err != nil {
			slog.Error("Error applying rule", "rule", rule.Name(), "error", err)
			return args
		}
	}

	return eCtx.Args
}
