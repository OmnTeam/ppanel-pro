package buildmeta

import "testing"

func TestVersionPrefersMainVersion(t *testing.T) {
	SetMainVersion(" v2.3.4 ")
	t.Cleanup(func() {
		SetMainVersion("")
	})

	if got := Version(); got != "v2.3.4" {
		t.Fatalf("Version() = %q, want %q", got, "v2.3.4")
	}
}
