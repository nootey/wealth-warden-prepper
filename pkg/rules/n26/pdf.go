package n26

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/nootey/wealth-warden-prepper/pkg/pdftext"
	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

// <description>   dd.mm.yyyy   ±1.234,56€
var (
	pdfTxnRe = regexp.MustCompile(
		`^\s*(.+?)\s+(\d{2}\.\d{2}\.\d{4})\s+([+-]\d{1,3}(?:\.\d{3})*,\d{2})€\s*$`,
	)
	ibanRe = regexp.MustCompile(`^IBAN:\s*(\S+)`)
)

func (N26) DetectPDF(lines []string) int {
	n := 0
	for _, l := range lines {
		if pdfTxnRe.MatchString(l) {
			n++
		}
	}
	return n
}

func (N26) ParsePDFLines(lines []string) ([]statement.Transaction, error) { return ParsePDFText(lines) }

func ParsePDF(r io.Reader) ([]statement.Transaction, error) {
	lines, err := pdftext.ExtractReader(r)
	if err != nil {
		return nil, err
	}
	return ParsePDFText(lines)
}

func ParsePDFFile(path string) ([]statement.Transaction, error) {
	lines, err := pdftext.Extract(path)
	if err != nil {
		return nil, err
	}
	return ParsePDFText(lines)
}

// Takes `pdftotext -layout` lines. A blank line ends the current
// transaction, which is enough to separate transactions from the repeated
// page header/footer, since one always precedes the other.
func ParsePDFText(lines []string) ([]statement.Transaction, error) {
	var (
		out     []statement.Transaction
		cur     *statement.Transaction
		purpose []string
	)
	flush := func() {
		if cur == nil {
			return
		}
		cur.Description = strings.TrimSpace(strings.Join(purpose, " "))
		out = append(out, *cur)
		cur = nil
		purpose = nil
	}

	for n, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.Contains(line, "\f") {
			flush()
			continue
		}

		if m := pdfTxnRe.FindStringSubmatch(line); m != nil {
			flush()
			date, err := statement.ParseDate(m[2])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", n+1, err)
			}
			amount, err := statement.ParseAmount(m[3])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", n+1, err)
			}
			dir := statement.Income
			if amount.IsNegative() {
				dir = statement.Expense
				amount = amount.Abs()
			}
			cur = &statement.Transaction{
				Date:         date,
				Direction:    dir,
				Amount:       amount,
				Currency:     "EUR",
				Counterparty: strings.TrimSpace(m[1]),
			}
			continue
		}

		if cur == nil {
			continue
		}
		if strings.HasPrefix(trimmed, "Value Date") {
			continue
		}
		if m := ibanRe.FindStringSubmatch(trimmed); m != nil {
			cur.Account = m[1]
			continue
		}
		purpose = append(purpose, trimmed)
	}
	flush()
	return out, nil
}
