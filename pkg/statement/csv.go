package statement

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

type Field string

const (
	FieldDescription  Field = "description"
	FieldCategory     Field = "category"
	FieldSign         Field = "sign"
	FieldAmount       Field = "amount"
	FieldCurrency     Field = "currency"
	FieldValueDate    Field = "value_date"
	FieldCounterparty Field = "counterparty"
	FieldAccount      Field = "account"
	FieldReference    Field = "reference"
	FieldSettleDate   Field = "settle_date"
	FieldTxnID        Field = "txn_id"
	FieldPurpose      Field = "purpose"
	FieldStatus       Field = "status"
	FieldBalance      Field = "balance"
)

// Common column names seen across bank CSV exports, keyed by their
// normalized (lowercased, bilingual-tail-stripped) form.
var baseAliases = map[string]Field{
	"description":           FieldDescription,
	"details":               FieldDescription,
	"narrative":             FieldDescription,
	"category":              FieldCategory,
	"amount":                FieldAmount,
	"currency":              FieldCurrency,
	"date":                  FieldValueDate,
	"value date":            FieldValueDate,
	"booking date":          FieldValueDate,
	"transaction date":      FieldValueDate,
	"completed date":        FieldValueDate,
	"settlement date":       FieldSettleDate,
	"started date":          FieldSettleDate,
	"counterparty":          FieldCounterparty,
	"counter party name":    FieldCounterparty,
	"payee":                 FieldCounterparty,
	"merchant":              FieldCounterparty,
	"account":               FieldAccount,
	"counter party account": FieldAccount,
	"iban":                  FieldAccount,
	"reference":             FieldReference,
	"creditor reference":    FieldReference,
	"transaction id":        FieldTxnID,
	"purpose":               FieldPurpose,
	"status":                FieldStatus,
	"state":                 FieldStatus,
	"balance":               FieldBalance,
}

// CSVHints lets a bank-specific rule set fill in what the generic engine
// cannot guess on its own.
type CSVHints struct {
	// Aliases maps a raw, untrimmed-of-slashes header exactly as it appears
	// in the file to a Field, checked before the base dictionary.
	Aliases map[string]Field
	// Signature lists raw header names that identify this bank for
	// auto-detection. When nil, detection falls back to Aliases' keys; set
	// this explicitly once a bank's real fingerprint headers have moved
	// into the base dictionary and are no longer in its own Aliases.
	Signature            []string
	DateLayouts          []string
	DefaultCurrency      string
	SkipStatusValues     []string
	MergePurposeIntoDesc bool
}

func normalizeHeader(h string) string {
	if i := strings.LastIndex(h, "/"); i >= 0 {
		if tail := strings.TrimSpace(h[i+1:]); tail != "" {
			h = tail
		}
	}
	return strings.ToLower(strings.TrimSpace(h))
}

func resolveField(rawHeader string, extra map[string]Field) (Field, bool) {
	h := strings.TrimSpace(strings.TrimPrefix(rawHeader, "\uFEFF"))
	if h == "+/-" {
		return FieldSign, true
	}
	if f, ok := extra[h]; ok {
		return f, true
	}
	if f, ok := baseAliases[normalizeHeader(h)]; ok {
		return f, true
	}
	return "", false
}

func sniffDelimiter(raw []byte) rune {
	header, _, _ := bytes.Cut(raw, []byte("\n"))
	best, bestCount := ',', -1
	for _, c := range []rune{';', ',', '\t', '|'} {
		if n := bytes.Count(header, []byte(string(c))); n > bestCount {
			best, bestCount = c, n
		}
	}
	return best
}

// DetectCSVScore counts how many header cells match hints' Signature
// (falling back to Aliases' keys when Signature is nil).
func DetectCSVScore(hints CSVHints, header []string) int {
	sig := hints.Signature
	if sig == nil {
		for k := range hints.Aliases {
			sig = append(sig, k)
		}
	}
	set := make(map[string]bool, len(sig))
	for _, s := range sig {
		set[s] = true
	}
	n := 0
	for _, h := range header {
		key := strings.TrimSpace(strings.TrimPrefix(h, "\uFEFF"))
		if set[key] {
			n++
		}
	}
	return n
}

func PeekCSVHeader(raw []byte) ([]string, error) {
	raw = bytes.TrimPrefix(raw, []byte("\uFEFF"))
	cr := csv.NewReader(bytes.NewReader(raw))
	cr.Comma = sniffDelimiter(raw)
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true
	return cr.Read()
}

// ParseGenericCSV parses a tabular bank export on a best-effort basis: it
// sniffs the delimiter and maps headers via the base dictionary plus any
// bank-specific hints. When no sign column is present, direction is
// inferred from the amount's own sign.
func ParseGenericCSV(r io.Reader, hints CSVHints) ([]Transaction, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimPrefix(raw, []byte("\uFEFF"))

	cr := csv.NewReader(bytes.NewReader(raw))
	cr.Comma = sniffDelimiter(raw)
	cr.LazyQuotes = true
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true

	rows, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read csv: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("csv is empty")
	}

	idx := map[Field]int{}
	for i, h := range rows[0] {
		if f, ok := resolveField(h, hints.Aliases); ok {
			if _, exists := idx[f]; !exists {
				idx[f] = i
			}
		}
	}
	if _, ok := idx[FieldAmount]; !ok {
		return nil, fmt.Errorf("csv header is missing an amount column")
	}
	_, hasValueDate := idx[FieldValueDate]
	_, hasSettleDate := idx[FieldSettleDate]
	if !hasValueDate && !hasSettleDate {
		return nil, fmt.Errorf("csv header is missing a date column")
	}

	get := func(row []string, f Field) string {
		i, ok := idx[f]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	var out []Transaction
	for n, row := range rows[1:] {
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}
		line := n + 2

		skip := false
		if status := get(row, FieldStatus); status != "" {
			for _, s := range hints.SkipStatusValues {
				if strings.EqualFold(status, s) {
					skip = true
					break
				}
			}
		}
		if skip {
			continue
		}

		amount, err := ParseAmount(get(row, FieldAmount))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}

		dateStr := get(row, FieldValueDate)
		if dateStr == "" {
			dateStr = get(row, FieldSettleDate)
		}
		date, err := ParseDate(dateStr, hints.DateLayouts...)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}

		var dir Direction
		if _, ok := idx[FieldSign]; ok {
			switch get(row, FieldSign) {
			case "+":
				dir = Income
			case "-":
				dir = Expense
			default:
				return nil, fmt.Errorf("line %d: unknown sign %q", line, get(row, FieldSign))
			}
		} else if amount.IsNegative() {
			dir = Expense
		} else {
			dir = Income
		}
		if amount.IsNegative() {
			amount = amount.Abs()
		}

		cur := get(row, FieldCurrency)
		if cur == "" {
			cur = hints.DefaultCurrency
		}

		desc := get(row, FieldDescription)
		if hints.MergePurposeIntoDesc {
			if p := get(row, FieldPurpose); p != "" && !strings.EqualFold(p, desc) {
				desc = strings.TrimSpace(desc + " " + p)
			}
		}

		out = append(out, Transaction{
			Date:         date,
			Direction:    dir,
			Amount:       amount,
			Currency:     cur,
			Counterparty: get(row, FieldCounterparty),
			Account:      get(row, FieldAccount),
			Description:  strings.TrimSpace(desc),
			Reference:    get(row, FieldReference),
			BankCategory: get(row, FieldCategory),
			ExternalID:   get(row, FieldTxnID),
		})
	}
	return out, nil
}
