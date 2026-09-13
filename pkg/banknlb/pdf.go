package banknlb

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/nootey/wealth-warden-prepper/pkg/pdftext"
	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

// dd.mm.yy   COUNTERPARTY   [ACCOUNT]   ±1.234,56   1.234,56[-]
var (
	pdfTxnRe = regexp.MustCompile(
		`^\s*(\d{2}\.\d{2}\.\d{2})\s+(.*?)\s+([+-]\d{1,3}(?:\.\d{3})*,\d{2})\s+(\d{1,3}(?:\.\d{3})*,\d{2}-?)\s*$`,
	)
	colSplitRe = regexp.MustCompile(`\s{2,}`)

	// Page chrome ends the current transaction.
	pdfChromePrefixes = []string{
		"št. izpiska", "æt. izpiska", "stran", "datum izpiska", "obvestilo",
		"naziv nalogodajalca", "datum ", "referenčna", "eur - evro",
		"stanje predhodnega", "skupni promet", "novo stanje", "denarna", "izpisek",
	}
)

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

// Takes `pdftotext -layout` lines.
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
		if trimmed == "" || strings.Contains(line, "\f") || isChrome(trimmed) {
			flush()
			continue
		}

		if m := pdfTxnRe.FindStringSubmatch(line); m != nil {
			flush()
			date, err := parseDate(m[1])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", n+1, err)
			}
			amount, err := parseAmount(m[3])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", n+1, err)
			}
			balance, err := parseAmount(m[4])
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", n+1, err)
			}
			dir := statement.Income
			if amount.IsNegative() {
				dir = statement.Expense
				amount = amount.Abs()
			}
			cols := colSplitRe.Split(strings.TrimSpace(m[2]), -1)
			t := statement.Transaction{
				Date:         date,
				Direction:    dir,
				Amount:       amount,
				Currency:     "EUR",
				Counterparty: cols[0],
				Balance:      &balance,
			}
			if len(cols) > 1 && cols[1] != "." {
				t.Account = cols[1]
			}
			cur = &t
			continue
		}

		if cur != nil {
			cols := colSplitRe.Split(trimmed, -1)
			purpose = append(purpose, cols[0])
			if len(cols) > 1 && cur.Reference == "" {
				cur.Reference = cols[1]
			}
		}
	}
	flush()
	return out, nil
}

func isChrome(trimmed string) bool {
	l := strings.ToLower(trimmed)
	for _, p := range pdfChromePrefixes {
		if strings.HasPrefix(l, p) {
			return true
		}
	}
	return false
}
