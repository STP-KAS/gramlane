package seq

import "testing"

func TestBurnPrepaidGrams(t *testing.T) {
	ResetForTest(t.TempDir())
	if !CanBurn(5_000) {
		t.Fatal("fresh ledger should cover heartbeat")
	}
	if !Accepts("grams") || !Accepts("prepaid") {
		t.Fatal("accepts")
	}
	led, err := BurnGrams("dag", 5_000, "grams")
	if err != nil {
		t.Fatal(err)
	}
	if led.Remaining != 495_000 {
		t.Fatalf("remaining %d", led.Remaining)
	}
	if CanBurn(500_000) {
		t.Fatal("should not cover full mint after a burn")
	}
	led, err = BurnGrams("batch", 40_000, "grams")
	if err != nil {
		t.Fatal(err)
	}
	if led.Remaining != 455_000 {
		t.Fatalf("remaining %d", led.Remaining)
	}
	snap := Snap()
	if snap.Remaining != 455_000 || len(snap.Burns) != 2 {
		t.Fatalf("%+v", snap)
	}
	if _, err := BurnGrams("batch", 500_000, "grams"); err == nil {
		t.Fatal("expected shortfall")
	}
}

func TestMintFromTx(t *testing.T) {
	ResetForTest(t.TempDir())
	tx := "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	led, err := MintFromTx(tx, 10_000, 1_000_000)
	if err != nil {
		t.Fatal(err)
	}
	if led.Remaining != 510_000 || led.Credits != 510_000 {
		t.Fatalf("%+v", led)
	}
}

func TestAcceptsSaleTx(t *testing.T) {
	ResetForTest(t.TempDir())
	if !Accepts("c1799b0de40f71cfd7a153684ef22326ad920d0dca2a8b519ce2c8379c4f7bc2") {
		t.Fatal("sale tx should settle prepaid")
	}
	if Accepts("not-a-receipt") {
		t.Fatal("junk")
	}
}
