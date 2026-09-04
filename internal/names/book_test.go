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
