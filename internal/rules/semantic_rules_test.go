package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/SalvucciFacundo/arch-vet/internal/analyzer"
	"github.com/SalvucciFacundo/arch-vet/internal/config"
)

func parseGoCode(t *testing.T, fset *token.FileSet, filename, src string) *ast.File {
	file, err := parser.ParseFile(fset, filename, src, parser.AllErrors)
	if err != nil {
		t.Fatalf("failed to parse code: %v", err)
	}
	return file
}

func TestDomainLeakRule_StructFieldLeak(t *testing.T) {
	cfg := config.HexagonalPreset()
	fset := token.NewFileSet()

	code := `package domain

import "database/sql"

type User struct {
	ID int
	DB *sql.DB
}
`
	parsed := parseGoCode(t, fset, "user.go", code)

	project := &analyzer.Project{
		ModulePath: "github.com/example/app",
		RootDir:    "/mock/app",
		Packages: map[string]*analyzer.PackageInfo{
			"github.com/example/app/internal/domain": {
				ID:      "github.com/example/app/internal/domain",
				Name:    "domain",
				RelPath: "internal/domain",
				Fset:    fset,
				Syntax:  []*ast.File{parsed},
			},
		},
	}

	ctx := &Context{Config: cfg, Project: project}
	rule := &DomainLeakRule{}
	diagnostics := rule.Evaluate(ctx)

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic for leaky struct field, got %d", len(diagnostics))
	}

	d := diagnostics[0]
	if d.RuleID != "ARCH002" {
		t.Errorf("expected RuleID ARCH002, got %s", d.RuleID)
	}
}

func TestDomainLeakRule_FuncParamLeak(t *testing.T) {
	cfg := config.HexagonalPreset()
	fset := token.NewFileSet()

	code := `package services

import "net/http"

type OrderService struct{}

func (s *OrderService) HandleOrder(r *http.Request) error {
	return nil
}
`
	parsed := parseGoCode(t, fset, "order.go", code)

	project := &analyzer.Project{
		ModulePath: "github.com/example/app",
		RootDir:    "/mock/app",
		Packages: map[string]*analyzer.PackageInfo{
			"github.com/example/app/internal/application": {
				ID:      "github.com/example/app/internal/application",
				Name:    "services",
				RelPath: "internal/application",
				Fset:    fset,
				Syntax:  []*ast.File{parsed},
			},
		},
	}

	ctx := &Context{Config: cfg, Project: project}
	rule := &DomainLeakRule{}
	diagnostics := rule.Evaluate(ctx)

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic for leaky func param, got %d", len(diagnostics))
	}

	d := diagnostics[0]
	if d.RuleID != "ARCH002" {
		t.Errorf("expected RuleID ARCH002, got %s", d.RuleID)
	}
}

func TestProducerInterfaceRule_FlagAdapterInterface(t *testing.T) {
	cfg := config.HexagonalPreset()
	fset := token.NewFileSet()

	code := `package postgres

type UserRepository interface {
	Find(id string) error
}

type PostgresRepo struct{}
`
	parsed := parseGoCode(t, fset, "repo.go", code)

	project := &analyzer.Project{
		ModulePath: "github.com/example/app",
		RootDir:    "/mock/app",
		Packages: map[string]*analyzer.PackageInfo{
			"github.com/example/app/internal/adapters/postgres": {
				ID:      "github.com/example/app/internal/adapters/postgres",
				Name:    "postgres",
				RelPath: "internal/adapters/postgres",
				Fset:    fset,
				Syntax:  []*ast.File{parsed},
			},
		},
	}

	ctx := &Context{Config: cfg, Project: project}
	rule := &ProducerInterfaceRule{}
	diagnostics := rule.Evaluate(ctx)

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 warning for producer interface in adapter, got %d", len(diagnostics))
	}

	d := diagnostics[0]
	if d.RuleID != "ARCH003" {
		t.Errorf("expected RuleID ARCH003, got %s", d.RuleID)
	}
	if d.Severity != config.SeverityWarning {
		t.Errorf("expected SeverityWarning, got %v", d.Severity)
	}
}
