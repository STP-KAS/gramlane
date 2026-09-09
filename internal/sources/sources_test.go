package sources

import "testing"

func TestKIP21IsOnTheDeskURL(t *testing.T) {
	p, ok := ByLocal("/kip-21")
	if !ok || p.URL == "" || p.ID != "kip-21" {
		t.Fatalf("%v %+v", ok, p)
	}
	if _, ok := ByLocal("/kip-2"); ok {
		t.Fatal("KIP-2 must not get a local pin URL")
	}
	if len(Used) < 10 {
		t.Fatalf("used %d", len(Used))
	}
	v1, ok := ByLocal("/v1.0.0")
	if !ok || v1.Chip != "pin" {
		t.Fatalf("v1.0.0 pin %+v", v1)
	}
	if _, ok := ByLocal("/darwin"); !ok {
		t.Fatal("darwin")
	}
}
