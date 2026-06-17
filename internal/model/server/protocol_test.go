package server

import "testing"

func TestProtocolNormalizeSimnetClearsDisabledFallback(t *testing.T) {
	protocol := &Protocol{
		Type:                          "simnet",
		SimnetFallbackEnabled:         false,
		SimnetFallbackTargetScheme:    "https",
		SimnetFallbackTargetHost:      "www.example.com",
		SimnetFallbackTargetPort:      443,
		SimnetFallbackHostHeader:      "www.example.com",
		SimnetFallbackTLSSNI:          "www.example.com",
		SimnetAfEnabled:               false,
		SimnetAfPathMode:              "random",
		SimnetAfMagicMode:             "random",
		SimnetAfResponseJitterMs:      20,
		SimnetAfHandshakePolymorphism: true,
		SimnetAfSettingsJitter:        true,
		SimnetAfFakeHeaderInjection:   true,
	}

	protocol.NormalizeSimnet()

	if protocol.SimnetFallbackEnabled {
		t.Fatal("fallback should remain disabled")
	}
	if protocol.SimnetFallbackTargetScheme != "" ||
		protocol.SimnetFallbackTargetHost != "" ||
		protocol.SimnetFallbackTargetPort != 0 ||
		protocol.SimnetFallbackHostHeader != "" ||
		protocol.SimnetFallbackTLSSNI != "" {
		t.Fatalf("disabled fallback fields were not cleared: %+v", protocol)
	}
	if protocol.SimnetPath != "/simnet/session" {
		t.Fatalf("expected default simnet path, got %q", protocol.SimnetPath)
	}
	if protocol.SimnetAfPathMode != "" || protocol.SimnetAfMagicMode != "" {
		t.Fatalf("expected disabled AF fields to be cleared, got path=%q magic=%q", protocol.SimnetAfPathMode, protocol.SimnetAfMagicMode)
	}
}

func TestProtocolNormalizeSimnetKeepsEnabledFallback(t *testing.T) {
	protocol := &Protocol{
		Type:                       "simnet",
		SimnetFallbackEnabled:      true,
		SimnetFallbackTargetScheme: " HTTPS ",
		SimnetFallbackTargetHost:   " www.example.com ",
		SimnetFallbackHostHeader:   " fallback.example.com ",
		SimnetFallbackTLSSNI:       " tls.example.com ",
	}

	protocol.NormalizeSimnet()

	if !protocol.SimnetFallbackEnabled {
		t.Fatal("fallback should stay enabled")
	}
	if protocol.SimnetFallbackTargetScheme != "https" {
		t.Fatalf("expected normalized scheme, got %q", protocol.SimnetFallbackTargetScheme)
	}
	if protocol.SimnetFallbackTargetHost != "www.example.com" {
		t.Fatalf("expected trimmed target host, got %q", protocol.SimnetFallbackTargetHost)
	}
	if protocol.SimnetFallbackHostHeader != "fallback.example.com" {
		t.Fatalf("expected trimmed host header, got %q", protocol.SimnetFallbackHostHeader)
	}
	if protocol.SimnetFallbackTLSSNI != "tls.example.com" {
		t.Fatalf("expected trimmed TLS SNI, got %q", protocol.SimnetFallbackTLSSNI)
	}
}

func TestProtocolNormalizeSimnetDefaultsEnabledAfSubFeatures(t *testing.T) {
	protocol := &Protocol{
		Type:                     "simnet",
		SimnetAfEnabled:          true,
		SimnetAfPathMode:         "random",
		SimnetAfMagicMode:        "derived",
		SimnetAfResponseJitterMs: 50,
	}

	protocol.NormalizeSimnet()

	if !protocol.SimnetAfHandshakePolymorphism {
		t.Fatal("expected enabled AF to default handshake polymorphism on")
	}
	if !protocol.SimnetAfSettingsJitter {
		t.Fatal("expected enabled AF to default settings jitter on")
	}
	if !protocol.SimnetAfFakeHeaderInjection {
		t.Fatal("expected enabled AF to default fake header injection on")
	}
}

