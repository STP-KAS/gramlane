package post

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStampAndList(t *testing.T) {
	dir := t.TempDir()
	old, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	if err := os.MkdirAll("data", 0o755); err != nil {
		t.Fatal(err)
	}
	s := StampMsg("kaspa:qtest", "hello from tests", 9000)
	if s.Bytes == 0 || s.SHA256 == "" || s.Grams != 9000 {
		t.Fatalf("%+v", s)
	}
	if _, err := os.Stat(filepath.Join("data", "postage.json")); err != nil {
		t.Fatal(err)
	}
	all := List()
	if len(all) != 1 || all[0].Text != "hello from tests" {
		t.Fatalf("%+v", all)
	}
}
