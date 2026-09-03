package genesis

import "testing"

func TestLiveKeys(t *testing.T) {
	if len(MustHex(IssuerPubHex)) != 32 || len(MustHex(HolderPubHex)) != 32 {
		t.Fatal("x-only pubkeys")
	}
	p := Live()
	if p.Credits != 500_000 || p.SaleKAS != "0.5" || p.VoucherOnChain {
		t.Fatalf("%+v", p)
	}
	if p.Desk == p.Holder {
		t.Fatal("desk must not equal holder")
	}
}
