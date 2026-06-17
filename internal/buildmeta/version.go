package buildmeta

import (
	"runtime/debug"
	"strings"
	"sync/atomic"
)

var mainVersion atomic.Value

// SetMainVersion records the version injected into package main at process start.
func SetMainVersion(version string) {
	mainVersion.Store(strings.TrimSpace(version))
}

// Version returns the preferred service version.
func Version() string {
	if version := MainVersion(); isKnownVersion(version) {
		return version
	}
	if version := BuildInfoMainVersion(); isKnownVersion(version) {
		return version
	}
	return "unknown version"
}

// MainVersion returns the version injected through main.Version.
func MainVersion() string {
	value := mainVersion.Load()
	if value == nil {
		return ""
	}
	version, _ := value.(string)
	return strings.TrimSpace(version)
}

// BuildInfoMainVersion returns the version embedded by the Go toolchain.
func BuildInfoMainVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	return strings.TrimSpace(buildInfo.Main.Version)
}

func isKnownVersion(version string) bool {
	version = strings.TrimSpace(version)
	return version != "" && version != "(devel)" && version != "unknown" && version != "unknown version"
}
