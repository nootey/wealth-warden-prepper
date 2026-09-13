package format

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
	"github.com/shopspring/decimal"
)

func TestBuild(t *testing.T) {
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	txns := []statement.Transaction{
		{Date: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Direction: statement.Income, Amount: decimal.RequireFromString("869.2"), Currency: "EUR", Counterparty: "ACME", Description: "PLAČA"},
		{Date: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC), Direction: statement.Expense, Amount: decimal.RequireFromString("5"), Currency: "EUR", Counterparty: "PROVIZIJA", Description: "PROVIZIJA"},
	}
	p := Build("nlb_2025", txns, now)
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"identifier":"nlb_2025","generated_at":"2026-09-13T12:00:00Z","transactions":[` +
		`{"transaction_type":"income","amount":"869.20","currency":"EUR","txn_date":"2025-01-01T00:00:00Z","category":"(uncategorized)","description":"ACME - PLAČA"},` +
		`{"transaction_type":"expense","amount":"5.00","currency":"EUR","txn_date":"2025-01-02T00:00:00Z","category":"(uncategorized)","description":"PROVIZIJA"}]}`
	if string(b) != want {
		t.Errorf("got\n%s\nwant\n%s", b, want)
	}
}
