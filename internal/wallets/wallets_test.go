package wallets

import "testing"

func TestCatalogCoversRequested(t *testing.T) {
	need := []string{
		"tangem", "ledger", "onekey", "ellipal", "safepal-x1",
		"kaspium", "kasware", "kastle", "web-wallet", "kdx", "kaskeeper", "kurncy",
		"zelcore", "okx", "now", "guarda", "bitget", "mathwallet", "safepal-app",
		"kaspa-ng",
	}
	have := map[string]bool{}
	for _, w := range All() {
		if w.URL == "" || w.Name == "" {
			t.Fatalf("incomplete %+v", w)
		}
		have[w.ID] = true
	}
	for _, id := range need {
		if !have[id] {
			t.Fatalf("missing %s", id)
		}
	}
	if len(Injected()) != 2 {
		t.Fatalf("inject count %d want kasware+kastle", len(Injected()))
	}
}
