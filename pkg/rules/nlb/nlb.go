package nlb

import (
	"io"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

type NLB struct{}

var (
	_ statement.Parser      = NLB{}
	_ statement.Detector    = NLB{}
	_ statement.LinesParser = NLB{}
)

// Exports since 2021 are comma-separated with Slovene-only headers; older
// exports are semicolon-separated with bilingual "Slovene/English" headers,
// which the shared engine's base dictionary already resolves via the
// English tail.
var csvHints = statement.CSVHints{
	Aliases: map[string]statement.Field{
		"Namen":                     statement.FieldDescription,
		"Kategorija":                statement.FieldCategory,
		"Znesek":                    statement.FieldAmount,
		"Valuta":                    statement.FieldCurrency,
		"Datum plačila":             statement.FieldValueDate,
		"Naziv prejemnika/plačnika": statement.FieldCounterparty,
		"Račun prejemnika/plačnika": statement.FieldAccount,
		"Referenca prejemnika":      statement.FieldReference,
		"Datum poravnave":           statement.FieldSettleDate,
		"ID transakcije":            statement.FieldTxnID,
	},
	DefaultCurrency:      "EUR",
	SkipStatusValues:     []string{"AVTORIZACIJA"},
	MergePurposeIntoDesc: true,
}

func ParseCSV(r io.Reader) ([]statement.Transaction, error) {
	return statement.ParseGenericCSV(r, csvHints)
}

func (NLB) ParseCSV(r io.Reader) ([]statement.Transaction, error) { return ParseCSV(r) }
func (NLB) ParsePDF(r io.Reader) ([]statement.Transaction, error) { return ParsePDF(r) }

// Only the 2021+ Slovene-only headers are covered here; older bilingual
// exports resolve via the shared base dictionary and carry no bank-specific
// signal, so they fall back to generic and need -bank nlb.
func (NLB) DetectCSV(header []string) int { return statement.DetectCSVScore(csvHints, header) }
