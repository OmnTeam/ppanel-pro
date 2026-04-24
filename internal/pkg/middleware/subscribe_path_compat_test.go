package middleware

import (
	"testing"

	"github.com/OmnTeam/ppanel-pro/internal/conf"
)

func TestShouldBlockSubscribePath(t *testing.T) {
	appConf := &conf.Application{
		Subscribe: &conf.Subscribe{
			SubscribePath: "/custom-subscribe",
		},
	}

	tests := []struct {
		path string
		want bool
	}{
		{path: "/api/subscribe", want: true},
		{path: "/api/subscribe/demo", want: true},
		{path: "/v1/subscribe/config", want: true},
		{path: "/v1/subscribe/demo", want: true},
		{path: "/custom-subscribe", want: false},
		{path: "/custom-subscribe/demo-token", want: false},
		{path: "/custom-subscribe/demo/token", want: false},
	}

	for _, tc := range tests {
		if got := shouldBlockSubscribePath(appConf, tc.path); got != tc.want {
			t.Fatalf("shouldBlockSubscribePath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
