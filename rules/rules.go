package rules

import (
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var goVersionRegex = regexp.MustCompile(`go(\d+)\.(\d+)\.(\d+)`)

type ExecContext struct {
	Command      string
	Args         []string
	Package      string
	PackageIndex int
	GoVersion    *GoVersion
}

func (ctx *ExecContext) Clone() *ExecContext {
	return &ExecContext{
		Command:      ctx.Command,
		Args:         slices.Clone(ctx.Args),
		Package:      ctx.Package,
		PackageIndex: ctx.PackageIndex,
		GoVersion:    ctx.GoVersion,
	}
}

type Rule interface {
	Apply(ctx *ExecContext) error
	Name() string
}

// GoVersion represents a Go version with major, minor, and patch components
type GoVersion struct {
	Major int
	Minor int
	Patch int
}

// Compare returns -1 if v < other, 0 if v == other, 1 if v > other
func (v GoVersion) Compare(other GoVersion) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// IsAtLeast returns true if this version is >= the specified version
func (v GoVersion) IsAtLeast(major, minor, patch int) bool {
	other := GoVersion{Major: major, Minor: minor, Patch: patch}
	return v.Compare(other) >= 0
}

// DetectGoVersion detects the Go version being used by the toolchain
func DetectGoVersion(goCommand string) (GoVersion, error) {
	cmd := exec.Command(goCommand, "version")
	output, err := cmd.Output()
	if err != nil {
		return GoVersion{}, err
	}

	// Parse output like "go version go1.25.1 darwin/arm64"
	versionStr := strings.TrimSpace(string(output))

	// Extract version using precompiled regex
	matches := goVersionRegex.FindStringSubmatch(versionStr)

	if len(matches) != 4 {
		return GoVersion{}, err
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return GoVersion{}, err
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return GoVersion{}, err
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return GoVersion{}, err
	}

	return GoVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// DetectGoVersionFromToolchain detects the Go version directly from the toolchain binary
// It uses the tool's own version flag (e.g., compile -V) to get version information
func DetectGoVersionFromToolchain(toolPath string) (GoVersion, error) {
	// Try to get version directly from the tool using -V flag
	cmd := exec.Command(toolPath, "-V")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to detecting from "go" command in PATH
		return DetectGoVersion("go")
	}

	// Parse output like "compile version go1.25.1"
	versionStr := strings.TrimSpace(string(output))

	// Extract version using precompiled regex
	matches := goVersionRegex.FindStringSubmatch(versionStr)

	if len(matches) != 4 {
		// Fallback to detecting from "go" command in PATH
		return DetectGoVersion("go")
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return GoVersion{}, err
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return GoVersion{}, err
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return GoVersion{}, err
	}

	return GoVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}
