package system

import (
	"context"
	"io"
	"testing"

	"github.com/OmnTeam/npanel-pro/internal/buildmeta"
	"github.com/go-kratos/kratos/v2/log"
)

func TestGetSystemModulePrefersMainVersion(t *testing.T) {
	buildmeta.SetMainVersion("v2.3.4")
	t.Cleanup(func() {
		buildmeta.SetMainVersion("")
	})

	uc := NewSystemUsecase(nil, log.NewStdLogger(io.Discard))
	module, err := uc.GetSystemModule(context.Background())
	if err != nil {
		t.Fatalf("GetSystemModule() error = %v", err)
	}
	if module.ServiceVersion != "2.3.4" {
		t.Fatalf("ServiceVersion = %q, want %q", module.ServiceVersion, "2.3.4")
	}
}
