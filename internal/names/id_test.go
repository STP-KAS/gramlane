package names

import "testing"

func TestNormalizeDomain(t *testing.T) {
	if Normalize("Kaspadao") != "kaspadao.kas" {
		t.Fatal(Normalize("Kaspadao"))
	}
}

func TestSortKdaoFirst(t *testing.T) {
	got := sortNames([]string{"zzz.kas", "knskdao.kas", "kaspadao.kas", "kdao.kas", "aaa.kas"})
	if got[0] != "kdao.kas" && got[0] != "kaspadao.kas" {
		t.Fatalf("%v", got)
	}
	if !knsMatch("kaspadao.kas", "kdao") || !knsMatch("knskdao.kas", "kdao") {
		t.Fatal("kdao match")
	}
}
