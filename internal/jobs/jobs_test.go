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
}

func TestPrepaidToken(t *testing.T) {
	if !prepaidToken("grams") || !prepaidToken("kaspa-work-credit") {
		t.Fatal("prepaid")
	}
	if prepaidToken("c1799b0de40f71cfd7a153684ef22326ad920d0dca2a8b519ce2c8379c4f7bc2") {
		t.Fatal("raw txid is kas-fallback unless seq.Accepts")
	}
}
