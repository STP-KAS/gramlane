package names

import "testing"

func TestNormalizeDomain(t *testing.T) {
	if Normalize("Kaspadao") != "kaspadao.kas" {
		t.Fatal(Normalize("Kaspadao"))
	}
}
