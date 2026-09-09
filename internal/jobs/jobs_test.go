package jobs

import (
	"strings"
	"testing"
)

func TestResolveJobIsFloorGrams(t *testing.T) {
	j, ok := Get("resolve")
	if !ok || j.Grams != Floor {
		t.Fatalf("%+v", j)
	}
	q, err := QuoteJob(j)
	if err != nil || q.USD != "not quoted" || q.KAS != 0.001 {
		t.Fatalf("%v %+v", err, q)
	}
	for _, c := range Catalog {
		if c.Grams != Floor {
			t.Fatalf("%s %d", c.ID, c.Grams)
		}
	}
}

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
	if len(Fits(Floor-1)) != 0 {
		t.Fatal("under floor")
	}
	if len(Fits(Floor)) != len(Catalog) {
		t.Fatalf("floor fits %d want %d", len(Fits(Floor)), len(Catalog))
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

func TestSequenceJob(t *testing.T) {
	j, ok := Get("sequence")
	if !ok || j.Grams != Floor || j.Lane != "SEQ1" {
		t.Fatalf("%+v", j)
	}
	rec, err := RunAs(j, "kaspadao.kas", "grams", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Settlement != "prepaid-grams" {
		t.Fatalf("settle %s", rec.Settlement)
	}
	if !strings.Contains(rec.Output, "1 dag") || !strings.Contains(rec.Output, "3 postage") {
		t.Fatalf("output %s", rec.Output)
	}
}
