package names

import "testing"

func TestNormalize(t *testing.T) {
	if Normalize("") != "kns.kas" {
		t.Fatal("empty")
	}
	if Normalize("KNS") != "kns.kas" {
		t.Fatal("tld")
	}
	if Normalize("kas://Alice.kas") != "alice.kas" {
		t.Fatal("kas uri")
	}
}
