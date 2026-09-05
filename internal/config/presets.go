package config

import (
	"fmt"
	"strings"
)

// HexagonalPreset defines default ports-and-adapters architecture rules.
func HexagonalPreset() *Config {
	return &Config{
		Version:      "v1",
		Architecture: "hexagonal",
		Layers: []Layer{
			{
				Name:        "domain",
				Description: "Pure business models, entities, and domain logic. Independent of external libraries.",
				Patterns: []string{
					"internal/domain/**",
					"internal/core/domain/**",
					"internal/domain",
					"internal/core/domain",
					"domain/**",
					"domain",
				},
				MayDependOn: []string{}, // Domain depends on nothing
				DenyExternal: []string{
					"database/sql",
					"net/http",
					"github.com/gin-gonic/gin",
					"github.com/gofiber/fiber",
					"github.com/labstack/echo",
					"gorm.io/gorm",
				},
			},
			{
				Name:        "ports",
				Description: "Input and output interface contracts.",
				Patterns: []string{
					"internal/ports/**",
					"internal/core/ports/**",
					"internal/ports",
					"internal/core/ports",
					"ports/**",
					"ports",
				},
				MayDependOn: []string{"domain"},
				DenyExternal: []string{
					"database/sql",
					"gorm.io/gorm",
				},
			},
			{
				Name:        "application",
				Description: "Use cases and application orchestration.",
				Patterns: []string{
					"internal/application/**",
					"internal/usecases/**",
					"internal/core/services/**",
					"internal/services/**",
					"internal/application",
					"internal/usecases",
					"internal/core/services",
					"internal/services",
					"application/**",
					"application",
				},
				MayDependOn: []string{"domain", "ports"},
				DenyExternal: []string{
					"database/sql",
					"net/http",
					"github.com/gin-gonic/gin",
					"github.com/gofiber/fiber",
					"gorm.io/gorm",
				},
			},
			{
				Name:        "adapters",
				Description: "Inbound (HTTP, gRPC, CLI) and outbound (DB, Redis, HTTP clients) adapters.",
				Patterns: []string{
					"internal/adapters/**",
					"internal/infra/**",
					"internal/infrastructure/**",
					"internal/adapters",
					"internal/infra",
					"internal/infrastructure",
					"adapters/**",
					"adapters",
					"infra/**",
					"infra",
				},
				MayDependOn: []string{"domain", "ports", "application"},
			},
		},
	}
}

// CleanPreset defines Clean Architecture layer rules.
func CleanPreset() *Config {
	return &Config{
		Version:      "v1",
		Architecture: "clean",
		Layers: []Layer{
			{
				Name:        "entities",
				Description: "Enterprise business rules and entities.",
				Patterns: []string{
					"internal/entity/**",
					"internal/entities/**",
					"internal/domain/entity/**",
					"entity/**",
					"entities/**",
				},
				MayDependOn: []string{},
				DenyExternal: []string{
					"database/sql",
					"net/http",
					"gorm.io/gorm",
				},
			},
			{
				Name:        "usecases",
				Description: "Application business rules.",
				Patterns: []string{
					"internal/usecase/**",
					"internal/usecases/**",
					"usecase/**",
					"usecases/**",
				},
				MayDependOn: []string{"entities"},
				DenyExternal: []string{
					"database/sql",
					"net/http",
					"gorm.io/gorm",
				},
			},
			{
				Name:        "controllers",
				Description: "Interface adapters: controllers, presenters, gateways.",
				Patterns: []string{
					"internal/controller/**",
					"internal/controllers/**",
					"internal/presenter/**",
					"internal/gateway/**",
				},
				MayDependOn: []string{"entities", "usecases"},
			},
			{
				Name:        "frameworks",
				Description: "Frameworks, web drivers, database repositories.",
				Patterns: []string{
					"internal/framework/**",
					"internal/infra/**",
					"internal/infrastructure/**",
				},
				MayDependOn: []string{"entities", "usecases", "controllers"},
			},
		},
	}
}

// GetPreset returns a preset by name.
func GetPreset(name string) (*Config, error) {
	switch strings.ToLower(name) {
	case "hexagonal", "hex":
		return HexagonalPreset(), nil
	case "clean":
		return CleanPreset(), nil
	default:
		return nil, fmt.Errorf("unknown preset %q (available: hexagonal, clean)", name)
	}
}
