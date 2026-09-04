package names

import "testing"

func TestP2SHDiffersPerName(t *testing.T) {
	ResetCovenantForTest()
	a, ha, ra, err := P2SHFor("bakery")
	if err != nil || a == "" || len(ra) < 32 || ha == "" {
		t.Fatalf("%v %s", err, a)
	}
	b, hb, _, err := P2SHFor("google")
	if err != nil {
		t.Fatal(err)
	}
	if a == b || ha == hb {
		t.Fatal("names must not share a P2SH")
	}
	if a[6] != 'p' {
		t.Fatalf("want script hash address, got %s", a)
	}
	c, _, _, err := P2SHFor("kns.kas")
	if err != nil || c == a || c == b {
		t.Fatalf("kns.kas must be a different covenant address: %v %s", err, c)
	}
}

func TestResolveCovenantNoKNS(t *testing.T) {
	ResetCovenantForTest()
	r := ResolveCovenant("google")
	if r.Evidence == "indexer" || r.Evidence == "live" {
		t.Fatal(r.Evidence)
	}
	if r.PayURI == "" || r.ScriptHash == "" {
		t.Fatalf("expected derived P2SH: %+v", r)
	}
	if r.Hit {
		t.Fatal("skipChain should not mark funded")
	}
	kns := ResolveCovenant("kns.kas")
	if kns.PayURI == r.PayURI {
		t.Fatal("kns.kas must not reuse google script")
	}
}

func TestDisplayName(t *testing.T) {
	if DisplayName("Bakery") != "bakery.kas" || DisplayName("google.kas") != "google.kas" {
		t.Fatal(DisplayName("Bakery"), DisplayName("google.kas"))
	}
}
