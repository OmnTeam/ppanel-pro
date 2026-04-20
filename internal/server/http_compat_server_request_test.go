package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCompatLegacyPopulateV1ServerCommonFallsBackToGETBodyForm(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/v1/server/user", strings.NewReader("server_id=12&protocol=vless&secret_key=test-secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var common compatLegacyServerCommon
	compatLegacyPopulateV1ServerCommon(req, &common)

	if common.ServerID != 12 {
		t.Fatalf("ServerID = %d, want 12", common.ServerID)
	}
	if common.Protocol != "vless" {
		t.Fatalf("Protocol = %q, want %q", common.Protocol, "vless")
	}
	if common.SecretKey != "test-secret" {
		t.Fatalf("SecretKey = %q, want %q", common.SecretKey, "test-secret")
	}
}

func TestCompatLegacyPopulateV1ServerCommonKeepsExistingValues(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/v1/server/user?server_id=34&protocol=trojan&secret_key=query-secret", strings.NewReader("server_id=56&protocol=vless&secret_key=body-secret"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	common := compatLegacyServerCommon{
		ServerID:  78,
		Protocol:  "vmess",
		SecretKey: "bound-secret",
	}
	compatLegacyPopulateV1ServerCommon(req, &common)

	if common.ServerID != 78 {
		t.Fatalf("ServerID = %d, want 78", common.ServerID)
	}
	if common.Protocol != "vmess" {
		t.Fatalf("Protocol = %q, want %q", common.Protocol, "vmess")
	}
	if common.SecretKey != "bound-secret" {
		t.Fatalf("SecretKey = %q, want %q", common.SecretKey, "bound-secret")
	}
}
