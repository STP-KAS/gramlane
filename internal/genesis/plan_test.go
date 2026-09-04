package genesis

import (
	"encoding/hex"
	"testing"
)

func TestLiveKeys(t *testing.T) {
	if len(MustHex(IssuerPubHex)) != 32 || len(MustHex(HolderPubHex)) != 32 {
		t.Fatal("x-only pubkeys")
	}
	p := Live()
	if p.Credits != 500_000 || p.SaleKAS != "0.5" {
		t.Fatalf("%+v", p)
	}
	if p.P2SH == "" || p.P2SH[6] != 'p' {
		t.Fatalf("p2sh %s", p.P2SH)
	}
	if p.Desk == p.Holder {
		t.Fatal("desk must not equal holder")
	}
}

func TestP2SHAddress(t *testing.T) {
	addr, hash, n, err := P2SH()
	if err != nil {
		t.Fatal(err)
	}
	if n < 100 || len(hash) != 64 {
		t.Fatalf("n=%d hash=%s", n, hash)
	}
	if len(addr) < 20 || addr[:6] != "kaspa:" {
		t.Fatal(addr)
	}
	if addr[6] != 'p' {
		t.Fatalf("P2SH should start kaspa:p… got %s", addr)
	}
	if addr == DeskAddress || addr == HolderAddress {
		t.Fatal("p2sh collided with p2pk")
	}
}

func TestRedeemFromSilAbiArray(t *testing.T) {
	live, err := RedeemScript()
	if err != nil || len(live) < 100 {
		t.Fatalf("live: %v n=%d", err, len(live))
	}
	// silverscript#232 example shape: contracts is an array, bytecode may be hex.
	art := []byte(`{"schema_version":1,"contracts":[{"name":"WorkCredit","compiled":{"script_hex":"` + hex.EncodeToString(live[:32]) + `"}}]}`)
	got, err := redeemFromArtifact(art)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 32 || got[0] != live[0] || got[31] != live[31] {
		t.Fatalf("array/hex ABI: n=%d", len(got))
	}
}
