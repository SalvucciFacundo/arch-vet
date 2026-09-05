package rules

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/SalvucciFacundo/arch-vet/internal/config"
)

// DomainLeakRule (ARCH002) checks for infrastructure or transport types leaking into domain and application layers.
type DomainLeakRule struct{}

func (r *DomainLeakRule) ID() string {
	return "ARCH002"
}

func (r *DomainLeakRule) Name() string {
	return "domain-leaky-type"
}

// Leaky types commonly introduced mistakenly by AI coding agents into core layers.
var leakyTypeNames = map[string]string{
	"Context":        "gin/echo/fiber web framework context",
	"Request":        "http.Request transport object",
	"ResponseWriter": "http.ResponseWriter transport object",
	"DB":             "database/sql or gorm database connection",
	"Tx":             "database transaction object",
	"Row":            "database row cursor",
	"Rows":           "database rows cursor",
}

func (r *DomainLeakRule) Evaluate(ctx *Context) []Diagnostic {
	var diagnostics []Diagnostic

	for _, pkg := range ctx.Project.Packages {
		layer := ctx.Config.FindLayerByPath(pkg.RelPath)
		if layer == nil {
			continue
		}

		// Only inner business layers (domain, ports, application, usecases, entities) are evaluated
		if layer.Name != "domain" && layer.Name != "ports" && layer.Name != "application" &&
			layer.Name != "entities" && layer.Name != "usecases" {
			continue
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				switch node := n.(type) {
				// 1. Inspect Struct Fields
				case *ast.StructType:
					if node.Fields != nil {
						for _, field := range node.Fields.List {
							if typeName, isLeaky := checkType(field.Type); isLeaky {
								pos := pkg.Fset.Position(field.Pos())
								diagnostics = append(diagnostics, Diagnostic{
									RuleID:   r.ID(),
									RuleName: r.Name(),
									Severity: config.SeverityError,
									Position: pos,
									Message: fmt.Sprintf("layer %q struct field leaks infrastructure type %q (%s)",
										layer.Name, typeName, leakyTypeNames[typeName]),
									Explanation: fmt.Sprintf("Inner layer %q must not depend on transport or driver types. Leaking %q violates encapsulation and binds business logic to external frameworks.",
										layer.Name, typeName),
									Remediation: "Replace framework/database types with plain domain values or custom application DTOs.",
								})
							}
						}
					}

				// 2. Inspect Function and Method Signatures
				case *ast.FuncDecl:
					if node.Type != nil && node.Type.Params != nil {
						for _, param := range node.Type.Params.List {
							if typeName, isLeaky := checkType(param.Type); isLeaky {
								pos := pkg.Fset.Position(param.Pos())
								diagnostics = append(diagnostics, Diagnostic{
									RuleID:   r.ID(),
									RuleName: r.Name(),
									Severity: config.SeverityError,
									Position: pos,
									Message: fmt.Sprintf("layer %q function %q parameter leaks infrastructure type %q (%s)",
										layer.Name, node.Name.Name, typeName, leakyTypeNames[typeName]),
									Explanation: fmt.Sprintf("Core business functions in layer %q must be transport-agnostic. Accepting %q directly couples use-cases to an HTTP or persistence implementation.",
										layer.Name, typeName),
									Remediation: fmt.Sprintf("Pass a pure Go struct (command/query DTO) or context.Context instead of %s.", typeName),
								})
							}
						}
					}
				}
				return true
			})
		}
	}

	return diagnostics
}

func checkType(expr ast.Expr) (string, bool) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return checkType(t.X)
	case *ast.SelectorExpr:
		ident := t.Sel.Name
		pkgIdent, ok := t.X.(*ast.Ident)
		if ok {
			full := pkgIdent.Name + "." + ident
			// Whitelist standard library context.Context
			if full == "context.Context" {
				return "", false
			}
			for leakyKey := range leakyTypeNames {
				if ident == leakyKey || strings.HasSuffix(full, "."+leakyKey) {
					return full, true
				}
			}
		}
	case *ast.Ident:
		// Bare identifiers like Context, Request, ResponseWriter
		if _, exists := leakyTypeNames[t.Name]; exists {
			return t.Name, true
		}
	}
	return "", false
}
