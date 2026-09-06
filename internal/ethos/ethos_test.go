package ethos

import "testing"

func TestVisionName(t *testing.T) {
	if Vision != "skip centralized stablecoins for dapps" {
		t.Fatal(Vision)
	}
}
