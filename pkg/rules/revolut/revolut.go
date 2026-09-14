package revolut

import (
	"io"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

type Revolut struct{}

var (
	_ statement.Parser      = Revolut{}
	_ statement.Detector    = Revolut{}
	_ statement.LinesParser = Revolut{}
)

// Started/Completed Date and State resolve via the shared base dictionary,
// so Signature lists them explicitly to keep detection working.
var csvHints = statement.CSVHints{
	Aliases: map[string]statement.Field{
		"Type": statement.FieldCategory,
	},
	Signature:        []string{"Started Date", "Completed Date", "State"},
	SkipStatusValues: []string{"REVERTED", "DECLINED", "FAILED", "PENDING"},
}

func ParseCSV(r io.Reader) ([]statement.Transaction, error) {
	return statement.ParseGenericCSV(r, csvHints)
}

func (Revolut) ParseCSV(r io.Reader) ([]statement.Transaction, error) { return ParseCSV(r) }
func (Revolut) ParsePDF(r io.Reader) ([]statement.Transaction, error) { return ParsePDF(r) }

func (Revolut) DetectCSV(header []string) int { return statement.DetectCSVScore(csvHints, header) }
