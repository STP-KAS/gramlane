package names

import "testing"

func TestParseWantBakery(t *testing.T) {
	w := ParseWant("Bakery")
	if !w.Valid || w.Name != "bakery.kas" || w.PriceKAS != 35 {
		t.Fatalf("%+v", w)
	}
	if w.Payload != `{"op":"create","p":"domain","v":"bakery"}` {
		t.Fatal(w.Payload)
	}
}

func TestParseWantRejectsJunk(t *testing.T) {
	if ParseWant("").Valid {
		t.Fatal("empty")
	}
	if ParseWant("hello world").Valid {
		t.Fatal("space")
	}
	if ParseWant("A_B").Valid {
		t.Fatal("underscore")
	}
	if ParseWant("opus..dei").Valid {
		t.Fatal("double dot")
	}
}

func TestParseWantDotted(t *testing.T) {
	w := ParseWant("opus.dei")
	if !w.Valid || w.Label != "opus.dei" || w.Name != "opus.dei.kas" {
		t.Fatalf("%+v", w)
	}
	if ParseWant("opus.dei.kas").Name != "opus.dei.kas" {
		t.Fatal(ParseWant("opus.dei.kas").Name)
	}
}

func TestPriceKASTiers(t *testing.T) {
	if PriceKAS("a") != 4200 || PriceKAS("ab") != 4200 {
		t.Fatal("short")
	}
	if PriceKAS("abc") != 2100 || PriceKAS("abcd") != 525 {
		t.Fatal("mid")
	}
	if PriceKAS("abcde") != 35 {
		t.Fatal("long")
	}
	if PriceKAS("opus.dei") != 35 {
		t.Fatal("dotted")
	}
}
