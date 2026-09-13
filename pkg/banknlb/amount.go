package banknlb

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// "47,00", "1", "19,9", "1.658,83", "0,87-"
func parseAmount(s string) (decimal.Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return decimal.Zero, fmt.Errorf("empty amount")
	}
	neg := false
	if strings.HasSuffix(s, "-") {
		neg = true
		s = strings.TrimSuffix(s, "-")
	}
	if strings.HasPrefix(s, "-") {
		neg = !neg
		s = strings.TrimPrefix(s, "-")
	}
	s = strings.TrimPrefix(s, "+")
	if strings.Contains(s, ",") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero, fmt.Errorf("amount %q: %w", s, err)
	}
	if neg {
		d = d.Neg()
	}
	return d, nil
}

var dateLayouts = []string{"02-01-2006", "02/01/2006", "02.01.2006", "02.01.06"}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, l := range dateLayouts {
		if t, err := time.ParseInLocation(l, s, time.UTC); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date %q", s)
}
