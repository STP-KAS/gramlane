package kachat

import "testing"

func TestEnvelopes(t *testing.T) {
	es := Envelopes("kns.kas")
	if len(es) < 4 {
		t.Fatal(len(es))
	}
	c := ContactFrom("kns.kas", "kaspa:qtest")
	if c.Address != "kaspa:qtest" || c.PayURI == "" {
		t.Fatal(c)
	}
	if DeepLink("") != Linktree {
		t.Fatal("docs")
	}
}
