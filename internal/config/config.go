package config

import "strings"

// RuleSeverity represents the severity level of a rule violation.
type RuleSeverity string

const (
	SeverityError   RuleSeverity = "error"
	SeverityWarning RuleSeverity = "warning"
	SeverityInfo    RuleSeverity = "info"
)

// Layer defines an architectural layer and its permitted dependencies.
type Layer struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	// Patterns are package path globs (e.g., "internal/core/domain/**", "internal/domain*").
	Patterns    []string `yaml:"patterns"`
	// MayDependOn lists layer names that this layer is allowed to import.
	MayDependOn []string `yaml:"may_depend_on"`
	// DenyExternal lists external packages strictly forbidden in this layer (e.g., "database/sql", "net/http").
	DenyExternal []string `yaml:"deny_external,omitempty"`
}

// Config is the root configuration structure for arch-vet.
type Config struct {
	Version      string   `yaml:"version"`
	Architecture string   `yaml:"architecture"` // e.g. "hexagonal", "clean", "layered", "custom"
	Layers       []Layer  `yaml:"layers"`
	Ignore       []string `yaml:"ignore,omitempty"`
}

// FindLayerByPath matches a relative package path to a configured layer.
func (c *Config) FindLayerByPath(pkgPath string) *Layer {
	for i := range c.Layers {
		for _, pattern := range c.Layers[i].Patterns {
			if matchPattern(pattern, pkgPath) {
				return &c.Layers[i]
			}
		}
	}
	return nil
}

// matchPattern evaluates if pkgPath matches pattern.
// Supports exact match, "/**" recursive subpackage match, and "*" prefix wildcard match.
func matchPattern(pattern, pkgPath string) bool {
	if pattern == pkgPath {
		return true
	}

	// Pattern like "internal/domain/**"
	if strings.HasSuffix(pattern, "/**") {
		base := strings.TrimSuffix(pattern, "/**")
		if pkgPath == base {
			return true
		}
		prefix := base + "/"
		return strings.HasPrefix(pkgPath, prefix)
	}

	// Pattern like "internal/domain/*"
	if strings.HasSuffix(pattern, "/*") {
		base := strings.TrimSuffix(pattern, "/*")
		prefix := base + "/"
		if strings.HasPrefix(pkgPath, prefix) {
			rest := strings.TrimPrefix(pkgPath, prefix)
			return !strings.Contains(rest, "/")
		}
		return false
	}

	// Pattern like "internal/infra*"
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(pkgPath, prefix)
	}

	return false
}
