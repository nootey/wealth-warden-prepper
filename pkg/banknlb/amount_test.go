package banknlb

import (
	"testing"
	"time"
)

func TestParseAmount(t *testing.T) {
	cases := map[string]string{
		"47,00": "47", "1": "1", "19,9": "19.9", "1.658,83": "1658.83",
		"+2.513,52": "2513.52", "-757,35": "-757.35", "0,87-": "-0.87",
		"4.49": "4.49", "1,234.56": "1234.56",
	}
	for in, want := range cases {
		got, err := parseAmount(in)
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if got.String() != want {
			t.Errorf("%q: got %s want %s", in, got, want)
		}
	}
	if _, err := parseAmount("abc"); err == nil {
		t.Error("expected error for abc")
	}
}

func TestParseDate(t *testing.T) {
	want := time.Date(2017, 12, 20, 0, 0, 0, 0, time.UTC)
	for _, in := range []string{"20-12-2017", "20/12/2017", "20.12.2017", "20.12.17", "20. 12. 2017"} {
		got, err := parseDate(in)
		if err != nil || !got.Equal(want) {
			t.Errorf("%q: got %v err %v", in, got, err)
		}
	}
}
