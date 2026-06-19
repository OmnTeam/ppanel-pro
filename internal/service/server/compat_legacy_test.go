package server

import (
	"context"
	"testing"

	"github.com/OmnTeam/npanel-pro/ent"
	"github.com/OmnTeam/npanel-pro/ent/enttest"
	"github.com/OmnTeam/npanel-pro/internal/conf"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	_ "github.com/mattn/go-sqlite3"
)

type compatTestProvider struct {
	db *ent.Client
}

func (p compatTestProvider) DB() *ent.Client {
	return p.db
}

func (p compatTestProvider) Redis() redis.UniversalClient {
	return nil
}

func (p compatTestProvider) Queue() *asynq.Client {
	return nil
}

func (p compatTestProvider) AppNodeConfig() *conf.Node {
	return nil
}

func (p compatTestProvider) LoadNodeConfig(ctx context.Context, module string) (*CompatLegacyNodeConfig, error) {
	return &CompatLegacyNodeConfig{
		NodeSecret:             "secret",
		NodePullInterval:       10,
		NodePushInterval:       60,
		TrafficReportThreshold: 0,
		IPStrategy:             "prefer_ipv4",
	}, nil
}

func TestCompatQueryServerProtocolConfigFiltersEnabledProtocols(t *testing.T) {
	ctx := context.Background()
	client := enttest.Open(t, "sqlite3", "file:compat_query_server_protocols?mode=memory&cache=shared&_fk=1")
	defer client.Close()

	client.ProxyServer.Create().
		SetID(2).
		SetName("HK").
		SetServerAddr("103.214.22.2").
		SetProtocol(`[{"type":"simnet","port":443,"enable":false,"ratio":1,"cert_mode":"none"},{"type":"omniflow","port":443,"enable":false,"security":"tls","ratio":1},{"type":"vmess","port":1566,"enable":true,"security":"none","fingerprint":"chrome","transport":"tcp","ratio":1,"cert_mode":"none"},{"type":"vless","port":1886,"enable":true,"security":"none","fingerprint":"chrome","transport":"tcp","flow":"none","xhttp_mode":"auto","encryption":"none","ratio":1,"cert_mode":"none"}]`).
		SaveX(ctx)

	service := &ServerService{}
	resp, err := service.CompatQueryServerProtocolConfig(ctx, compatTestProvider{db: client}, &CompatLegacyQueryServerConfigRequest{
		ServerID:  2,
		Protocols: []string{"vmess", "vless"},
	})
	if err != nil {
		t.Fatalf("CompatQueryServerProtocolConfig returned error: %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("Total = %d, want 2", resp.Total)
	}
	if len(resp.Protocols) != 2 {
		t.Fatalf("Protocols length = %d, want 2", len(resp.Protocols))
	}
	if resp.Protocols[0].Type != "vmess" || resp.Protocols[0].Port != 1566 {
		t.Fatalf("first protocol = %s:%d, want vmess:1566", resp.Protocols[0].Type, resp.Protocols[0].Port)
	}
	if resp.Protocols[1].Type != "vless" || resp.Protocols[1].Port != 1886 {
		t.Fatalf("second protocol = %s:%d, want vless:1886", resp.Protocols[1].Type, resp.Protocols[1].Port)
	}
	if resp.IPStrategy != "prefer_ipv4" {
		t.Fatalf("IPStrategy = %q, want prefer_ipv4", resp.IPStrategy)
	}
}
