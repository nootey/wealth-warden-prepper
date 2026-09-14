package statement

import (
	"io"
	"time"

	"github.com/shopspring/decimal"
)

type Direction string

const (
	Income  Direction = "income"
	Expense Direction = "expense"
)

type Transaction struct {
	Date         time.Time
	Direction    Direction
	Amount       decimal.Decimal // always positive; Direction carries the sign
	Currency     string
	Counterparty string
	Account      string // counterparty account or IBAN, when the bank gives one
	Description  string
	Reference    string
	BankCategory string // category text the bank itself assigned, if any
	ExternalID   string // bank-side transaction ID, used for de-duplication
	Balance      *decimal.Decimal
}

type Parser interface {
	ParseCSV(r io.Reader) ([]Transaction, error)
	ParsePDF(r io.Reader) ([]Transaction, error)
}

// Detector scores: higher is a better match, 0 means no match.
type Detector interface {
	DetectCSV(header []string) int
	DetectPDF(lines []string) int
}

type LinesParser interface {
	ParsePDFLines(lines []string) ([]Transaction, error)
}

// Keeps the first of each ExternalID.
func Dedupe(txns []Transaction) []Transaction {
	seen := make(map[string]bool, len(txns))
	out := make([]Transaction, 0, len(txns))
	for _, t := range txns {
		if t.ExternalID != "" {
			if seen[t.ExternalID] {
				continue
			}
			seen[t.ExternalID] = true
		}
		out = append(out, t)
	}
	return out
}
