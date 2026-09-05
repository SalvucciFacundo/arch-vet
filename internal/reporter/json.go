package reporter

import (
	"encoding/json"
	"io"

	"github.com/SalvucciFacundo/arch-vet/internal/rules"
)

// JSONReporter formats diagnostics as JSON.
type JSONReporter struct{}

type jsonReport struct {
	TotalViolations int                `json:"total_violations"`
	Violations      []rules.Diagnostic `json:"violations"`
}

func (r *JSONReporter) Report(w io.Writer, diagnostics []rules.Diagnostic) error {
	report := jsonReport{
		TotalViolations: len(diagnostics),
		Violations:      diagnostics,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
