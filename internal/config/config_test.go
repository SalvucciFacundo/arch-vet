package config

import (
	"testing"
)

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		path     string
		expected bool
	}{
		{"internal/domain/**", "internal/domain/user", true},
		{"internal/domain/**", "internal/domain/sub/user", true},
		{"internal/domain/**", "internal/adapters/http", false},
		{"internal/domain", "internal/domain", true},
		{"internal/domain", "internal/domain/user", false},
		{"internal/infra*", "internal/infrastructure", true},
		{"internal/infra*", "internal/infra", true},
	}

	for _, tt := range tests {
		got := matchPattern(tt.pattern, tt.path)
		if got != tt.expected {
			t.Errorf("matchPattern(%q, %q) = %v; want %v", tt.pattern, tt.path, got, tt.expected)
		}
	}
}

func TestHexagonalPresetFindLayer(t *testing.T) {
	cfg := HexagonalPreset()

	domainLayer := cfg.FindLayerByPath("internal/domain/model")
	if domainLayer == nil || domainLayer.Name != "domain" {
		t.Fatalf("expected domain layer, got: %v", domainLayer)
	}

	adapterLayer := cfg.FindLayerByPath("internal/adapters/http")
	if adapterLayer == nil || adapterLayer.Name != "adapters" {
		t.Fatalf("expected adapters layer, got: %v", adapterLayer)
	}

	portsLayer := cfg.FindLayerByPath("internal/ports")
	if portsLayer == nil || portsLayer.Name != "ports" {
		t.Fatalf("expected ports layer, got: %v", portsLayer)
	}
}
