package quote

import (
	"strings"
	"testing"
)

func TestMillionGramsIsOneKAS(t *testing.T) {
	q, err := Grams(1_000_000, "SIGN1")
	if err != nil {
		t.Fatal(err)
	}
	if q.KAS != 1 || q.Credits != 1_000_000 || q.USD != "not quoted" {
		t.Fatalf("%+v", q)
	}
}

func TestKaswareFloorIsOneCentKAS(t *testing.T) {
	q, err := Grams(5_000, "SEQ1")
	if err != nil {
		t.Fatal(err)
	}
	if q.Sompi != 500_000 || q.KAS != 0.005 {
		t.Fatalf("policy %+v", q)
	}
	if q.PaySompi != KaswareMinSompi || q.PayKASText != "0.5" {
		t.Fatalf("kasware floor %+v", q)
	}
}

func TestConvertHalfKASIs500kGrams(t *testing.T) {
	c, err := Parse("kas", "0.5")
	if err != nil {
		t.Fatal(err)
	}
	if c.Grams != 500_000 || c.KASText != "0.5" || c.KCC20 {
		t.Fatalf("%+v", c)
	}
	g, err := Parse("grams", "1000000")
	if err != nil || g.KASText != "1" {
		t.Fatalf("%v %+v", err, g)
	}
	s, err := FromSompi(150)
	if err != nil || s.Grams != 1 || s.Dust != 50 {
		t.Fatalf("%v %+v", err, s)
	}
}

func TestFormatKAS1OneDecimal(t *testing.T) {
	if FormatKAS1(SompiPerKAS) != "1.0" || FormatKAS1(50_000_000) != "0.5" {
		t.Fatal(FormatKAS1(SompiPerKAS), FormatKAS1(50_000_000))
	}
}

func TestFromUSDSettlesKASNotGramsNote(t *testing.T) {
	b, err := FromUSD("1.00", "0.10")
	if err != nil {
		t.Fatal(err)
	}
	if b.Ccy != "USD" || b.Sompi != 10*SompiPerKAS || b.KASText != "10" {
		t.Fatalf("%+v", b)
	}
	if b.Note == "" || !strings.Contains(b.Note, "KAS on L1") {
		t.Fatalf("want KAS settlement note %+v", b)
	}
	if strings.Contains(b.Note, "grams at 100") {
		t.Fatal("FromUSD must not settle in grams")
	}
	c, err := USDCents("$0.50")
	if err != nil || c != 50 {
		t.Fatalf("%v %d", err, c)
	}
	if FormatUSD(40) != "$0.40" || FloorPay(1) != KaswareMinSompi {
		t.Fatal(FormatUSD(40), FloorPay(1))
	}
}

func TestThreeEuroCoffee(t *testing.T) {
	b, err := FromFiat("3", "EUR", "1")
	if err != nil {
		t.Fatal(err)
	}
	if b.Grams != 3_000_000 || b.Ccy != "EUR" {
		t.Fatalf("%+v", b)
	}
	b, err = FromFiat("3", "eur", "0.5")
	if err != nil {
		t.Fatal(err)
	}
	if b.Grams != 6_000_000 {
		t.Fatalf("half-kas rate %+v", b)
	}
}

func TestSetAsideHalf(t *testing.T) {
	a, err := SetAside("2", 50)
	if err != nil {
		t.Fatal(err)
	}
	if a.Spend.Grams != 1_000_000 || a.Hold.Grams != 1_000_000 {
		t.Fatalf("%+v", a)
	}
}
