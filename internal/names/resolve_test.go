package names

import "testing"

func TestResolveGramsQuote(t *testing.T) {
	if ResolveGrams != 12_000 {
		t.Fatal(ResolveGrams)
	}
}

func TestLooksAddr(t *testing.T) {
	if !looksAddr("kaspa:qtest") || looksAddr("kns.kas") {
		t.Fatal("addr")
	}
}
