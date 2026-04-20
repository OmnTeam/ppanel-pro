package server

import (
	"encoding/json"
	"testing"
)

func TestCompatLegacyV1PostSuccessEnvelopeMatchesLegacyJSON(t *testing.T) {
	t.Parallel()

	got, err := json.Marshal(compatEnvelope{
		Code: 200,
		Msg:  "success",
	})
	if err != nil {
		t.Fatalf("marshal success envelope failed: %v", err)
	}

	const want = `{"code":200,"msg":"success"}`
	if string(got) != want {
		t.Fatalf("success envelope mismatch\nwant: %s\ngot:  %s", want, got)
	}
}

func TestCompatLegacyV1PostErrorEnvelopeMatchesLegacyJSON(t *testing.T) {
	t.Parallel()

	got, err := json.Marshal(compatEnvelope{
		Code: 500,
		Msg:  "Internal Server Error",
	})
	if err != nil {
		t.Fatalf("marshal error envelope failed: %v", err)
	}

	const want = `{"code":500,"msg":"Internal Server Error"}`
	if string(got) != want {
		t.Fatalf("error envelope mismatch\nwant: %s\ngot:  %s", want, got)
	}
}

func TestCompatLegacyV1GetServerConfigResponseMatchesLegacyJSON(t *testing.T) {
	t.Parallel()

	allowInsecure := true
	got, err := json.Marshal(compatLegacyGetServerConfigResponse{
		Basic: compatLegacyServerBasic{
			PushInterval: 60,
			PullInterval: 10,
		},
		Protocol: "vless",
		Config: compatLegacyVlessNode{
			Port:    443,
			Flow:    "xtls-rprx-vision",
			Network: "tcp",
			TransportConfig: &compatLegacyTransportConfig{
				Path:                 "/",
				Host:                 "example.com",
				ServiceName:          "grpc",
				DisableSNI:           true,
				ReduceRtt:            true,
				UDPRelayMode:         "native",
				CongestionController: "bbr",
			},
			Security: "reality",
			SecurityConfig: &compatLegacySecurityConfig{
				SNI:                  "example.com",
				AllowInsecure:        &allowInsecure,
				Fingerprint:          "chrome",
				RealityServerAddress: "1.1.1.1",
				RealityServerPort:    443,
				RealityPrivateKey:    "private",
				RealityPublicKey:     "public",
				RealityShortID:       "abcd",
				RealityMldsa65Seed:   "seed",
				PaddingScheme:        "",
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal config response failed: %v", err)
	}

	const want = `{"basic":{"push_interval":60,"pull_interval":10},"protocol":"vless","config":{"port":443,"flow":"xtls-rprx-vision","transport":"tcp","transport_config":{"path":"/","host":"example.com","service_name":"grpc","disable_sni":true,"reduce_rtt":true,"udp_relay_mode":"native","congestion_controller":"bbr"},"security":"reality","security_config":{"sni":"example.com","allow_insecure":true,"fingerprint":"chrome","reality_server_addr":"1.1.1.1","reality_server_port":443,"reality_private_key":"private","reality_public_key":"public","reality_short_id":"abcd","reality_mldsa65seed":"seed","padding_scheme":""}}}`
	if string(got) != want {
		t.Fatalf("config response mismatch\nwant: %s\ngot:  %s", want, got)
	}
}

func TestCompatLegacyV1GetServerUserListResponseMatchesLegacyJSON(t *testing.T) {
	t.Parallel()

	got, err := json.Marshal(compatLegacyGetServerUserListResponse{
		Users: []compatLegacyServerUser{
			{
				ID:          1,
				UUID:        "uuid-1",
				SpeedLimit:  100,
				DeviceLimit: 2,
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal user list response failed: %v", err)
	}

	const want = `{"users":[{"id":1,"uuid":"uuid-1","speed_limit":100,"device_limit":2}]}`
	if string(got) != want {
		t.Fatalf("user list response mismatch\nwant: %s\ngot:  %s", want, got)
	}
}
