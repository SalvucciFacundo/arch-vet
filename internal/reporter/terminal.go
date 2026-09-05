package reporter

import (
	"fmt"
	"io"

	"github.com/SalvucciFacundo/arch-vet/internal/rules"
)

// ANSI Color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorBold   = "\033[1m"
)

// TerminalReporter prints colorized diagnostics for human terminal users.
type TerminalReporter struct {
	NoColor bool
}

func (r *TerminalReporter) Report(w io.Writer, diagnostics []rules.Diagnostic) error {
	if len(diagnostics) == 0 {
		fmt.Fprintln(w, "✓ Architecture clean: no violations detected.")
		return nil
	}

	fmt.Fprintf(w, "%d architectural violation(s) found:\n\n", len(diagnostics))

	for _, d := range diagnostics {
		badge := "[ERROR]"
		cStart := colorRed + colorBold
		if d.Severity == "warning" {
			badge = "[WARN]"
			cStart = colorYellow + colorBold
		}

		if r.NoColor {
			fmt.Fprintf(w, "%s %s (%s)\n", badge, d.Position, d.RuleID)
			fmt.Fprintf(w, "  Message:     %s\n", d.Message)
			fmt.Fprintf(w, "  Remediation: %s\n\n", d.Remediation)
		} else {
			fmt.Fprintf(w, "%s%s%s %s%s%s (%s%s%s)\n",
				cStart, badge, colorReset,
				colorBold, d.Position, colorReset,
				colorCyan, d.RuleID, colorReset,
			)
			fmt.Fprintf(w, "  %sMessage:%s     %s\n", colorBold, colorReset, d.Message)
			fmt.Fprintf(w, "  %sRemediation:%s %s\n\n", colorBold, colorReset, d.Remediation)
		}
	}

	return nil
}