func TestProtocolNormalizeOmniflowClearsDisabledAfPath(t *testing.T) {
	protocol := &Protocol{
		Type:                         "omniflow",
		OmniflowFallbackEnabled:      false,
		OmniflowFallbackTargetScheme: "https",
		OmniflowFallbackTargetHost:   "www.example.com",
		OmniflowFallbackTargetPort:   443,
		OmniflowFallbackHostHeader:   "www.example.com",
		OmniflowFallbackTLSSNI:       "www.example.com",
		OmniflowAfEnabled:            false,
		OmniflowAfPathMode:           "random",
		OmniflowAfPathPrefix:         "/cdn",
		OmniflowAfPathSuffix:         ".woff2",
		OmniflowAfPathRotationSecs:   120,
		OmniflowAfPathSkewSlots:      2,
	}

	protocol.NormalizeOmniflow()

	if protocol.OmniflowAfPathMode != "" ||
		protocol.OmniflowAfPathPrefix != "" ||
		protocol.OmniflowAfPathSuffix != "" ||
		protocol.OmniflowAfPathRotationSecs != 0 ||
		protocol.OmniflowAfPathSkewSlots != 0 {
		t.Fatalf("expected disabled OmniFlow AF path fields to be cleared, got %+v", protocol)
	}
	if protocol.OmniflowFallbackEnabled ||
		protocol.OmniflowFallbackTargetScheme != "" ||
		protocol.OmniflowFallbackTargetHost != "" ||
		protocol.OmniflowFallbackTargetPort != 0 ||
		protocol.OmniflowFallbackHostHeader != "" ||
		protocol.OmniflowFallbackTLSSNI != "" {
		t.Fatalf("expected disabled OmniFlow fallback fields to be cleared, got %+v", protocol)
	}
}

func TestProtocolNormalizeOmniflowKeepsEnabledFallback(t *testing.T) {
	protocol := &Protocol{
		Type:                         "omniflow-h3",
		OmniflowFallbackEnabled:      true,
		OmniflowFallbackTargetScheme: " HTTP ",
		OmniflowFallbackTargetHost:   " www.example.com ",
		OmniflowFallbackHostHeader:   " fallback.example.com ",
		OmniflowFallbackTLSSNI:       " tls.example.com ",
	}

	protocol.NormalizeOmniflow()

	if !protocol.OmniflowFallbackEnabled {
		t.Fatal("fallback should stay enabled")
	}
	if protocol.OmniflowFallbackTargetScheme != "http" {
		t.Fatalf("expected normalized scheme, got %q", protocol.OmniflowFallbackTargetScheme)
	}
	if protocol.OmniflowFallbackTargetHost != "www.example.com" {
		t.Fatalf("expected trimmed target host, got %q", protocol.OmniflowFallbackTargetHost)
	}
	if protocol.OmniflowFallbackHostHeader != "fallback.example.com" {
		t.Fatalf("expected trimmed host header, got %q", protocol.OmniflowFallbackHostHeader)
	}
	if protocol.OmniflowFallbackTLSSNI != "tls.example.com" {
		t.Fatalf("expected trimmed TLS SNI, got %q", protocol.OmniflowFallbackTLSSNI)
	}
}

func TestProtocolNormalizeOmniflowDefaultsEnabledAfPath(t *testing.T) {
	protocol := &Protocol{
		Type:              "omniflow-h3",
		OmniflowAfEnabled: true,
	}

	protocol.NormalizeOmniflow()

	if protocol.OmniflowAfPathMode != "random" {
		t.Fatalf("expected default random path mode, got %q", protocol.OmniflowAfPathMode)
	}
	if protocol.OmniflowAfPathRotationSecs != 300 {
		t.Fatalf("expected default rotation 300, got %d", protocol.OmniflowAfPathRotationSecs)
	}
	if protocol.OmniflowAfPathSkewSlots != 1 {
		t.Fatalf("expected default skew slots 1, got %d", protocol.OmniflowAfPathSkewSlots)
	}
}
