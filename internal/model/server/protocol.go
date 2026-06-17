package server

import (
	"encoding/json"
	"strings"
)

// Protocol represents a server protocol configuration - 完全按照原项目定义
type Protocol struct {
	Type                          string  `json:"type"`
	Port                          int32   `json:"port"`
	Enable                        bool    `json:"enable"`
	Security                      string  `json:"security,omitempty"`
	SNI                           string  `json:"sni,omitempty"`
	AllowInsecure                 bool    `json:"allow_insecure,omitempty"`
	Fingerprint                   string  `json:"fingerprint,omitempty"`
	RealityServerAddr             string  `json:"reality_server_addr,omitempty"`
	RealityServerPort             int32   `json:"reality_server_port,omitempty"`
	RealityPrivateKey             string  `json:"reality_private_key,omitempty"`
	RealityPublicKey              string  `json:"reality_public_key,omitempty"`
	RealityShortId                string  `json:"reality_short_id,omitempty"`
	Transport                     string  `json:"transport,omitempty"`
	Host                          string  `json:"host,omitempty"`
	Path                          string  `json:"path,omitempty"`
	ServiceName                   string  `json:"service_name,omitempty"`
	Cipher                        string  `json:"cipher,omitempty"`
	ServerKey                     string  `json:"server_key,omitempty"`
	Flow                          string  `json:"flow,omitempty"`
	HopPorts                      string  `json:"hop_ports,omitempty"`
	HopInterval                   int32   `json:"hop_interval,omitempty"`
	ObfsPassword                  string  `json:"obfs_password,omitempty"`
	DisableSNI                    bool    `json:"disable_sni,omitempty"`
	ReduceRtt                     bool    `json:"reduce_rtt,omitempty"`
	UDPRelayMode                  string  `json:"udp_relay_mode,omitempty"`
	CongestionController          string  `json:"congestion_controller,omitempty"`
	Multiplex                     string  `json:"multiplex,omitempty"`
	PaddingScheme                 string  `json:"padding_scheme,omitempty"`
	UpMbps                        int32   `json:"up_mbps,omitempty"`
	DownMbps                      int32   `json:"down_mbps,omitempty"`
	Obfs                          string  `json:"obfs,omitempty"`
	ObfsHost                      string  `json:"obfs_host,omitempty"`
	ObfsPath                      string  `json:"obfs_path,omitempty"`
	XhttpMode                     string  `json:"xhttp_mode,omitempty"`
	XhttpExtra                    string  `json:"xhttp_extra,omitempty"`
	Encryption                    string  `json:"encryption,omitempty"`
	EncryptionMode                string  `json:"encryption_mode,omitempty"`
	EncryptionRtt                 string  `json:"encryption_rtt,omitempty"`
	EncryptionTicket              string  `json:"encryption_ticket,omitempty"`
	EncryptionServerPadding       string  `json:"encryption_server_padding,omitempty"`
	EncryptionPrivateKey          string  `json:"encryption_private_key,omitempty"`
	EncryptionClientPadding       string  `json:"encryption_client_padding,omitempty"`
	EncryptionPassword            string  `json:"encryption_password,omitempty"`
	Ratio                         float64 `json:"ratio,omitempty"`
	CertMode                      string  `json:"cert_mode,omitempty"`
	CertDNSProvider               string  `json:"cert_dns_provider,omitempty"`
	CertDNSEnv                    string  `json:"cert_dns_env,omitempty"`
	SimnetPsk                     string  `json:"simnet_psk,omitempty"`
	SimnetKeyID                   int32   `json:"simnet_key_id,omitempty"`
	SimnetTicketID                string  `json:"simnet_ticket_id,omitempty"`
	SimnetPath                    string  `json:"simnet_path,omitempty"`
	SimnetCarrier                 string  `json:"simnet_carrier,omitempty"`
	SimnetAfEnabled               bool    `json:"simnet_af_enabled,omitempty"`
	SimnetAfPathMode              string  `json:"simnet_af_path_mode,omitempty"`
	SimnetAfPathPrefix            string  `json:"simnet_af_path_prefix,omitempty"`
	SimnetAfPathSuffix            string  `json:"simnet_af_path_suffix,omitempty"`
	SimnetAfMagicMode             string  `json:"simnet_af_magic_mode,omitempty"`
	SimnetAfResponseJitterMs      int32   `json:"simnet_af_response_jitter_ms,omitempty"`
	SimnetAfHandshakePolymorphism bool    `json:"simnet_af_handshake_polymorphism,omitempty"`
	SimnetAfSettingsJitter        bool    `json:"simnet_af_settings_jitter,omitempty"`
	SimnetAfFakeHeaderInjection   bool    `json:"simnet_af_fake_header_injection,omitempty"`
	SimnetReverseEnabled          bool    `json:"simnet_reverse_enabled,omitempty"`
	SimnetReverseListenAddr       string  `json:"simnet_reverse_listen_addr,omitempty"`
	SimnetReverseListenPort       int32   `json:"simnet_reverse_listen_port,omitempty"`
	SimnetReverseTargetHost       string  `json:"simnet_reverse_target_host,omitempty"`
	SimnetReverseTargetPort       int32   `json:"simnet_reverse_target_port,omitempty"`
	SimnetFallbackEnabled         bool    `json:"simnet_fallback_enabled,omitempty"`
	SimnetFallbackTargetScheme    string  `json:"simnet_fallback_target_scheme,omitempty"`
	SimnetFallbackTargetHost      string  `json:"simnet_fallback_target_host,omitempty"`
	SimnetFallbackTargetPort      int32   `json:"simnet_fallback_target_port,omitempty"`
	SimnetFallbackHostHeader      string  `json:"simnet_fallback_host_header,omitempty"`
	SimnetFallbackTLSSNI          string  `json:"simnet_fallback_tls_sni,omitempty"`

	// OmniFlow 基础配置
	OmniflowCarrier     string `json:"omniflow_carrier,omitempty"`
	OmniflowPath        string `json:"omniflow_path,omitempty"`
	OmniflowContentType string `json:"omniflow_content_type,omitempty"`
	OmniflowProfilePath string `json:"omniflow_profile_path,omitempty"`
	OmniflowProfileJson string `json:"omniflow_profile_json,omitempty"`
	OmniflowServerHost  string `json:"omniflow_server_host,omitempty"`
	OmniflowServerPort  int32  `json:"omniflow_server_port,omitempty"`
	OmniflowCaCertPath  string `json:"omniflow_ca_cert_path,omitempty"`
	OmniflowTargetMeta  string `json:"omniflow_target_meta,omitempty"`
	OmniflowSpkiPin     string `json:"omniflow_spki_pin,omitempty"`

	// OmniFlow H3 Fallback 策略
	OmniflowH3FallbackEnabled          bool   `json:"omniflow_h3_fallback_enabled,omitempty"`
	OmniflowH3FallbackPolicy           string `json:"omniflow_h3_fallback_policy,omitempty"`
	OmniflowH3FallbackTimeoutMs        int32  `json:"omniflow_h3_fallback_timeout_ms,omitempty"`
	OmniflowH3FallbackRetryBudget      int32  `json:"omniflow_h3_fallback_retry_budget,omitempty"`
	OmniflowH3FallbackSmokeEnabled     bool   `json:"omniflow_h3_fallback_smoke_enabled,omitempty"`
	OmniflowH3FallbackSmokeIntervalSec int32  `json:"omniflow_h3_fallback_smoke_interval_sec,omitempty"`
	OmniflowH3FallbackSmokeTimeoutMs   int32  `json:"omniflow_h3_fallback_smoke_timeout_ms,omitempty"`

	// OmniFlow 连接管理
	OmniflowMaxAgeSec      int32 `json:"omniflow_max_age_sec,omitempty"`
	OmniflowIdleTimeoutSec int32 `json:"omniflow_idle_timeout_sec,omitempty"`
	OmniflowMaxConnections int32 `json:"omniflow_max_connections,omitempty"`

	// OmniFlow 抗指纹
	OmniflowAdaptiveTlsEnabled    bool   `json:"omniflow_adaptive_tls_enabled,omitempty"`
	OmniflowTlsFingerprint        string `json:"omniflow_tls_fingerprint,omitempty"`
	OmniflowSniMode               string `json:"omniflow_sni_mode,omitempty"`
	OmniflowPaddingMode           string `json:"omniflow_padding_mode,omitempty"`
	OmniflowTrafficShapingEnabled bool   `json:"omniflow_traffic_shaping_enabled,omitempty"`
	OmniflowAfEnabled             bool   `json:"omniflow_af_enabled,omitempty"`
	OmniflowAfPathMode            string `json:"omniflow_af_path_mode,omitempty"`
	OmniflowAfPathPrefix          string `json:"omniflow_af_path_prefix,omitempty"`
	OmniflowAfPathSuffix          string `json:"omniflow_af_path_suffix,omitempty"`
	OmniflowAfPathRotationSecs    int32  `json:"omniflow_af_path_rotation_secs,omitempty"`
	OmniflowAfPathSkewSlots       int32  `json:"omniflow_af_path_skew_slots,omitempty"`

	// OmniFlow 同端口浏览器 Fallback 反向代理
	OmniflowFallbackEnabled      bool   `json:"omniflow_fallback_enabled,omitempty"`
	OmniflowFallbackTargetScheme string `json:"omniflow_fallback_target_scheme,omitempty"`
	OmniflowFallbackTargetHost   string `json:"omniflow_fallback_target_host,omitempty"`
	OmniflowFallbackTargetPort   int32  `json:"omniflow_fallback_target_port,omitempty"`
	OmniflowFallbackHostHeader   string `json:"omniflow_fallback_host_header,omitempty"`
	OmniflowFallbackTLSSNI       string `json:"omniflow_fallback_tls_sni,omitempty"`

	// OmniFlow 回退 Carrier
	OmniflowFallbackCarrierEnabled bool `json:"omniflow_fallback_carrier_enabled,omitempty"`
	OmniflowFallbackConnectTunnel  bool `json:"omniflow_fallback_connect_tunnel,omitempty"`
	OmniflowFallbackWssEnabled     bool `json:"omniflow_fallback_wss_enabled,omitempty"`
}

