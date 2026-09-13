package format

import (
	"strings"
	"time"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

const Uncategorized = "(uncategorized)"

type Txn struct {
	TransactionType string `json:"transaction_type"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
	TxnDate         string `json:"txn_date"`
	Category        string `json:"category"`
	Description     string `json:"description"`
	ExternalID      string `json:"external_txn_id"` // empty when the statement gives none (PDF)
}

type Payload struct {
	Identifier   string `json:"identifier"`
	GeneratedAt  string `json:"generated_at"`
	Transactions []Txn  `json:"transactions"`
}

func Build(identifier string, txns []statement.Transaction, now time.Time) Payload {
	out := Payload{
		Identifier:   identifier,
		GeneratedAt:  now.UTC().Format(time.RFC3339Nano),
		Transactions: make([]Txn, 0, len(txns)),
	}
	for _, t := range txns {
		desc := t.Counterparty
		if t.Description != "" && !strings.EqualFold(t.Description, t.Counterparty) {
			if desc != "" {
				desc += " - "
			}
			desc += t.Description
		}
		out.Transactions = append(out.Transactions, Txn{
			TransactionType: string(t.Direction),
			Amount:          t.Amount.StringFixed(2),
			Currency:        t.Currency,
			TxnDate:         t.Date.UTC().Format("2006-01-02T15:04:05Z"),
			Category:        Uncategorized,
			Description:     desc,
			ExternalID:      t.ExternalID,
		})
	}
	return out
}
