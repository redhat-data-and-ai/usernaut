package config

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var versionFile string

// Version returns the SDK version.
func Version() string {
	return strings.TrimSpace(versionFile)
}
