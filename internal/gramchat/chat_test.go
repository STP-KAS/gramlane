package gramchat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRoomAndPut(t *testing.T) {
	dir := t.TempDir()
	ResetForTest(dir)
	if Room("Bakery") != "bakery.kas" || Room("#board") != "board" {
		t.Fatal(Room("Bakery"), Room("#board"))
	}
	if _, err := Put("bakery", "aa", "bb", 1); err == nil {
		t.Fatal("short nonce")
	}
	nonce := "00112233445566778899aabb"
	box := "00112233445566778899aabbccddeeff00112233"
	n, err := Put("bakery", nonce, box, 9000)
	if err != nil || n.Room != "bakery.kas" || n.Grams != 9000 {
		t.Fatalf("%v %+v", err, n)
	}
	got := List("bakery.kas")
	if len(got) != 1 || got[0].Box != box {
		t.Fatalf("%+v", got)
	}
	if len(List("other.kas")) != 0 {
		t.Fatal("leak")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "telegram.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), `"from"`) {
		t.Fatalf("desk must not store sender: %s", raw)
	}
}

func TestDropOldFrom(t *testing.T) {
	dir := t.TempDir()
	ResetForTest(dir)
	nonce := "00112233445566778899aabb"
	box := "00112233445566778899aabbccddeeff00112233"
	old := `[{"room":"bakery.kas","when":"2026-01-01T00:00:00Z","from":"kaspa:qsecret","nonce":"` + nonce + `","box":"` + box + `","grams":9}]`
	if err := os.WriteFile(filepath.Join(dir, "telegram.json"), []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	if n := List("bakery.kas"); len(n) != 1 {
		t.Fatalf("%+v", n)
	}
	saved, err := os.ReadFile(filepath.Join(dir, "telegram.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(saved), "qsecret") || strings.Contains(string(saved), `"from"`) {
		t.Fatalf("old sender still on disk: %s", saved)
	}
}