func (p *Protocol) NormalizeSimnet() {
	if p == nil || p.Type != "simnet" {
		return
	}
	if strings.TrimSpace(p.SimnetPath) == "" {
		p.SimnetPath = "/simnet/session"
	}
	if !p.SimnetFallbackEnabled || strings.TrimSpace(p.SimnetFallbackTargetHost) == "" {
		p.SimnetFallbackEnabled = false
		p.SimnetFallbackTargetScheme = ""
		p.SimnetFallbackTargetHost = ""
		p.SimnetFallbackTargetPort = 0
		p.SimnetFallbackHostHeader = ""
		p.SimnetFallbackTLSSNI = ""
	} else {
		p.SimnetFallbackTargetHost = strings.TrimSpace(p.SimnetFallbackTargetHost)
		p.SimnetFallbackHostHeader = strings.TrimSpace(p.SimnetFallbackHostHeader)
		p.SimnetFallbackTLSSNI = strings.TrimSpace(p.SimnetFallbackTLSSNI)
		switch strings.ToLower(strings.TrimSpace(p.SimnetFallbackTargetScheme)) {
		case "http", "https":
			p.SimnetFallbackTargetScheme = strings.ToLower(strings.TrimSpace(p.SimnetFallbackTargetScheme))
		default:
			p.SimnetFallbackTargetScheme = "https"
		}
	}
	if !p.SimnetAfEnabled {
		p.SimnetAfPathMode = ""
		p.SimnetAfMagicMode = ""
		p.SimnetAfPathPrefix = ""
		p.SimnetAfPathSuffix = ""
		p.SimnetAfResponseJitterMs = 0
		p.SimnetAfHandshakePolymorphism = false
		p.SimnetAfSettingsJitter = false
		p.SimnetAfFakeHeaderInjection = false
		return
	}
	if p.SimnetAfPathMode == "" {
		p.SimnetAfPathMode = "api"
	}
	if p.SimnetAfMagicMode == "" {
		p.SimnetAfMagicMode = "derived"
	}
	if p.SimnetAfResponseJitterMs == 0 {
		p.SimnetAfResponseJitterMs = 50
	}
	if !p.SimnetAfHandshakePolymorphism {
		p.SimnetAfHandshakePolymorphism = true
	}
	if !p.SimnetAfSettingsJitter {
		p.SimnetAfSettingsJitter = true
	}
	if !p.SimnetAfFakeHeaderInjection {
		p.SimnetAfFakeHeaderInjection = true
	}
}

