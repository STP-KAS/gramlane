package prior

import (
	"strings"
	"testing"

	"gramlane/internal/names"
)

func TestStampAndLock(t *testing.T) {
	dir := t.TempDir()
	names.ResetCovenantForTest()
	ResetForTest(dir)
	h := HashOf("the bakery recipe")
	if len(h) != 64 {
		t.Fatal(h)
	}
	r, err := File(h, "recipe", "kaspa:qauthor00000000000000000000000000000000000000000000000000000000000", "bakery", "", "", 14000)
	if err != nil {
		t.Fatal(err)
	}
	if r.Hash != h || r.Lock == "" || !strings.HasPrefix(r.Lock, "kaspa:p") {
		t.Fatalf("%+v", r)
	}
	nameLock, _, _, err := names.P2SHFor("bakery")
	if err != nil {
		t.Fatal(err)
	}
	if r.Lock == nameLock {
		t.Fatal("prior lock must not be bakery.kas")
	}
	got := Lookup(h)
	if got == nil || got.Owner != r.Owner {
		t.Fatalf("%+v", got)
	}
	if _, err := File(h, "", "kaspa:qother00000000000000000000000000000000000000000000000000000000000", "", "", "", 1); err == nil {
		t.Fatal("stranger overwrite")
	}
}

func TestBadHash(t *testing.T) {
	if NormalizeHash("zz") != "" || NormalizeHash("") != "" {
		t.Fatal("bad")
	}
}
