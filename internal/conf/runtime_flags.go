package conf

import "sync/atomic"

var legacyDebugMode atomic.Bool

func SetLegacyDebugMode(enabled bool) {
	legacyDebugMode.Store(enabled)
}

func LegacyDebugMode() bool {
	return legacyDebugMode.Load()
}
