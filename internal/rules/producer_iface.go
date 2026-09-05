package rules

import (
	"fmt"
	"go/ast"

	"github.com/SalvucciFacundo/arch-vet/internal/config"
)

// ProducerInterfaceRule (ARCH003) flags interfaces defined inside adapter or infrastructure layers.
// In idiomatic Go, interfaces belong in the consumer/client package, while adapters should export concrete structs.
type ProducerInterfaceRule struct{}

func (r *ProducerInterfaceRule) ID() string {
	return "ARCH003"
}

func (r *ProducerInterfaceRule) Name() string {
	return "producer-interface"
}

func (r *ProducerInterfaceRule) Evaluate(ctx *Context) []Diagnostic {
	var diagnostics []Diagnostic

	for _, pkg := range ctx.Project.Packages {
		layer := ctx.Config.FindLayerByPath(pkg.RelPath)
		if layer == nil {
			continue
		}

		// Check if package belongs to an outer infrastructure or adapter layer
		if layer.Name != "adapters" && layer.Name != "infra" && layer.Name != "infrastructure" && layer.Name != "frameworks" {
			continue
		}

		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				if _, isIface := typeSpec.Type.(*ast.InterfaceType); isIface {
					pos := pkg.Fset.Position(typeSpec.Pos())
					diagnostics = append(diagnostics, Diagnostic{
						RuleID:   r.ID(),
						RuleName: r.Name(),
						Severity: config.SeverityWarning,
						Position: pos,
						Message: fmt.Sprintf("interface %q defined in adapter/infrastructure layer %q",
							typeSpec.Name.Name, layer.Name),
						Explanation: "In idiomatic Go, interfaces belong in the package that consumes them (ports/domain/application), not in the package that provides the concrete implementation.",
						Remediation: fmt.Sprintf("Move interface %q to the consuming package or ports layer, and export a concrete struct from %q.",
							typeSpec.Name.Name, layer.Name),
					})
				}
				return true
			})
		}
	}

	return diagnostics
}
