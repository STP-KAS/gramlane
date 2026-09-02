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
