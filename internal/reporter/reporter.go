package reporter

import (
	"io"

	"github.com/SalvucciFacundo/arch-vet/internal/rules"
)

// Reporter formats and prints architectural diagnostics.
type Reporter interface {
	Report(w io.Writer, diagnostics []rules.Diagnostic) error
}
