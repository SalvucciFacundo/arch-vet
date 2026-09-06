package reporter

import (
	"bytes"
	"go/token"
	"strings"
	"testing"

	"github.com/SalvucciFacundo/arch-vet/internal/config"
	"github.com/SalvucciFacundo/arch-vet/internal/rules"
)

func sampleDiagnostics() []rules.Diagnostic {
	return []rules.Diagnostic{
		{
			RuleID:      "ARCH001",
			RuleName:    "forbidden-import",
			Severity:    config.SeverityError,
			Position:    token.Position{Filename: "internal/domain/user.go", Line: 12, Column: 2},
			Message:     "layer \"domain\" cannot import external package \"database/sql\"",
			Explanation: "Domain must remain pure.",
			Remediation: "Define an interface in domain and implement it in adapters.",
		},
	}
}

func TestTerminalReporter(t *testing.T) {
	rep := &TerminalReporter{NoColor: true}
	var buf bytes.Buffer
	err := rep.Report(&buf, sampleDiagnostics())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "[ERROR]") || !strings.Contains(out, "ARCH001") {
		t.Errorf("expected terminal output to contain [ERROR] and ARCH001, got: %s", out)
	}
}

func TestJSONReporter(t *testing.T) {
	rep := &JSONReporter{}
	var buf bytes.Buffer
	err := rep.Report(&buf, sampleDiagnostics())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"total_violations": 1`) || !strings.Contains(out, `"ARCH001"`) {
		t.Errorf("expected json output to contain violation count and ARCH001, got: %s", out)
	}
}

func TestAgentReporter(t *testing.T) {
	rep := &AgentReporter{}
	var buf bytes.Buffer
	err := rep.Report(&buf, sampleDiagnostics())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "## Architecture Violations Detected (1)") ||
		!strings.Contains(out, "* **Self-Healing Action:**") {
		t.Errorf("expected agent output to contain markdown header and self-healing action, got: %s", out)
	}
}

func TestSARIFReporter(t *testing.T) {
	rep := &SARIFReporter{ToolVersion: "0.2.0-test"}
	var buf bytes.Buffer
	err := rep.Report(&buf, sampleDiagnostics())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"version": "2.1.0"`) ||
		!strings.Contains(out, `"name": "arch-vet"`) ||
		!strings.Contains(out, `"ruleId": "ARCH001"`) ||
		!strings.Contains(out, `"startLine": 12`) {
		t.Errorf("expected SARIF output to contain schema 2.1.0, arch-vet driver, and ARCH001 result, got: %s", out)
	}
}
