package quote

import "testing"

func TestMillionGramsIsOneKAS(t *testing.T) {
	q, err := Grams(1_000_000, "KNS1")
	if err != nil {
		t.Fatal(err)
	}
	if q.KAS != 1 || q.Credits != 1_000_000 || q.USD != "not quoted" {
		t.Fatalf("%+v", q)
	}
}