func (p *Protocol) NormalizeOmniflow() {
	if p == nil || (p.Type != "omniflow" && p.Type != "omniflow-h3") {
		return
	}
	if !p.OmniflowFallbackEnabled || strings.TrimSpace(p.OmniflowFallbackTargetHost) == "" {
		p.OmniflowFallbackEnabled = false
		p.OmniflowFallbackTargetScheme = ""
		p.OmniflowFallbackTargetHost = ""
		p.OmniflowFallbackTargetPort = 0
		p.OmniflowFallbackHostHeader = ""
		p.OmniflowFallbackTLSSNI = ""
	} else {
		p.OmniflowFallbackTargetHost = strings.TrimSpace(p.OmniflowFallbackTargetHost)
		p.OmniflowFallbackHostHeader = strings.TrimSpace(p.OmniflowFallbackHostHeader)
		p.OmniflowFallbackTLSSNI = strings.TrimSpace(p.OmniflowFallbackTLSSNI)
		switch strings.ToLower(strings.TrimSpace(p.OmniflowFallbackTargetScheme)) {
		case "http", "https":
			p.OmniflowFallbackTargetScheme = strings.ToLower(strings.TrimSpace(p.OmniflowFallbackTargetScheme))
		default:
			p.OmniflowFallbackTargetScheme = "https"
		}
	}
	if !p.OmniflowAfEnabled {
		p.OmniflowAfPathMode = ""
		p.OmniflowAfPathPrefix = ""
		p.OmniflowAfPathSuffix = ""
		p.OmniflowAfPathRotationSecs = 0
		p.OmniflowAfPathSkewSlots = 0
		return
	}
	if strings.TrimSpace(p.OmniflowAfPathMode) == "" {
		p.OmniflowAfPathMode = "random"
	} else {
		p.OmniflowAfPathMode = strings.ToLower(strings.TrimSpace(p.OmniflowAfPathMode))
	}
	p.OmniflowAfPathPrefix = strings.TrimSpace(p.OmniflowAfPathPrefix)
	p.OmniflowAfPathSuffix = strings.TrimSpace(p.OmniflowAfPathSuffix)
	if p.OmniflowAfPathRotationSecs <= 0 {
		p.OmniflowAfPathRotationSecs = 300
	}
	if p.OmniflowAfPathSkewSlots <= 0 {
		p.OmniflowAfPathSkewSlots = 1
	}
}

// MarshalProtocols converts protocol array to JSON string
func MarshalProtocols(protocols []*Protocol) (string, error) {
	if len(protocols) == 0 {
		return "", nil
	}
	data, err := json.Marshal(protocols)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalProtocols converts JSON string to protocol array
func UnmarshalProtocols(protocolsJSON string) ([]*Protocol, error) {
	var protocols []*Protocol
	if protocolsJSON == "" {
		return protocols, nil
	}
	err := json.Unmarshal([]byte(protocolsJSON), &protocols)
	if err != nil {
		return nil, err
	}
	return protocols, nil
}

// ValidateProtocols validates protocol list
func ValidateProtocols(protocols []*Protocol) error {
	// 验证protocol类型唯一性
	seen := make(map[string]bool)
	for _, p := range protocols {
		if p.Type == "" {
			return ErrProtocolTypeRequired
		}
		if seen[p.Type] {
			return ErrDuplicateProtocolType
		}
		seen[p.Type] = true
	}
	return nil
}
