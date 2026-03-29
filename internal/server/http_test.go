package server

import (
	"testing"

	"github.com/OmnTeam/ppanel-pro/internal/conf"
)

func TestSubscribeCompatPathsDefaultMatchesLegacy(t *testing.T) {
	paths := subscribeCompatPaths(nil)
	if len(paths) != 1 || paths[0] != "/v1/subscribe/config" {
		t.Fatalf("unexpected default subscribe compat paths: %#v", paths)
	}
}

func TestSubscribeCompatPathsHonorsConfiguredLegacyPath(t *testing.T) {
	paths := subscribeCompatPaths(&conf.Application{
		Subscribe: &conf.Subscribe{
			SubscribePath: "sub/custom",
		},
	})

	if len(paths) != 2 {
		t.Fatalf("unexpected path count: %#v", paths)
	}
	if paths[0] != "/v1/subscribe/config" || paths[1] != "/sub/custom" {
		t.Fatalf("unexpected subscribe compat paths: %#v", paths)
	}
}
