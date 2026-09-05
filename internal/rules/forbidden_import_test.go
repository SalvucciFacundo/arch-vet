package rules

import (
	"go/token"
	"testing"

	"github.com/SalvucciFacundo/arch-vet/internal/analyzer"
	"github.com/SalvucciFacundo/arch-vet/internal/config"
)

func TestForbiddenImportRule_InternalViolation(t *testing.T) {
	cfg := config.HexagonalPreset()
	project := &analyzer.Project{
		ModulePath: "github.com/example/app",
		RootDir:    "/mock/app",
		Packages: map[string]*analyzer.PackageInfo{
			"github.com/example/app/internal/domain": {
				ID:      "github.com/example/app/internal/domain",
				Name:    "domain",
				RelPath: "internal/domain",
				Imports: []analyzer.ImportInfo{
					{
						Path:     "github.com/example/app/internal/adapters/db",
						Position: token.Position{Filename: "user.go", Line: 10, Column: 2},
					},
				},
			},
			"github.com/example/app/internal/adapters/db": {
				ID:      "github.com/example/app/internal/adapters/db",
				Name:    "db",
				RelPath: "internal/adapters/db",
			},
		},
	}

	ctx := &Context{
		Config:  cfg,
		Project: project,
	}

	rule := &ForbiddenImportRule{}
	diagnostics := rule.Evaluate(ctx)

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diagnostics))
	}

	d := diagnostics[0]
	if d.RuleID != "ARCH001" {
		t.Errorf("expected RuleID ARCH001, got %s", d.RuleID)
	}
	if d.Severity != config.SeverityError {
		t.Errorf("expected SeverityError, got %v", d.Severity)
	}
	if d.Remediation == "" {
		t.Errorf("expected non-empty remediation guidance for AI agents")
	}
}

func TestForbiddenImportRule_DenyExternal(t *testing.T) {
	cfg := config.HexagonalPreset()
	project := &analyzer.Project{
		ModulePath: "github.com/example/app",
		RootDir:    "/mock/app",
		Packages: map[string]*analyzer.PackageInfo{
			"github.com/example/app/internal/domain": {
				ID:      "github.com/example/app/internal/domain",
				Name:    "domain",
				RelPath: "internal/domain",
				Imports: []analyzer.ImportInfo{
					{
						Path:     "database/sql",
						Position: token.Position{Filename: "model.go", Line: 5, Column: 2},
					},
				},
			},
		},
	}

	ctx := &Context{
		Config:  cfg,
		Project: project,
	}

	rule := &ForbiddenImportRule{}
	diagnostics := rule.Evaluate(ctx)

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 external deny diagnostic, got %d", len(diagnostics))
	}

	d := diagnostics[0]
	if d.RuleID != "ARCH001" {
		t.Errorf("expected RuleID ARCH001, got %s", d.RuleID)
	}
}

func TestForbiddenImportRule_AllowedDependency(t *testing.T) {
	cfg := config.HexagonalPreset()
	project := &analyzer.Project{
		ModulePath: "github.com/example/app",
		RootDir:    "/mock/app",
		Packages: map[string]*analyzer.PackageInfo{
			"github.com/example/app/internal/adapters/http": {
				ID:      "github.com/example/app/internal/adapters/http",
				Name:    "http",
				RelPath: "internal/adapters/http",
				Imports: []analyzer.ImportInfo{
					{
						Path:     "github.com/example/app/internal/domain",
						Position: token.Position{Filename: "handler.go", Line: 8, Column: 2},
					},
				},
			},
			"github.com/example/app/internal/domain": {
				ID:      "github.com/example/app/internal/domain",
				Name:    "domain",
				RelPath: "internal/domain",
			},
		},
	}

	ctx := &Context{
		Config:  cfg,
		Project: project,
	}

	rule := &ForbiddenImportRule{}
	diagnostics := rule.Evaluate(ctx)

	if len(diagnostics) != 0 {
		t.Fatalf("expected 0 diagnostics for allowed adapter->domain dependency, got %d", len(diagnostics))
	}
}
