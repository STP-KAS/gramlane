package names

import (
	"strings"
	"testing"
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
