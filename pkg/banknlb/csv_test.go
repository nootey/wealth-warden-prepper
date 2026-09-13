package banknlb

import (
	"strings"
	"testing"

	"github.com/nootey/wealth-warden-prepper/pkg/statement"
)

const csvHeader = "Opis/Description;Kategorija/Category;+/-;Znesek/Amount;Valuta/Currency;Datum plačila/Value date;Naziv prejemnika/plačnika/Counter party name;Račun prejemnika/plačnika/Counter party Account;Status/Status;BIC koda/BIC Code;Menjalni tečaj/Foreign Exchange rate;Referenca prejemnika/Creditor Reference;Datum poravnave/Settlement date;Dodatni stroški/Additional Charges;Naslov prejemnika/plačnika/Counter party Address;ID transakcije/Transaction ID;Namen/Purpose"

func TestParseCSV(t *testing.T) {
	in := "\uFEFF" + csvHeader + "\r\n" +
		"PLACILO;;+;47,00;EUR;20-12-2017;CVETEK BARBARA;SI56031211000474822;;SKBASI2X;;NRC;20-12-2017;;RIMSKA ULICA 11;TX1;PLACILO\r\n" +
		"NADOMESTILO DB;;-;0,26;EUR;20/12/2017;;;;;;NRC;20/12/2017;;;TX2;NADOMESTILO DB\r\n" +
		"ABOS FITINN;Šport;-;19,9;EUR;20/12/2017;FITINNLjubljana 2;AT692011182114618803;;GIBAATWWXXX;;SI00400415086;20/12/2017;;;TX3;ABOS MNR:4601719\r\n"

	txns, err := ParseCSV(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 3 {
		t.Fatalf("got %d transactions, want 3", len(txns))
	}

	a := txns[0]
	if a.Direction != statement.Income || a.Amount.String() != "47" || a.Counterparty != "CVETEK BARBARA" || a.ExternalID != "TX1" {
		t.Errorf("row 1 parsed wrong: %+v", a)
	}
	if a.Description != "PLACILO" {
		t.Errorf("purpose equal to description should not be repeated, got %q", a.Description)
	}
	if a.Date.Format("2006-01-02") != "2017-12-20" {
		t.Errorf("row 1 date %v", a.Date)
	}

	b := txns[1]
	if b.Direction != statement.Expense || b.Amount.String() != "0.26" || b.Date.Format("2006-01-02") != "2017-12-20" {
		t.Errorf("row 2 parsed wrong: %+v", b)
	}

	c := txns[2]
	if c.Amount.String() != "19.9" || c.BankCategory != "Šport" || c.Description != "ABOS FITINN ABOS MNR:4601719" || c.Reference != "SI00400415086" {
		t.Errorf("row 3 parsed wrong: %+v", c)
	}
}

const csvHeader2025 = "Namen,Kategorija,+/-,Znesek,Valuta,Datum plačila,Naziv prejemnika/plačnika,Naslov prejemnika/plačnika,Račun prejemnika/plačnika,BIC koda,Status,Menjalni tečaj,Referenca prejemnika,Datum poravnave,Dodatni stroški,ID transakcije"

func TestParseCSV2025(t *testing.T) {
	in := "\uFEFF" + csvHeader2025 + "\n" +
		"STR VOD. PAKETA,Finance & Zavarovanja,-,4.49,EUR,31. 12. 2025,,,,,,,,31. 12. 2025,,1643150897\n" +
		"\"ČEVELJCI, ANA Č. 0038640737471\",Prenosi,+,32.00,EUR,5. 3. 2025,ANA Č.,.,.,HDELSI22,,,NRC00,5. 3. 2025,,1630401993\n" +
		"AVTORIZACIJA DM - S021 LJUBLJANA,,-,17.10,EUR,31. 12. 2025,,,,,AVTORIZACIJA,,,31. 12. 2025,,1642229965\n"

	txns, err := ParseCSV(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	if len(txns) != 2 {
		t.Fatalf("got %d transactions, want 2 (pending authorisation skipped)", len(txns))
	}

	a := txns[0]
	if a.Direction != statement.Expense || a.Amount.String() != "4.49" || a.BankCategory != "Finance & Zavarovanja" ||
		a.Description != "STR VOD. PAKETA" || a.ExternalID != "1643150897" || a.Date.Format("2006-01-02") != "2025-12-31" {
		t.Errorf("row 1 parsed wrong: %+v", a)
	}
	b := txns[1]
	if b.Direction != statement.Income || b.Amount.String() != "32" || b.Counterparty != "ANA Č." ||
		b.Description != "ČEVELJCI, ANA Č. 0038640737471" || b.Reference != "NRC00" || b.ExternalID != "1630401993" ||
		b.Date.Format("2006-01-02") != "2025-03-05" {
		t.Errorf("row 2 parsed wrong: %+v", b)
	}
}

func TestParseCSVMissingColumn(t *testing.T) {
	if _, err := ParseCSV(strings.NewReader("a;b;c\n1;2;3\n")); err == nil {
		t.Error("expected error for unknown header")
	}
}
