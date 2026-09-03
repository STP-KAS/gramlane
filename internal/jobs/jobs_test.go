package jobs

import "testing"

func TestCatalogQuoted(t *testing.T) {
	if len(Catalog) < 3 {
		t.Fatal("catalog")
	}
	j, ok := Get("resolve")
	if !ok {
		t.Fatal("resolve")
	}
	q, err := QuoteJob(j)
	if err != nil || q.Grams != j.Grams {
		t.Fatalf("%v %+v", err, q)
	}
	if _, ok := Get("vault"); !ok {
		t.Fatal("vault job")
	}
	if _, ok := Get("postage"); !ok {
		t.Fatal("postage job")
	}
}

func TestFits(t *testing.T) {
	f := Fits(8_000)
	if len(f) < 2 {
		t.Fatalf("%d", len(f))
	}
	if Fits(4_000) != nil && len(Fits(4_000)) != 0 {
		t.Fatal("under heartbeat")
	}
}

func TestPrepaidToken(t *testing.T) {
	if !prepaidToken("grams") || !prepaidToken("kaspa-work-credit") {
		t.Fatal("prepaid")
	}
	if prepaidToken("c1799b0de40f71cfd7a153684ef22326ad920d0dca2a8b519ce2c8379c4f7bc2") {
		t.Fatal("raw txid is kas-fallback unless seq.Accepts")
	}
}

func TestVaultJob(t *testing.T) {
	v, _ := Get("vault")
	rec, err := RunAs(v, "", "grams", "", "")
	if err != nil || rec.Output == "" {
		t.Fatalf("vault %v %s", err, rec.Output)
	}
	if rec.Settlement != "prepaid-grams" {
		t.Fatalf("settle %s", rec.Settlement)
	}
	if _, ok := Get("agent"); !ok {
		t.Fatal("agent")
	}
	if _, ok := Get("site"); !ok {
		t.Fatal("site")
	}
}
