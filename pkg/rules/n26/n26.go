package n26

import (
	"io"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

type N26 struct{}

var _ statement.Parser = N26{}

var csvHints = statement.CSVHints{
	Aliases: map[string]statement.Field{
		"Partner Name":      statement.FieldCounterparty,
		"Partner Iban":      statement.FieldAccount,
		"Payment Reference": statement.FieldDescription,
		"Amount (EUR)":      statement.FieldAmount,
		"Type":              statement.FieldCategory,
	},
	DefaultCurrency: "EUR",
}

func ParseCSV(r io.Reader) ([]statement.Transaction, error) {
	return statement.ParseGenericCSV(r, csvHints)
}

func (N26) ParseCSV(r io.Reader) ([]statement.Transaction, error) { return ParseCSV(r) }
func (N26) ParsePDF(r io.Reader) ([]statement.Transaction, error) { return ParsePDF(r) }
