package names

import "testing"

func TestFirstRegisteredIsFace(t *testing.T) {
	ResetCovenantForTest()
	ResetBookForTest(t.TempDir())
	addr := "kaspa:qtestface0000000000000000000000000000000000000000000000000000000000"
	b, err := Record(addr, "alpha", "tx1")
	if err != nil {
		t.Fatal(err)
	}
	if b.Primary != "alpha.kas" || b.First != "alpha.kas" {
		t.Fatalf("%+v", b)
	}
	b, err = Record(addr, "beta", "tx2")
	if err != nil {
		t.Fatal(err)
	}
	if b.Primary != "alpha.kas" || len(b.Names) != 2 {
		t.Fatalf("second must stay in drawer %+v", b)
	}
	if err := SetPrimary(addr, "beta"); err != nil {
		t.Fatal(err)
	}
	got := Book(addr)
	if got.Primary != "beta.kas" {
		t.Fatalf("switch face %+v", got)
	}
	var face, drawer int
	for _, h := range got.Names {
		if h.Face {
			face++
		} else {
			drawer++
		}
	}
	if face != 1 || drawer != 1 {
		t.Fatalf("face=%d drawer=%d", face, drawer)
	}
}

func TestVaultFreeMintAndSwitchFace(t *testing.T) {
	dir := t.TempDir()
	ResetCovenantForTest()
	ResetBookForTest(dir)
	ResetReserveForTest(dir)
	ResetSignForTest(dir)
	if err := SeedVault(); err != nil {
		t.Fatal(err)
	}
	other := "kaspa:qbuyer000000000000000000000000000000000000000000000000000000000000"
	if _, err := Record(other, "google", "tx"); err == nil {
		t.Fatal("retailer")
	}
	if _, err := Record(other, "zz", "tx"); err == nil {
		t.Fatal("sale reserved")
	}
	if _, err := Record(AdoptionVault, "google", "test-free"); err != nil {
		t.Fatal(err)
	}
	if _, err := Record(AdoptionVault, "zz", "test-free"); err != nil {
		t.Fatal(err)
	}
	if _, err := Record(AdoptionVault, "bakery", "test-free"); err != nil {
		t.Fatal(err)
	}
	b := Book(AdoptionVault)
	if b == nil || len(b.Names) != 3 || b.Primary != "google.kas" {
		t.Fatalf("%+v", b)
	}
	if err := Pin(AdoptionVault, "bakery"); err != nil {
		t.Fatal(err)
	}
	id := Face(AdoptionVault)
	if id.Linked != "bakery.kas" || len(id.Names) != 3 || id.Address != AdoptionVault {
		t.Fatalf("%+v", id)
	}
	if err := Pin(other, "bakery"); err == nil {
		t.Fatal("other wallet")
	}
}
