package names

import (
	"strings"
	"testing"

	"gramlane/internal/quote"
)

func TestSuggestUSDAffordableLong(t *testing.T) {
	ResetSignForTest(t.TempDir())
	g := SuggestUSDCents("bakery")
	if g < 10 || g > 200 {
		t.Fatalf("bakery %d cents", g)
	}
	short := SuggestUSDCents("x")
	if short <= g {
		t.Fatalf("short %d long %d", short, g)
	}
	if SuggestUSDCents("aaaaaaaa") >= SuggestUSDCents("shop") {
		t.Fatal("common word should list above repeating junk")
	}
	a := AskFor("bakery")
	if a.USD == "" || a.PaySompi == 0 || !strings.Contains(a.Note, "KAS") {
		t.Fatalf("%+v", a)
	}
	if strings.Contains(a.Note, "grams at 100") {
		t.Fatal("name settlement is KAS, not grams")
	}
}

func TestFundMintHalfToVault(t *testing.T) {
	m := FundMint()
	if m.SharePct != 50 || m.NameSompi != m.VaultSompi || m.TotalSompi != m.NameSompi*2 {
		t.Fatalf("%+v", m)
	}
	if m.NameSompi != quote.KaswareMinSompi || m.TotalKAS != "1" {
		t.Fatalf("wallet floor %+v", m)
	}
	if m.Vault != AdoptionVault || !strings.HasPrefix(m.Vault, "kaspa:q") {
		t.Fatalf("vault %s", m.Vault)
	}
	q := FundQuote()
	if q.PaySompi != m.NameSompi || q.PayKASText != "1" {
		t.Fatalf("quote %+v mint %+v", q, m)
	}
}
