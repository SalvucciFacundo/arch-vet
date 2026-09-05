package analyzer

import (
	"go/ast"
	"go/token"
	"go/types"
)

// ImportInfo represents a single import statement in a source file.
type ImportInfo struct {
	Path     string        // e.g. "github.com/SalvucciFacundo/arch-vet/internal/adapters"
	Position token.Position
	File     string
}

// PackageInfo represents a Go package analyzed by arch-vet.
type PackageInfo struct {
	ID          string                  // Full package path
	Name        string                  // Package name (e.g. "domain")
	RelPath     string                  // Relative directory path to module root (e.g. "internal/domain")
	Imports     []ImportInfo            // All package imports across files
	Fset        *token.FileSet
	Syntax      []*ast.File
	Types       *types.Package
	TypesInfo   *types.Info
}

// Project represents the loaded Go module and its constituent packages.
type Project struct {
	ModulePath string
	RootDir    string
	Packages   map[string]*PackageInfo // keyed by full package ID
}
