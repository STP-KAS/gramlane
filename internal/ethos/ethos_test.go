package ethos

import "testing"

func TestVisionName(t *testing.T) {
	if Vision != "skip centralised stablecoins for dapps" {
		t.Fatal(Vision)
	}
	if Line == "" || JarIsNakamoto || SecurityBudget != "l1-fee" || FillIsBusiness || FillAmount == "" || FillBurnsKAS {
		t.Fatal(Line, JarIsNakamoto, SecurityBudget, FillIsBusiness, FillAmount, FillBurnsKAS)
	}
}
