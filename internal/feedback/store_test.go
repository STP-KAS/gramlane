package feedback

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRejectsSeedShaped(t *testing.T) {
	seed := "abandon ability able about above absent absorb abstract absurd abuse access accident"
	if !LooksLikeSeed(seed) {
		t.Fatal("expected seed detect")
	}
	if LooksLikeSeed("the desk job failed and the quote was wrong") {
		t.Fatal("false positive")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("KASPA_FEEDBACK_DIR", dir)
	n, err := Save("kns", "please add a logout button", "")
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "kns", n.ID+".json")
	if _, err := os.Stat(p); err != nil {
		t.Fatal(err)
	}
	if _, err := Save("kns", "abandon ability able about above absent absorb abstract absurd abuse access accident", ""); err == nil {
		t.Fatal("seed should fail")
	}
}
