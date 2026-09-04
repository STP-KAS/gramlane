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
	if it.USD != "$3.00" || it.KAS == "" {
		t.Fatalf("USD shelf + KAS till %+v", it)
	}
	inv, err := Ticket("bakery", "coffee", "both")
	if err != nil || inv.Merchant != "bakery.kas" {
		t.Fatalf("%v %+v", err, inv)
	}
	v := View("bakery")
	if v.Evidence == "indexer" || v.Evidence != "live" {
		t.Fatalf("want live covenant: %s", v.Evidence)
	}
	if v.Shop == nil || v.NameSettle != "kas" || v.TillUnit != "kas" {
		t.Fatalf("%+v", v)
	}
	if v.Web4["kns"] != false || v.Web4["vprogs"] != "roadmap" {
		t.Fatalf("web4 %+v", v.Web4)
	}
}

func TestHangNamesTheField(t *testing.T) {
	seller := setup(t)
	_, err := Hang("", seller, "", "Bread", "We bake.")
	fe, ok := AsField(err)
	if !ok || fe.Field != "name" {
		t.Fatalf("%v", err)
	}
	_, err = Hang("bakery", "not-an-addr", "", "Bread", "")
	fe, ok = AsField(err)
	if !ok || fe.Field != "address" {
		t.Fatalf("%v", err)
	}
	_, err = AddItem("bakery", seller, "", 100)
	fe, ok = AsField(err)
	if !ok || fe.Field != "label" {
		t.Fatalf("%v", err)
	}
}

func TestChoicesMarksShop(t *testing.T) {
	seller := setup(t)
	if _, err := names.Record(seller, "extra", "tx2"); err != nil {
		t.Fatal(err)
	}
	if _, err := Hang("bakery", seller, seller, "Bread", ""); err != nil {
		t.Fatal(err)
	}
	cs := Choices(seller)
	if len(cs) != 2 {
		t.Fatalf("%+v", cs)
	}
	var bakery, extra Choice
	for _, c := range cs {
		if c.Name == "bakery.kas" {
			bakery = c
		}
		if c.Name == "extra.kas" {
			extra = c
		}
	}
	if !bakery.HasShop || extra.HasShop {
		t.Fatalf("%+v %+v", bakery, extra)
	}
	if FromPath("/bakery.kas") != "bakery.kas" || FromPath("/s/bakery.kas") != "bakery.kas" {
		t.Fatal(FromPath("/bakery.kas"), FromPath("/s/bakery.kas"))
	}
	if FromHost("gramlane.bakery.kas:8081") != "bakery.kas" {
		t.Fatal(FromHost("gramlane.bakery.kas:8081"))
	}
	if PagePath("bakery") != "/bakery.kas" {
		t.Fatal(PagePath("bakery"))
	}
}

func TestViewDoesNotNeedShop(t *testing.T) {
	setup(t)
	v := View("ghost")
	if v.Shop != nil {
		t.Fatal("empty shop")
	}
	if v.P2SH == "" || v.Evidence != "live" {
		t.Fatalf("%+v", v)
	}
}
