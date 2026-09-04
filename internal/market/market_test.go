package market

import (
	"testing"

	"gramlane/internal/seq"
)

func TestEmptyList(t *testing.T) {
	seq.ResetForTest(t.TempDir())
	path = t.TempDir() + "/market.json"
	live = nil
	if List() == nil {
		t.Fatal("list")
	}
}
