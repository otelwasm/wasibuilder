package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/tsuzu/wasibuilder/rules"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s go [args...]\n", os.Args[0])
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help", "--help", "-h":
		fmt.Println("wasibuilder: A tool to modify Go build arguments for WASI.")
		fmt.Println("Usage: wasibuilder go build ...")
		os.Exit(1)
	case "toolexec":
		toolexec()
	default:
		passthrough()
	}
}

func passthrough() {
	command := os.Args[1]
	if len(os.Args) < 3 {
		execCommands(command, nil, nil)
		return
	}

	replaceEnviron, err := replaceEnviron(command, os.Environ())
	if err != nil {
		slog.Error("Error replacing environment variables", "error", err)
		os.Exit(1)
	}

	args := os.Args[2:]

	subcommandIndex := -1
	for i, cmd := range args {
		if !strings.HasPrefix(cmd, "-") {
			subcommandIndex = i
			break
		}
	}
	if subcommandIndex == -1 {
		execCommands(command, args, replaceEnviron)
		return
	}

	switch args[subcommandIndex] {
	case "build", "install", "run", "test":
		// These commands are passed through to the original tool
	default:
		execCommands(command, args, replaceEnviron)
		return
	}

	args = append(
		append(
			slices.Clone(args[:subcommandIndex+1]),
			[]string{"-toolexec", os.Args[0] + " toolexec"}...,
		),
		args[subcommandIndex+1:]...,
	)

	execCommands(command, args, replaceEnviron)
}

func replaceEnviron(command string, environ []string) ([]string, error) {
	for i, env := range environ {
		if strings.HasPrefix(env, "GOOS=") {
			environ = append(environ[:i], environ[i+1:]...)
		}
		if strings.HasPrefix(env, "GOARCH=") {
			environ = append(environ[:i], environ[i+1:]...)
		}
		if strings.HasPrefix(env, "GOCACHE=") {
			environ = append(environ[:i], environ[i+1:]...)
		}
	}
	environ = append(
		environ,
		"GOOS=wasip1",
		"GOARCH=wasm",
	)

	b, err := exec.Command(command, "env").CombinedOutput()
	if err != nil {
		slog.Error("Error executing command", "command", command, "error", err)
		return nil, err
	}

	gocacheDir := ""
	for _, line := range strings.Split(string(b), "\n") {
		cut, ok := strings.CutPrefix(line, "GOCACHE='")
		if !ok {
			continue
		}

		gocacheDir, ok = strings.CutSuffix(cut, "'")
		if !ok {
			gocacheDir = ""
			continue
		}
		break
	}
	if gocacheDir != "" {
		environ = append(environ, "GOCACHE="+gocacheDir+"-wasibuilder")
	}

	return environ, nil
}

func toolexec() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s toolexec [tool] [args...]\n", os.Args[0])
		os.Exit(1)
	}

	// The first argument is the compiler tool to run (like compile, link, etc.)
	toolName := os.Args[2]
	args := os.Args[3:]

	// Check if we need to modify args based on the package being compiled
	args = modifyArgsIfNeeded(toolName, args)

	execCommands(toolName, args, nil)
}

func execCommands(toolName string, args []string, environ []string) {
	// Execute the original tool with possibly modified arguments
	cmd := exec.Command(toolName, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if environ != nil {
		cmd.Env = environ
	} else {
		cmd.Env = os.Environ()
	}

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
