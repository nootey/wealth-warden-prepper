package generic

import (
	"fmt"
	"io"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

// Generic runs the shared best-effort CSV engine with no bank-specific
// hints, for statements from a bank with no rule set of its own yet.
type Generic struct{}

var _ statement.Parser = Generic{}

func (Generic) ParseCSV(r io.Reader) ([]statement.Transaction, error) {
	return statement.ParseGenericCSV(r, statement.CSVHints{})
}

func (Generic) ParsePDF(r io.Reader) ([]statement.Transaction, error) {
	return nil, fmt.Errorf("generic PDF parsing is not supported yet; add a bank-specific rule set")
}
