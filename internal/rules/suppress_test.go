package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/SalvucciFacundo/arch-vet/internal/analyzer"
	"github.com/SalvucciFacundo/arch-vet/internal/config"
)

func TestInlineSuppression(t *testing.T) {
	cfg := config.HexagonalPreset()
	fset := token.NewFileSet()

	code := `package domain

// arch-vet:ignore ARCH001
import "database/sql"

import "net/http" // arch-vet:ignore

type User struct {
	DB *sql.DB // arch-vet:ignore ARCH002
}
`
	parsed, err := parser.ParseFile(fset, "user.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse code: %v", err)
	}

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
				Imports: []analyzer.ImportInfo{
					{
						Path:     "database/sql",
						Position: token.Position{Filename: "user.go", Line: 4, Column: 2},
					},
					{
						Path:     "net/http",
						Position: token.Position{Filename: "user.go", Line: 6, Column: 2},
					},
				},
			},
		},
	}

	ctx := &Context{Config: cfg, Project: project}
	engine := NewEngine()
	diagnostics := engine.Evaluate(ctx)

	// Since database/sql (ARCH001), net/http (ARCH001), and DB *sql.DB (ARCH002) all have ignore comments,
	// all violations should be suppressed!
	if len(diagnostics) != 0 {
		t.Fatalf("expected 0 diagnostics after inline suppression, got %d: %+v", len(diagnostics), diagnostics)
	}
}

func TestInlineSuppression_Unsuppressed(t *testing.T) {
	cfg := config.HexagonalPreset()
	fset := token.NewFileSet()

	code := `package domain

import "database/sql"
`
	parsed, err := parser.ParseFile(fset, "order.go", code, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse code: %v", err)
	}

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
				Imports: []analyzer.ImportInfo{
					{
						Path:     "database/sql",
						Position: token.Position{Filename: "order.go", Line: 3, Column: 2},
					},
				},
			},
		},
	}

	ctx := &Context{Config: cfg, Project: project}
	engine := NewEngine()
	diagnostics := engine.Evaluate(ctx)

	if len(diagnostics) != 1 {
		t.Fatalf("expected 1 unsuppressed diagnostic, got %d", len(diagnostics))
	}
}
