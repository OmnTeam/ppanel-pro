package subscription

import "testing"

func TestIsOfficialClient(t *testing.T) {
	cases := []struct {
		name      string
		userAgent string
		want      bool
	}{
		// 自有客户端/SDK（白名单）应放行
		{"omnxt exact lower", "omnxt/1.2.3", true},
		{"omnxt upper", "OMNXT 2.0", true},
		{"omnxt embedded", "Mozilla/5.0 omnxt-client/0.9", true},
		{"slaglab exact lower", "slaglab/0.4", true},
		{"slaglab upper", "SlaGLab/0.4", true},
		{"slag android", "Slag/1.0.0 (Android; Android W528JS release-keys; arm64) AnyiNet/2.0", true},

		// 非白名单应隐藏实验协议
		{"clash", "clash-verge/1.0", false},
		{"v2ray", "v2rayN/6.0", false},
		{"sing-box", "sing-box/1.8", false},
		{"browser", "Mozilla/5.0 (Macintosh) Chrome/120", false},
		{"curl", "curl/7.81.0", false},
		{"empty", "", false},
		{"whitespace", "   ", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isOfficialClient(c.userAgent); got != c.want {
				t.Fatalf("isOfficialClient(%q) = %v, want %v", c.userAgent, got, c.want)
			}
		})
	}
}

func TestFilterSubscriptionNodesByUserAgent(t *testing.T) {
	mkNodes := func() []*NodeInfo {
		return []*NodeInfo{
			{ID: 1, Type: "trojan"},
			{ID: 2, Type: "simnet"},
			{ID: 3, Type: "omniflow"},
			{ID: 4, Type: "vmess"},
		}
	}

	t.Run("official client omnxt keeps all nodes", func(t *testing.T) {
		got := filterSubscriptionNodesByUserAgent(mkNodes(), "omnxt/1.0")
		if len(got) != 4 {
			t.Fatalf("expected 4 nodes for official client, got %d", len(got))
		}
	})

	t.Run("official client slaglab keeps all nodes", func(t *testing.T) {
		got := filterSubscriptionNodesByUserAgent(mkNodes(), "slaglab/0.4")
		if len(got) != 4 {
			t.Fatalf("expected 4 nodes for official client, got %d", len(got))
		}
	})

	t.Run("third-party client hides experimental protocols", func(t *testing.T) {
		got := filterSubscriptionNodesByUserAgent(mkNodes(), "clash-verge/1.0")
		if len(got) != 2 {
			t.Fatalf("expected 2 nodes after filtering, got %d", len(got))
		}
		for _, n := range got {
			if isExperimentalSubscriptionProtocol(n.Type) {
				t.Fatalf("experimental protocol %q leaked to third-party client", n.Type)
			}
		}
	})

	t.Run("empty user agent hides experimental protocols", func(t *testing.T) {
		got := filterSubscriptionNodesByUserAgent(mkNodes(), "")
		if len(got) != 2 {
			t.Fatalf("expected 2 nodes after filtering, got %d", len(got))
		}
	})

	t.Run("unknown user agent hides experimental protocols", func(t *testing.T) {
		got := filterSubscriptionNodesByUserAgent(mkNodes(), "some-random-client/3.1")
		if len(got) != 2 {
			t.Fatalf("expected 2 nodes after filtering, got %d", len(got))
		}
	})

	t.Run("nil input returns nil", func(t *testing.T) {
		if got := filterSubscriptionNodesByUserAgent(nil, "clash"); got != nil {
			t.Fatalf("expected nil, got %v", got)
		}
	})
}
