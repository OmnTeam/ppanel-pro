package middleware

import "testing"

func TestLegacyRouteGuardBlocksNewOnlyPaths(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{path: "/v1/admin/server/migrate/has", want: true},
		{path: "/v1/admin/server/migrate/run", want: true},
		{path: "/v1/admin/group/migrate", want: true},
		{path: "/v1/auth/check-telephone", want: true},
		{path: "/v1/payment/demo/alipay/notify", want: true},
		{path: "/v1/subscribe/demo", want: false},
		{path: "/v1/auth/check/telephone", want: false},
		{path: "/v1/subscribe/config", want: false},
	}

	for _, tc := range tests {
		if got := isBlockedLegacyExtraPath(tc.path); got != tc.want {
			t.Fatalf("isBlockedLegacyExtraPath(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}
