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
