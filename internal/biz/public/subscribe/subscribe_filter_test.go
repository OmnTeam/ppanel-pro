package subscribe

import "testing"

func TestFilterExperimentalNodesForClient(t *testing.T) {
	mkList := func() []*UserSubscribeInfo {
		return []*UserSubscribeInfo{
			{
				ID: 10,
				Nodes: []*UserSubscribeNodeInfo{
					{ID: 1, Protocol: "trojan"},
					{ID: 2, Protocol: "simnet"},
					{ID: 3, Protocol: "omniflow"},
					{ID: 4, Protocol: "vmess"},
				},
			},
		}
	}

	t.Run("official client omnxt keeps all nodes", func(t *testing.T) {
		list := mkList()
		FilterExperimentalNodesForClient(list, "omnxt/1.0")
		if len(list[0].Nodes) != 4 {
			t.Fatalf("expected 4 nodes for official client, got %d", len(list[0].Nodes))
		}
	})

	t.Run("official client slaglab keeps all nodes", func(t *testing.T) {
		list := mkList()
		FilterExperimentalNodesForClient(list, "SlaGLab/0.4")
		if len(list[0].Nodes) != 4 {
			t.Fatalf("expected 4 nodes for official client, got %d", len(list[0].Nodes))
		}
	})

	t.Run("official android slag keeps all nodes", func(t *testing.T) {
		list := mkList()
		FilterExperimentalNodesForClient(list, "Slag/1.0.0 (Android; Android W528JS release-keys; arm64) AnyiNet/2.0")
		if len(list[0].Nodes) != 4 {
			t.Fatalf("expected 4 nodes for official client, got %d", len(list[0].Nodes))
		}
	})

	t.Run("vmess node keeps mixed protocol metadata", func(t *testing.T) {
		list := []*UserSubscribeInfo{
			{
				ID: 10,
				Nodes: []*UserSubscribeNodeInfo{
					{
						ID:       4,
						Protocol: "vmess",
						Protocols: `[{"type":"simnet","enable":false},{"type":"omniflow","enable":false},` +
							`{"type":"vmess","enable":true,"port":1566}]`,
					},
				},
			},
		}
		FilterExperimentalNodesForClient(list, "clash-verge/1.0")
		if len(list[0].Nodes) != 1 {
			t.Fatalf("expected vmess node to survive mixed protocol metadata, got %d", len(list[0].Nodes))
		}
	})

	t.Run("third-party client hides experimental nodes", func(t *testing.T) {
		list := mkList()
		FilterExperimentalNodesForClient(list, "clash-verge/1.0")
		if len(list[0].Nodes) != 2 {
			t.Fatalf("expected 2 nodes after filtering, got %d", len(list[0].Nodes))
		}
		for _, n := range list[0].Nodes {
			if isExperimentalProtocol(n.Protocol) {
				t.Fatalf("experimental protocol %q leaked", n.Protocol)
			}
		}
	})

	t.Run("empty user agent hides experimental nodes", func(t *testing.T) {
		list := mkList()
		FilterExperimentalNodesForClient(list, "")
		if len(list[0].Nodes) != 2 {
			t.Fatalf("expected 2 nodes after filtering, got %d", len(list[0].Nodes))
		}
	})
}

func TestIsExperimentalProtocol(t *testing.T) {
	cases := []struct {
		protocol string
		want     bool
	}{
		{"simnet", true},
		{"Simnet", true},
		{"omniflow", true},
		{"OmniFlow", true},
		{"trojan", false},
		{"vmess", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isExperimentalProtocol(c.protocol); got != c.want {
			t.Fatalf("isExperimentalProtocol(%q) = %v, want %v", c.protocol, got, c.want)
		}
	}
}
