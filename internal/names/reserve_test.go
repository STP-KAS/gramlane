package names

import "testing"

func TestSalePriceByLength(t *testing.T) {
	if SalePriceKAS("a") != 1000 || SalePriceKAS("7") != 1000 {
		t.Fatal("1")
	}
	if SalePriceKAS("ab") != 500 || SalePriceKAS("42") != 500 {
		t.Fatal("2")
	}
	if SalePriceKAS("abc") != 250 {
		t.Fatal("3")
	}
	if SalePriceKAS("bakery") != 30 {
		t.Fatal("rest")
	}
}

func TestSeedVaultShortAndRetailer(t *testing.T) {
	dir := t.TempDir()
	ResetBookForTest(dir)
	ResetReserveForTest(dir)
	ResetCovenantForTest()
	ResetSignForTest(dir)
	if err := SeedVault(); err != nil {
		t.Fatal(err)
	}
	a := LookupHold("a")
	if a == nil || a.Kind != HoldSale || a.KAS != 1000 || a.Owner != AdoptionVault {
		t.Fatalf("%+v", a)
	}
	ab := LookupHold("ab")
	if ab == nil || ab.Kind != HoldSale || ab.KAS != 500 {
		t.Fatalf("%+v", ab)
	}
	g := LookupHold("google")
	if g == nil || g.Kind != HoldRetailer {
		t.Fatalf("google %+v", g)
	}
	bank := LookupHold("bank")
	if bank == nil || bank.Kind != HoldRetailer {
		t.Fatalf("bank %+v", bank)
	}
	if LookupHold("bakery") != nil {
		t.Fatal("bakery should stay open to mint")
	}
	if _, err := BuyReserved("google", "kaspa:qbuyer000000000000000000000000000000000000000000000000000000000000", "tx"); err == nil {
		t.Fatal("retailer sold")
	}
	buyer := "kaspa:qbuyer000000000000000000000000000000000000000000000000000000000000"
	got, err := BuyReserved("zz", buyer, "txid")
	if err != nil || got.Kind != HoldSold || got.Buyer != buyer {
		t.Fatalf("%v %+v", err, got)
	}
	if !Holds(buyer, "zz") {
		t.Fatal("buyer should receive")
	}
}
