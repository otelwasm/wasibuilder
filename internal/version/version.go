package version

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

var goVersionRegex = regexp.MustCompile(`go(\d+)\.(\d+)\.(\d+)`)

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) Compare(other Version) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	return v.Patch - other.Patch
}

// IsAtLeast returns true if this version is >= the specified version
func (v Version) IsAtLeast(major, minor, patch int) bool {
	return v.Compare(Version{Major: major, Minor: minor, Patch: patch}) >= 0
}

// GetGoVersion gets the Go version from the GOVERSION environment variable
// This is the most reliable method as it's automatically set by the Go runtime
func GetGoVersion() (Version, error) {
	goVersion := os.Getenv("GOVERSION")
	if goVersion == "" {
		return Version{}, fmt.Errorf("GOVERSION environment variable not set")
	}

	matches := goVersionRegex.FindStringSubmatch(goVersion)
	if len(matches) != 4 {
		return Version{}, fmt.Errorf("failed to parse GOVERSION: %s", goVersion)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return Version{}, fmt.Errorf("failed to parse major version from GOVERSION: %s", goVersion)
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return Version{}, fmt.Errorf("failed to parse minor version from GOVERSION: %s", goVersion)
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return Version{}, fmt.Errorf("failed to parse patch version from GOVERSION: %s", goVersion)
	}

	return Version{Major: major, Minor: minor, Patch: patch}, nil
}
