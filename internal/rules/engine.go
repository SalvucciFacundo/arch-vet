package rules

// Engine manages the registry and evaluation of architectural rules.
type Engine struct {
	rules []Rule
}

// NewEngine initializes an Engine equipped with default core rules.
func NewEngine() *Engine {
	return &Engine{
		rules: []Rule{
			&ForbiddenImportRule{},
			&DomainLeakRule{},
			&ProducerInterfaceRule{},
		},
	}
}

// RegisterRule adds a custom rule to the engine.
func (e *Engine) RegisterRule(r Rule) {
	e.rules = append(e.rules, r)
}

// Evaluate runs all registered rules against the project context.
func (e *Engine) Evaluate(ctx *Context) []Diagnostic {
	var allDiagnostics []Diagnostic
	for _, rule := range e.rules {
		diags := rule.Evaluate(ctx)
		allDiagnostics = append(allDiagnostics, diags...)
	}
	return allDiagnostics
}
