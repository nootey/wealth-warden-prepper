package n26

import (
	"strings"
	"testing"

	"github.com/nootey/wealth-warden-prepper/pkg/pdftext"
	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

// Trimmed, fabricated `pdftotext -layout` output for a two-page statement.
const pdfFixture = `
Bank Statement Nr. 08/2026
01.08.2026 until 31.08.2026



Description                                       Booking Date     Amount
From Little piggie                                  01.08.2026      +0,10€
Value Date 01.08.2026

ACCOUNT HOLDER                                        01.08.2026       -0,10€
Outgoing Transfers
IBAN: SI56000000000000001 • BIC: TESTSI2XXXX
Sent from N26
Value Date 01.08.2026

ACCOUNT HOLDER                                        14.08.2026   +1.000,00€
Income
IBAN: SI56000000000000001 • BIC: TESTSI2X
DEPOSIT
Value Date 14.08.2026




ACCOUNT HOLDER                                                      Issued on
Sample Street 1, 1000 Sample City                                 14.09.2026
IBAN: DE00000000000000000000 • BIC: TESTDEB1XXX                  Nr. 08/2026
                                                                         1/7
Bank Statement Nr. 08/2026
01.08.2026 until 31.08.2026



Description                                       Booking Date     Amount
Sample Payments                                     25.08.2026   +103,00€
Income
IBAN: LU000000000000000000000 • BIC: TESTLUL1XXX
Vinted
Value Date 25.08.2026
`

func TestParsePDFText(t *testing.T) {
	txns, err := ParsePDFText(pdftext.SplitLines(pdfFixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 4 {
		t.Fatalf("got %d transactions, want 4", len(txns))
	}

	check := func(i int, dir statement.Direction, amount, cp, account, desc string) {
		t.Helper()
		x := txns[i]
		if x.Direction != dir || x.Amount.String() != amount || x.Counterparty != cp || x.Account != account || x.Description != desc {
			t.Errorf("txn %d: %+v", i, x)
		}
	}
	check(0, statement.Income, "0.1", "From Little piggie", "", "")
	check(1, statement.Expense, "0.1", "ACCOUNT HOLDER", "SI56000000000000001", "Outgoing Transfers Sent from N26")
	check(2, statement.Income, "1000", "ACCOUNT HOLDER", "SI56000000000000001", "Income DEPOSIT")
	check(3, statement.Income, "103", "Sample Payments", "LU000000000000000000000", "Income Vinted")

	if txns[0].Date.Format("2006-01-02") != "2026-08-01" || txns[3].Date.Format("2006-01-02") != "2026-08-25" {
		t.Errorf("dates parsed wrong: %v %v", txns[0].Date, txns[3].Date)
	}
	for _, x := range txns {
		if x.Currency != "EUR" {
			t.Errorf("currency %q, want EUR", x.Currency)
		}
	}
}

func TestParsePDFTextIgnoresPageChrome(t *testing.T) {
	txns, err := ParsePDFText(pdftext.SplitLines(pdfFixture))
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range txns {
		if x.Counterparty == "Description" || x.Counterparty == "Sample Street 1, 1000 Sample City" ||
			strings.Contains(x.Description, "Issued on") || strings.Contains(x.Description, "Nr. 08/2026") {
			t.Errorf("page chrome leaked into a transaction: %+v", x)
		}
	}
}
