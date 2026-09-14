package revolut

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/nootey/wealth-warden-prepper/pkg/pdftext"
	"github.com/nootey/wealth-warden-prepper/pkg/statement"
	"github.com/shopspring/decimal"
)

// <Mon D, YYYY>   <description>   [€amount]   €balance
var (
	pdfTxnRe     = regexp.MustCompile(`^\s*([A-Za-z]{3} \d{1,2}, \d{4})\s{2,}(.+?)\s*$`)
	colSplitRe   = regexp.MustCompile(`\s{2,}`)
	moneyRe      = regexp.MustCompile(`^€([\d,]+\.\d{2})$`)
	totalRowRe   = regexp.MustCompile(`^Total\s`)
	moneyTokenRe = regexp.MustCompile(`€[\d,]+\.\d{2}`)
)

func ParsePDF(r io.Reader) ([]statement.Transaction, error) {
	lines, err := pdftext.ExtractReader(r)
	if err != nil {
		return nil, err
	}
	return ParsePDFText(lines)
}

func (Revolut) ParsePDFLines(lines []string) ([]statement.Transaction, error) {
	return ParsePDFText(lines)
}

func (Revolut) DetectPDF(lines []string) int {
	n := 0
	for _, l := range lines {
		if pdfTxnRe.MatchString(l) {
			n++
		}
	}
	return n
}

// Takes `pdftotext -layout` lines. Revolut prints an unsigned amount under
// either a "Money out" or "Money in" column, and whitespace-collapsing loses
// which one; direction is inferred instead from whether the running balance
// (also printed per line) rose or fell, seeded from the statement's own
// printed opening balance.
func ParsePDFText(lines []string) ([]statement.Transaction, error) {
	balance, err := openingBalance(lines)
	if err != nil {
		return nil, err
	}

	var out []statement.Transaction
	for n, line := range lines {
		m := pdfTxnRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		cols := colSplitRe.Split(m[2], -1)
		if len(cols) < 2 {
			continue
		}
		amountM := moneyRe.FindStringSubmatch(cols[len(cols)-2])
		balanceM := moneyRe.FindStringSubmatch(cols[len(cols)-1])
		if amountM == nil || balanceM == nil {
			continue
		}

		date, err := statement.ParseDate(m[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", n+1, err)
		}
		amount, err := statement.ParseAmount(amountM[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", n+1, err)
		}
		newBalance, err := statement.ParseAmount(balanceM[1])
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", n+1, err)
		}

		dir := statement.Income
		if newBalance.LessThan(balance) {
			dir = statement.Expense
		}
		balance = newBalance

		out = append(out, statement.Transaction{
			Date:         date,
			Direction:    dir,
			Amount:       amount,
			Currency:     "EUR",
			Counterparty: strings.TrimSpace(strings.Join(cols[:len(cols)-2], " ")),
		})
	}
	return out, nil
}

func openingBalance(lines []string) (decimal.Decimal, error) {
	for _, line := range lines {
		if !totalRowRe.MatchString(strings.TrimSpace(line)) {
			continue
		}
		tokens := moneyTokenRe.FindAllString(line, -1)
		if len(tokens) == 0 {
			break
		}
		return statement.ParseAmount(strings.TrimPrefix(tokens[0], "€"))
	}
	return decimal.Zero, fmt.Errorf("could not find opening balance in statement")
}
