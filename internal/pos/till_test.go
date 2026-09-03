package pos

import (
	"testing"

	"gramlane/internal/seq"
)

func TestInvoicePay(t *testing.T) {
	dir := t.TempDir()
	seq.ResetForTest(dir)
	ResetForTest(dir)
	inv, err := Create("coffee", 50_000, "cafe.kas", "", "location")
	if err != nil || inv.ID == "" || inv.Status != "open" {
		t.Fatalf("%v %+v", err, inv)
	}
	got, ok := Get(inv.ID)
	if !ok || got.Item != "coffee" {
		t.Fatal("get")
	}
	paid, err := Pay(inv.ID, "kaspa:qtest")
	if err != nil {
		t.Fatal(err)
	}
	if paid.Status != "paid" {
		t.Fatal(paid.Status)
	}
	ms := Merchants()
	if len(ms) != 1 || ms[0].Owed != 50_000 {
		t.Fatalf("%+v", ms)
	}
	if seq.Snap().Remaining != 450_000 {
		t.Fatalf("remaining %d", seq.Snap().Remaining)
	}
}

func TestMintThenPay(t *testing.T) {
	dir := t.TempDir()
	seq.ResetForTest(dir)
	tx := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	led, err := seq.MintFromTx(tx, 100_000, 10_000_000)
	if err != nil || led.Remaining != 600_000 {
		t.Fatalf("%v %+v", err, led)
	}
	if _, err := seq.MintFromTx(tx, 1, 1); err == nil {
		t.Fatal("dup mint")
	}
}
