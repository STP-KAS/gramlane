package shop

import (
	"strings"
	"testing"

	"gramlane/internal/names"
	"gramlane/internal/pos"
	"gramlane/internal/seq"
)

func setup(t *testing.T) (seller string) {
	t.Helper()
	dir := t.TempDir()
	names.ResetBookForTest(dir)
	names.ResetSignForTest(dir)
	names.ResetCovenantForTest()
	ResetForTest(dir)
	pos.ResetForTest(dir)
	seq.ResetForTest(dir)
	seller = "kaspa:qshop000000000000000000000000000000000000000000000000000000000000"
	if _, err := names.Record(seller, "bakery", "tx1"); err != nil {
		t.Fatal(err)
	}
	return seller
}

func TestHangOnlyHolder(t *testing.T) {
	seller := setup(t)
	if _, err := Hang("bakery", "kaspa:qother", "", "Bread", "We bake."); err == nil {
		t.Fatal("stranger hung a shop")
	}
	sh, err := Hang("bakery", seller, "", "Bread", "We bake.")
	if err != nil || sh.Name != "bakery.kas" || sh.Headline != "Bread" {
		t.Fatalf("%v %+v", err, sh)
	}
	if sh.P2SH == "" || !strings.HasPrefix(sh.P2SH, "kaspa:p") {
		t.Fatalf("want covenant P2SH %+v", sh)
	}
}

func TestMenuGramsNotKNS(t *testing.T) {
	seller := setup(t)
	if _, err := Hang("bakery", seller, seller, "Bakery", "Open."); err != nil {
		t.Fatal(err)
	}
	sh, err := AddItem("bakery", seller, "coffee", 300)
	if err != nil || len(sh.Items) != 1 {
		t.Fatalf("%v %+v", err, sh)
	}
	it := sh.Items[0]
	if it.USD != "$3.00" || it.Grams == 0 {
		t.Fatalf("USD shelf + grams till %+v", it)
	}
	inv, err := Ticket("bakery", it.ID, "both")
	if err != nil || inv.Merchant != "bakery.kas" || inv.Grams != it.Grams {
		t.Fatalf("%v %+v", err, inv)
	}
	v := View("bakery")
	if v.Evidence == "indexer" || v.Evidence == "live" {
		t.Fatal(v.Evidence)
	}
	if v.Shop == nil || v.NameSettle != "kas" || v.TillUnit != "gram" {
		t.Fatalf("%+v", v)
	}
	if v.Web4["kns"] != false || v.Web4["vprogs"] != "roadmap" {
		t.Fatalf("web4 %+v", v.Web4)
	}
}

func TestViewDoesNotNeedShop(t *testing.T) {
	setup(t)
	v := View("ghost")
	if v.Shop != nil {
		t.Fatal("empty shop")
	}
	if v.P2SH == "" || v.Evidence != "compiled" {
		t.Fatalf("%+v", v)
	}
}
