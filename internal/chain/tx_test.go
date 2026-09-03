package chain

import "testing"

func TestIsTxID(t *testing.T) {
	if !IsTxID("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatal("64 hex")
	}
	if IsTxID("local-demo") {
		t.Fatal("not a txid")
	}
	if IsTxID("") {
		t.Fatal("empty")
	}
}
