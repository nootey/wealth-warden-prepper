package banknlb

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

// Headers are bilingual; match the English tail after the last slash.
const (
	colDescription  = "description"
	colCategory     = "category"
	colSign         = "+/-"
	colAmount       = "amount"
	colCurrency     = "currency"
	colValueDate    = "value date"
	colCounterparty = "counter party name"
	colAccount      = "counter party account"
	colReference    = "creditor reference"
	colSettleDate   = "settlement date"
	colTxnID        = "transaction id"
	colPurpose      = "purpose"
	colStatus       = "status"
)

// Exports since 2021 are comma-separated with Slovene-only headers.
var slHeaders = map[string]string{
	"Namen":                     colDescription,
	"Kategorija":                colCategory,
	"Znesek":                    colAmount,
	"Valuta":                    colCurrency,
	"Datum plačila":             colValueDate,
	"Naziv prejemnika/plačnika": colCounterparty,
	"Račun prejemnika/plačnika": colAccount,
	"Referenca prejemnika":      colReference,
	"Datum poravnave":           colSettleDate,
	"ID transakcije":            colTxnID,
}

// Card authorisations that are not booked yet; they show up again once booked.
const statusPending = "AVTORIZACIJA"

func headerKey(h string) string {
	h = strings.TrimSpace(strings.TrimPrefix(h, "\uFEFF"))
	if h == colSign {
		return colSign
	}
	if k, ok := slHeaders[h]; ok {
		return k
	}
	if i := strings.LastIndex(h, "/"); i >= 0 {
		h = h[i+1:]
	}
	return strings.ToLower(strings.TrimSpace(h))
}

func ParseCSV(r io.Reader) ([]statement.Transaction, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimPrefix(raw, []byte("\uFEFF"))

	cr := csv.NewReader(bytes.NewReader(raw))
	cr.Comma = ';'
	if header, _, _ := bytes.Cut(raw, []byte("\n")); !bytes.Contains(header, []byte(";")) {
		cr.Comma = ','
	}
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

	idx := map[string]int{}
	for i, h := range rows[0] {
		idx[headerKey(h)] = i
	}
	for _, need := range []string{colAmount, colSign, colValueDate, colDescription} {
		if _, ok := idx[need]; !ok {
			return nil, fmt.Errorf("csv header is missing the %q column", need)
		}
	}
	get := func(row []string, key string) string {
		i, ok := idx[key]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	var out []statement.Transaction
	for n, row := range rows[1:] {
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}
		line := n + 2
		if strings.EqualFold(get(row, colStatus), statusPending) {
			continue
		}

		amount, err := parseAmount(get(row, colAmount))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		dateStr := get(row, colValueDate)
		if dateStr == "" {
			dateStr = get(row, colSettleDate)
		}
		date, err := parseDate(dateStr)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}

		var dir statement.Direction
		switch get(row, colSign) {
		case "+":
			dir = statement.Income
		case "-":
			dir = statement.Expense
		default:
			return nil, fmt.Errorf("line %d: unknown sign %q", line, get(row, colSign))
		}
		if amount.IsNegative() {
			amount = amount.Abs()
		}

		cur := get(row, colCurrency)
		if cur == "" {
			cur = "EUR"
		}

		desc := get(row, colDescription)
		if p := get(row, colPurpose); p != "" && !strings.EqualFold(p, desc) {
			desc = desc + " " + p
		}

		out = append(out, statement.Transaction{
			Date:         date,
			Direction:    dir,
			Amount:       amount,
			Currency:     cur,
			Counterparty: get(row, colCounterparty),
			Account:      get(row, colAccount),
			Description:  strings.TrimSpace(desc),
			Reference:    get(row, colReference),
			BankCategory: get(row, colCategory),
			ExternalID:   get(row, colTxnID),
		})
	}
	return out, nil
}
