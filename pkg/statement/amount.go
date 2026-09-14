package statement

import (
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// "47,00", "1", "19,9", "1.658,83", "0,87-", "4.49", "1,234.56"
func ParseAmount(s string) (decimal.Decimal, error) {
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
	// The last separator is the decimal one.
	if strings.LastIndex(s, ",") > strings.LastIndex(s, ".") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	} else {
		s = strings.ReplaceAll(s, ",", "")
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

var DateLayouts = []string{
	"02-01-2006", "02/01/2006", "02.01.2006", "02.01.06", "2. 1. 2006", "2006-01-02",
	"2006-01-02 15:04:05", "Jan 2, 2006",
}

func ParseDate(s string, extra ...string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, l := range DateLayouts {
		if t, err := time.ParseInLocation(l, s, time.UTC); err == nil {
			return t, nil
		}
	}
	for _, l := range extra {
		if t, err := time.ParseInLocation(l, s, time.UTC); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date %q", s)
}
