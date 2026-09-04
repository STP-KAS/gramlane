package names

import (
	"fmt"
	"testing"
)

func TestResolveGramsQuote(t *testing.T) {
	if ResolveGrams != 12_000 {
		t.Fatal(ResolveGrams)
	}
}

func TestIsFreeOnIndex(t *testing.T) {
	if !isFreeOnIndex(fmt.Errorf("indexer 404 Not Found")) {
		t.Fatal("404")
	}
	if isFreeOnIndex(fmt.Errorf("timeout")) {
		t.Fatal("timeout")
	}
}

func TestLooksAddr(t *testing.T) {
	if !looksAddr("kaspa:qtest") || looksAddr("kns.kas") {
		t.Fatal("addr")
	}
}
