package rules

import (
	"go/token"

	"github.com/SalvucciFacundo/arch-vet/internal/analyzer"
	"github.com/SalvucciFacundo/arch-vet/internal/config"
)

// Diagnostic represents an architectural rule violation.
type Diagnostic struct {
	RuleID      string              `json:"rule_id"`
	RuleName    string              `json:"rule_name"`
	Severity    config.RuleSeverity `json:"severity"`
	Position    token.Position      `json:"position"`
	Message     string              `json:"message"`
	Explanation string              `json:"explanation"`
	Remediation string              `json:"remediation"`
}

// Context contains all loaded project data and configuration needed to evaluate rules.
type Context struct {
	Config  *config.Config
	Project *analyzer.Project
}

// Rule defines the interface that all architectural rules must implement.
type Rule interface {
	ID() string
	Name() string
	Evaluate(ctx *Context) []Diagnostic
}
