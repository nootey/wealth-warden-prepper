package revolut

import (
	"strings"
	"testing"

	"github.com/nootey/wealth-warden-prepper/pkg/pdftext"
	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

// Trimmed, fabricated `pdftotext -layout` output.
const pdfFixture = `
EUR Statement

Balance summary

 Product                       Opening balance          Money out          Money in          Closing balance

 Account (Current Account)     €0.00                    €12.44             €150.00           €137.56

 Total                         €0.00                    €12.44             €150.00           €137.56


Account transactions from Jan 1, 2026 to Jan 31, 2026

 Date               Description                                            Money out          Money in          Balance

 Jan 1, 2026         Payment from SAMPLE PERSON                                                €50.00            €50.00
                     Reference: Sent from Revolut
                     From: SAMPLE PERSON, DE000000000000000000


 Jan 3, 2026         SAMPLE MERCHANT                                       €12.44                                €37.56
                     To: SAMPLE MERCHANT, Vilnius
                     Card: 000000******0000


 Jan 5, 2026         Payment from OTHER SAMPLE PERSON                                          €100.00           €137.56
                     Reference: Sent from Revolut
                     From: OTHER SAMPLE PERSON, DE111111111111111111
`

func TestParsePDFText(t *testing.T) {
	txns, err := ParsePDFText(pdftext.SplitLines(pdfFixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 3 {
		t.Fatalf("got %d transactions, want 3", len(txns))
	}

	check := func(i int, dir statement.Direction, amount, cp string) {
		t.Helper()
		x := txns[i]
		if x.Direction != dir || x.Amount.String() != amount || x.Counterparty != cp {
			t.Errorf("txn %d: %+v", i, x)
		}
	}
	check(0, statement.Income, "50", "Payment from SAMPLE PERSON")
	check(1, statement.Expense, "12.44", "SAMPLE MERCHANT")
	check(2, statement.Income, "100", "Payment from OTHER SAMPLE PERSON")

	if txns[0].Date.Format("2006-01-02") != "2026-01-01" || txns[2].Date.Format("2006-01-02") != "2026-01-05" {
		t.Errorf("dates parsed wrong: %v %v", txns[0].Date, txns[2].Date)
	}
	for _, x := range txns {
		if x.Currency != "EUR" {
			t.Errorf("currency %q, want EUR", x.Currency)
		}
	}
}

func TestParsePDFTextIgnoresDetailLines(t *testing.T) {
	txns, err := ParsePDFText(pdftext.SplitLines(pdfFixture))
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range txns {
		if strings.Contains(x.Counterparty, "Reference:") || strings.Contains(x.Counterparty, "From:") || strings.Contains(x.Counterparty, "Card:") {
			t.Errorf("detail line leaked into a transaction: %+v", x)
		}
	}
}
