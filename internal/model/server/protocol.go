package server

import "encoding/json"

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
