package banknlb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nootey/wealth-warden-prepper/pkg/pdftext"
	"github.com/nootey/wealth-warden-prepper/pkg/statement"
	"github.com/shopspring/decimal"
)

// Trimmed, anonymised output of `pdftotext -layout` for a two-page statement.
const pdfFixture = `
Datum          Naziv nalogodajalca/prejemnika                                    Račun prejemnika                                   Promet v dobro/breme                           Stanje
               Namen/opis spremembe                                              Referenčna številka
EUR - EVRO

               STANJE PREDHODNEGA IZPISKA                                  1.658,83
               SKUPNI PROMET V DOBRO                                       4.413,11
               NOVO STANJE                                                 2.148,17

29.07.26       JP LPP                                                                                                                                     -1,50                   1.657,33
01.08.26       JANE DOE                                                       DE97 1001 1001 2292 2644 32                                               +0,10                   1.657,43
               SENT FROM N26
02.08.26       TANISA P.                                                        .                                                                       +97,00                    1.754,43
               TANISA P. 0038670573277
02.11.16       PROVIZIJA                                                                                                                                  -5,00                       0,87-
20.08.26       CF Fitness d.o.o.                       AT10 3910 0000 0088 5418                               -39,90                                                             2.341,60
               LJR--0021-0001347 BASIC 39.90 EUR         LJR--0021-0001347
               01.08.26 - 31.08.26
24.08.26       MRKONJIĆ A.                               .                                                     +25,00                                                             1.852,62
               ., MRKONJIĆ A.
                                                                                Št. izpiska 08
                                                                                Stran 03

Obvestilo o spremembi splošnih pogojev poslovanja
Obveščamo vas, da bodo s 1. januarjem 2017 nekaj.
`

func TestParsePDFText(t *testing.T) {
	txns, err := ParsePDFText(pdftext.SplitLines(pdfFixture))
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 6 {
		t.Fatalf("got %d transactions, want 6", len(txns))
	}

	check := func(i int, dir statement.Direction, amount, cp, account, desc, ref, balance string) {
		t.Helper()
		x := txns[i]
		if x.Direction != dir || x.Amount.String() != amount || x.Counterparty != cp || x.Account != account ||
			x.Description != desc || x.Reference != ref || x.Balance == nil || x.Balance.String() != balance {
			t.Errorf("txn %d: %+v (balance %v)", i, x, x.Balance)
		}
	}
	check(0, statement.Expense, "1.5", "JP LPP", "", "", "", "1657.33")
	check(1, statement.Income, "0.1", "JANE DOE", "DE97 1001 1001 2292 2644 32", "SENT FROM N26", "", "1657.43")
	check(2, statement.Income, "97", "TANISA P.", "", "TANISA P. 0038670573277", "", "1754.43")
	check(3, statement.Expense, "5", "PROVIZIJA", "", "", "", "-0.87")
	check(4, statement.Expense, "39.9", "CF Fitness d.o.o.", "AT10 3910 0000 0088 5418",
		"LJR--0021-0001347 BASIC 39.90 EUR 01.08.26 - 31.08.26", "LJR--0021-0001347", "2341.6")
	check(5, statement.Income, "25", "MRKONJIĆ A.", "", "., MRKONJIĆ A.", "", "1852.62")

	if txns[3].Date.Year() != 2016 || txns[0].Date.Year() != 2026 {
		t.Errorf("two-digit years not expanded: %v %v", txns[3].Date, txns[0].Date)
	}
}

// Sample statements live in data/input/nlb and are not committed. When they
// exist and pdftotext is installed, check the parsed sums against the totals
// printed on the statement itself.
func TestParsePDFSamples(t *testing.T) {
	path := filepath.Join("..", "..", "data", "input", "nlb", "Izpisek_2026_09_13_14_11_00.pdf")
	if _, err := os.Stat(path); err != nil {
		t.Skip("sample statement not present")
	}
	txns, err := ParsePDFFile(path)
	if err == pdftext.ErrNotInstalled {
		t.Skip(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	in, out := decimal.Zero, decimal.Zero
	for _, x := range txns {
		if x.Direction == statement.Income {
			in = in.Add(x.Amount)
		} else {
			out = out.Add(x.Amount)
		}
	}
	if len(txns) != 94 || in.String() != "4413.11" || out.String() != "3923.77" {
		t.Errorf("got %d txns, in %s, out %s", len(txns), in, out)
	}
}
