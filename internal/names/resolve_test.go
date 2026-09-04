package names

import (
	"os"
	"path/filepath"
	"testing"

	"gramlane/internal/appenv"
)

func TestResolveCovenantDoesNotUseKNS(t *testing.T) {
	ResetFixturesForTest()
	r := ResolveCovenant("google")
	if r.Hit || r.Owner != "" || r.Evidence != "roadmap" {
		t.Fatalf("google must not be a KNS hit: %+v", r)
	}
	if r.Evidence == "live" || r.Evidence == "indexer" {
		t.Fatal(r.Evidence)
	}
	kns := ResolveCovenant("kns.kas")
	if kns.Hit || kns.Owner != "" {
		t.Fatalf("kns.kas is not kasdomain: %+v", kns)
	}
}

func TestResolveCovenantLocalFixture(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DATA_DIR", dir)
	ResetFixturesForTest()
	body := []byte(`{"names":[{"name":"demo","ownerPub":"aa","payUri":"kaspa:qtest","scriptHash":"bb","hex":"deadbeef","layout":"State { label, owner }"}]}`)
	if err := os.WriteFile(filepath.Join(dir, "kasdomain-fixture.json"), body, 0o644); err != nil {
		t.Fatal(err)
	}
	// appenv.File uses DATA_DIR at call time; loadFixtures reads it.
	_ = appenv.File("kasdomain-fixture.json")
	r := ResolveCovenant("demo")
	if !r.Hit || r.Evidence != "local" || r.Hex != "deadbeef" {
		t.Fatalf("%+v", r)
	}
	if r.Evidence == "indexer" || r.Evidence == "live" {
		t.Fatal(r.Evidence)
	}
}

func TestDisplayName(t *testing.T) {
	if DisplayName(" Google ") != "google" {
		t.Fatal(DisplayName(" Google "))
	}
}
