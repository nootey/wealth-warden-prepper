package bank

import (
	"strings"
	"testing"
)

func TestDetectCSV(t *testing.T) {
	tests := []struct {
		name   string
		header []string
		want   string
	}{
		{
			name:   "nlb",
			header: []string{"Namen", "Kategorija", "+/-", "Znesek", "Valuta", "Datum plačila", "ID transakcije"},
			want:   "nlb",
		},
		{
			name:   "n26",
			header: []string{"Partner Name", "Partner Iban", "Payment Reference", "Amount (EUR)", "Type"},
			want:   "n26",
		},
		{
			name:   "revolut",
			header: []string{"Type", "Product", "Started Date", "Completed Date", "Description", "Amount", "Fee", "Currency", "State", "Balance"},
			want:   "revolut",
		},
		{
			name:   "unrecognized falls back to generic",
			header: []string{"Date", "Description", "Amount"},
			want:   "generic",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := DetectCSV(tc.header); got != tc.want {
				t.Errorf("DetectCSV(%v) = %q, want %q", tc.header, got, tc.want)
			}
		})
	}
}

func TestDetectPDF(t *testing.T) {
	tests := []struct {
		name   string
		lines  []string
		want   string
		wantOK bool
	}{
		{
			name:   "nlb",
			lines:  []string{"20.12.17  SAMPLE PERSON A  SI56000000000000001  +47,00  1.234,56"},
			want:   "nlb",
			wantOK: true,
		},
		{
			name:   "n26",
			lines:  []string{"Sample Payments                                     25.08.2026   +103,00€"},
			want:   "n26",
			wantOK: true,
		},
		{
			name:   "revolut",
			lines:  []string{"Jan 1, 2026         Sample Payment                                                €50.00            €50.00"},
			want:   "revolut",
			wantOK: true,
		},
		{
			name:   "unrecognized layout",
			lines:  []string{"this is not a bank statement line"},
			wantOK: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := DetectPDF(tc.lines)
			if ok != tc.wantOK {
				t.Fatalf("DetectPDF(%v) ok = %v, want %v", tc.lines, ok, tc.wantOK)
			}
			if ok && got != tc.want {
				t.Errorf("DetectPDF(%v) = %q, want %q", tc.lines, got, tc.want)
			}
		})
	}
}

func TestParseAuto(t *testing.T) {
	csv := "Date,Amount\n2026-01-01,10.00\n"

	t.Run("routes .csv to ParseCSVAuto", func(t *testing.T) {
		txns, name, err := ParseAuto(strings.NewReader(csv), ".csv")
		if err != nil {
			t.Fatalf("ParseAuto() error = %v", err)
		}
		if name != "generic" {
			t.Errorf("ParseAuto() name = %q, want %q", name, "generic")
		}
		if len(txns) != 1 {
			t.Errorf("ParseAuto() got %d transactions, want 1", len(txns))
		}
	})

	t.Run("extension match is case-insensitive", func(t *testing.T) {
		if _, _, err := ParseAuto(strings.NewReader(csv), ".CSV"); err != nil {
			t.Fatalf("ParseAuto() error = %v", err)
		}
	})

	t.Run("unsupported extension errors", func(t *testing.T) {
		if _, _, err := ParseAuto(strings.NewReader(csv), ".txt"); err == nil {
			t.Fatal("ParseAuto() error = nil, want error for unsupported extension")
		}
	})
}
