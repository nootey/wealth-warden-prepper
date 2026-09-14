package revolut

import (
	"strings"
	"testing"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

const csvHeader = "Type,Product,Started Date,Completed Date,Description,Amount,Fee,Currency,State,Balance"

func TestParseCSV(t *testing.T) {
	in := csvHeader + "\n" +
		"Topup,Current,2026-01-01 08:00:00,2026-01-01 08:00:01,Payment from SAMPLE PERSON,50.00,0.00,EUR,COMPLETED,50.00\n" +
		"Card Payment,Current,2026-01-02 10:00:00,2026-01-03 12:00:00,SAMPLE MERCHANT,-12.44,0.00,EUR,COMPLETED,37.56\n" +
		"Card Payment,Current,2026-01-04 09:00:00,,SAMPLE MERCHANT,-11.39,0.00,EUR,REVERTED,\n"

	txns, err := ParseCSV(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 2 {
		t.Fatalf("got %d transactions, want 2 (reverted payment skipped)", len(txns))
	}

	a := txns[0]
	if a.Direction != statement.Income || a.Amount.String() != "50" || a.Description != "Payment from SAMPLE PERSON" {
		t.Errorf("row 1 parsed wrong: %+v", a)
	}
	if a.Date.Format("2006-01-02") != "2026-01-01" {
		t.Errorf("row 1 should use the completed date, got %v", a.Date)
	}

	b := txns[1]
	if b.Direction != statement.Expense || b.Amount.String() != "12.44" || b.Description != "SAMPLE MERCHANT" {
		t.Errorf("row 2 parsed wrong: %+v", b)
	}
	if b.Date.Format("2006-01-02") != "2026-01-03" {
		t.Errorf("row 2 should use the completed date, not the started date, got %v", b.Date)
	}
}

func TestDetectCSV(t *testing.T) {
	header := strings.Split(csvHeader, ",")
	if score := (Revolut{}).DetectCSV(header); score == 0 {
		t.Errorf("DetectCSV(%v) = 0, want > 0", header)
	}
}
