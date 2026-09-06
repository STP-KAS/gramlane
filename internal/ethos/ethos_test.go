package ethos

import "testing"

func TestVisionName(t *testing.T) {
	if Headline != "project delusional" {
		t.Fatal(Headline)
	}
	if Vision != "skip centralised stablecoins for dapps" {
		t.Fatal(Vision)
	}
	if Line != "Kaspa master file, ideas and principles. Still a delusional idea." {
		t.Fatal(Line)
	}
	if JarIsNakamoto || SecurityBudget != "l1-fee" || FillIsBusiness || FillAmount == "" || FillBurnsKAS {
		t.Fatal(JarIsNakamoto, SecurityBudget, FillIsBusiness, FillAmount, FillBurnsKAS)
	}
}
