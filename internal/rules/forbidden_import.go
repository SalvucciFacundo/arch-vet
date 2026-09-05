package rules

import (
	"fmt"
	"strings"

	"github.com/SalvucciFacundo/arch-vet/internal/config"
)

// ForbiddenImportRule (ARCH001) verifies that packages do not import layers or external libraries that violate boundary constraints.
type ForbiddenImportRule struct{}

func (r *ForbiddenImportRule) ID() string {
	return "ARCH001"
}

func (r *ForbiddenImportRule) Name() string {
	return "forbidden-import"
}

func (r *ForbiddenImportRule) Evaluate(ctx *Context) []Diagnostic {
	var diagnostics []Diagnostic

	for _, pkg := range ctx.Project.Packages {
		srcLayer := ctx.Config.FindLayerByPath(pkg.RelPath)
		if srcLayer == nil {
			continue // Package is not governed by any layer definition
		}

		for _, imp := range pkg.Imports {
			// 1. Check external library deny-list
			for _, denied := range srcLayer.DenyExternal {
				if imp.Path == denied || strings.HasPrefix(imp.Path, denied+"/") {
					diagnostics = append(diagnostics, Diagnostic{
						RuleID:   r.ID(),
						RuleName: r.Name(),
						Severity: config.SeverityError,
						Position: imp.Position,
						Message: fmt.Sprintf("layer %q cannot import external package %q",
							srcLayer.Name, imp.Path),
						Explanation: fmt.Sprintf("Layer %q is an inner layer that must remain pure and free from concrete infrastructure drivers or frameworks (%s).",
							srcLayer.Name, imp.Path),
						Remediation: fmt.Sprintf("Define an abstraction or interface in layer %q or an adjacent ports layer, and inject the concrete implementation from an outer adapter layer.",
							srcLayer.Name),
					})
				}
			}

			// 2. Check internal layer-to-layer dependency rules
			if ctx.Project.ModulePath != "" && strings.HasPrefix(imp.Path, ctx.Project.ModulePath) {
				relTarget := strings.TrimPrefix(imp.Path, ctx.Project.ModulePath)
				relTarget = strings.TrimPrefix(relTarget, "/")

				targetLayer := ctx.Config.FindLayerByPath(relTarget)
				if targetLayer == nil || targetLayer.Name == srcLayer.Name {
					continue // Same layer or unmanaged layer
				}

				allowed := false
				for _, allowedLayer := range srcLayer.MayDependOn {
					if targetLayer.Name == allowedLayer {
						allowed = true
						break
					}
				}

				if !allowed {
					diagnostics = append(diagnostics, Diagnostic{
						RuleID:   r.ID(),
						RuleName: r.Name(),
						Severity: config.SeverityError,
						Position: imp.Position,
						Message: fmt.Sprintf("architectural violation: layer %q is not permitted to import layer %q (%s)",
							srcLayer.Name, targetLayer.Name, imp.Path),
						Explanation: fmt.Sprintf("In %s architecture, dependencies must point inward. %q is not allowed to depend on %q.",
							ctx.Config.Architecture, srcLayer.Name, targetLayer.Name),
						Remediation: fmt.Sprintf("Invert the dependency using the Dependency Inversion Principle (DIP): define an interface in %q and implement it inside %q.",
							srcLayer.Name, targetLayer.Name),
					})
				}
			}
		}
	}

	return diagnostics
}
