package rules

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
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

// GetGoVersion gets the Go version from the GOVERSION environment variable
// This is the most reliable method as it's automatically set by the Go runtime
func GetGoVersion() (GoVersion, error) {
	goVersion := os.Getenv("GOVERSION")
	if goVersion == "" {
		return GoVersion{}, fmt.Errorf("GOVERSION environment variable not set")
	}

	// Extract version using precompiled regex
	matches := goVersionRegex.FindStringSubmatch(goVersion)

	if len(matches) != 4 {
		return GoVersion{}, fmt.Errorf("failed to parse GOVERSION: %s", goVersion)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return GoVersion{}, fmt.Errorf("invalid major version in GOVERSION: %s", goVersion)
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return GoVersion{}, fmt.Errorf("invalid minor version in GOVERSION: %s", goVersion)
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return GoVersion{}, fmt.Errorf("invalid patch version in GOVERSION: %s", goVersion)
	}

	return GoVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}
