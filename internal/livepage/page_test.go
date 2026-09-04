package livepage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	old := path
	path = filepath.Join(dir, "live-pages.json")
	book = nil
	t.Cleanup(func() {
		path = old
		book = nil
	})
	first := Ensure("kaspadao.kas", "A", "about A", "")
	if first.Headline != "A" {
		t.Fatalf("%+v", first)
	}
	second := Ensure("kaspadao.kas", "B", "about B", "")
	if second.Headline != "A" {
		t.Fatalf("overwrote: %+v", second)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
