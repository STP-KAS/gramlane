package names

import "testing"

func TestSuggestGramsAffordableLong(t *testing.T) {
	g := SuggestGrams("bakery")
	if g < 10_000 || g > 200_000 {
		t.Fatalf("bakery %d", g)
	}
	short := SuggestGrams("x")
	if short <= g {
		t.Fatalf("short %d long %d", short, g)
	}
	if SuggestGrams("aaaaaaaa") >= SuggestGrams("shop") {
		t.Fatal("common word should list above repeating junk")
	}
}
