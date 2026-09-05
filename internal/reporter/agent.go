package reporter

import (
	"fmt"
	"io"

	"github.com/SalvucciFacundo/arch-vet/internal/rules"
)

// AgentReporter formats diagnostics as rich markdown optimized for LLMs and AI coding agents.
type AgentReporter struct{}

func (r *AgentReporter) Report(w io.Writer, diagnostics []rules.Diagnostic) error {
	if len(diagnostics) == 0 {
		fmt.Fprintf(w, "## Architecture Verification Passed\n\nNo architectural violations detected. Changes align with configured layer boundaries.\n")
		return nil
	}

	fmt.Fprintf(w, "## Architecture Violations Detected (%d)\n\n", len(diagnostics))
	fmt.Fprintln(w, "Please fix the following architectural regressions before finalizing your changes:")
	fmt.Fprintln(w)

	for i, d := range diagnostics {
		fmt.Fprintf(w, "### %d. [%s] %s\n", i+1, d.RuleID, d.RuleName)
		fmt.Fprintf(w, "* **Location:** `%s`\n", d.Position)
		fmt.Fprintf(w, "* **Problem:** %s\n", d.Message)
		fmt.Fprintf(w, "* **Architectural Rationale:** %s\n", d.Explanation)
		fmt.Fprintf(w, "* **Self-Healing Action:** %s\n\n", d.Remediation)
	}

	return nil
}
