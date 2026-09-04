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
	if ParseWant("pay.bakery").Valid {
		t.Fatal("subname")
	}
	if ParseWant("hello world").Valid {
		t.Fatal("space")
	}
	if ParseWant("A_B").Valid {
		t.Fatal("underscore")
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
}
