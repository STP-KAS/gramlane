package market

import (
	"testing"

	"gramlane/internal/names"
)

func TestEmptyList(t *testing.T) {
	dir := t.TempDir()
	names.ResetBookForTest(dir)
	names.ResetSignForTest(dir)
	names.ResetCovenantForTest()
	path = dir + "/market.json"
	live = nil
	if List() == nil {
		t.Fatal("list")
	}
}

func TestOpenBuyKASNotGrams(t *testing.T) {
	dir := t.TempDir()
	names.ResetBookForTest(dir)
	names.ResetSignForTest(dir)
	names.ResetCovenantForTest()
	path = dir + "/market.json"
	live = nil
	seller := "kaspa:qseller00000000000000000000000000000000000000000000000000000000000"
	buyer := "kaspa:qbuyer000000000000000000000000000000000000000000000000000000000000"
	if _, err := names.Record(seller, "bakery", "tx-face"); err != nil {
		t.Fatal(err)
	}
	if _, err := names.Record(seller, "extra", "tx-extra"); err != nil {
		t.Fatal(err)
	}
	L, err := Open("extra", seller, 100)
	if err != nil {
		t.Fatal(err)
	}
	if L.USDCents != 100 || L.USD == "" || L.Sompi == 0 || L.KAS == "" {
		t.Fatalf("want USD shelf + KAS settle %+v", L)
	}
	if _, err := Open("extra", seller, 0); err == nil {
		t.Fatal("usd required")
	}
	sold, err := Buy(L.ID, buyer, "kas-txid")
	if err != nil {
		t.Fatal(err)
	}
	if sold.Status != "sold" || sold.Buyer != buyer || sold.Tx != "kas-txid" {
		t.Fatalf("%+v", sold)
	}
	if names.Holds(seller, "extra") {
		t.Fatal("seller still holds")
	}
	if !names.Holds(buyer, "extra") {
		t.Fatal("buyer must receive")
	}
	got := names.Book(buyer)
	if got == nil || len(got.Names) != 1 {
		t.Fatalf("wallet tab %+v", got)
	}
}
