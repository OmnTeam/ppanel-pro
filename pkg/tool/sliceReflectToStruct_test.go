package tool

import "testing"

func TestSystemConfigSliceReflectToStructSkipsUnknownKeys(t *testing.T) {
	type nodeConfig struct {
		NodeSecret       string
		NodePullInterval int
		Enabled          bool
	}

	var cfg nodeConfig
	SystemConfigSliceReflectToStruct([]*SystemConfig{
		{Key: "NodeSecret", Value: "secret", Type: "string"},
		{Key: "NodeMultiplierConfig", Value: "[]", Type: "string"},
		{Key: "NodePullInterval", Value: "10", Type: "int"},
		{Key: "Enabled", Value: "true", Type: "bool"},
	}, &cfg)

	if cfg.NodeSecret != "secret" {
		t.Fatalf("NodeSecret = %q, want %q", cfg.NodeSecret, "secret")
	}
	if cfg.NodePullInterval != 10 {
		t.Fatalf("NodePullInterval = %d, want 10", cfg.NodePullInterval)
	}
	if !cfg.Enabled {
		t.Fatal("Enabled = false, want true")
	}
}
