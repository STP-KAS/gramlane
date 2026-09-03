package quote

import "testing"

func TestMillionGramsIsOneKAS(t *testing.T) {
	q, err := Grams(1_000_000, "KNS1")
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

func TestSetAsideHalf(t *testing.T) {
	a, err := SetAside("2", 50)
	if err != nil {
		t.Fatal(err)
	}
	if a.Spend.Grams != 1_000_000 || a.Hold.Grams != 1_000_000 {
		t.Fatalf("%+v", a)
	}
}
