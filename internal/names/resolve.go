package names

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"gramlane/internal/appenv"
	"gramlane/internal/quote"
)

const ResolveGrams uint64 = 12_000

// Result is a kasdomain hit. Evidence is never live or indexer.
type Result struct {
	Query      string          `json:"query"`
	Name       string          `json:"name"`
	Owner      string          `json:"owner,omitempty"`
	PayURI     string          `json:"payUri,omitempty"`
	Evidence   string          `json:"evidence"`
	Warning    string          `json:"warning"`
	Quote      quote.Quote     `json:"quote"`
	Raw        json.RawMessage `json:"raw,omitempty"`
	ScriptHash string          `json:"scriptHash,omitempty"`
	Hex        string          `json:"hex,omitempty"`
	Layout     string          `json:"layout,omitempty"`
	Hit        bool            `json:"hit"`
}

type fixture struct {
	Name       string `json:"name"`
	OwnerPub   string `json:"ownerPub"`
	PayURI     string `json:"payUri"`
	ScriptHash string `json:"scriptHash"`
	Hex        string `json:"hex"`
	Layout     string `json:"layout"`
}

var (
	fixOnce sync.Once
	fixBook map[string]fixture
)

func DisplayName(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func loadFixtures() {
	fixOnce.Do(func() {
		fixBook = map[string]fixture{}
		b, err := os.ReadFile(appenv.File("kasdomain-fixture.json"))
		if err != nil || len(b) == 0 {
			return
		}
		var env struct {
			Names []fixture `json:"names"`
		}
		if json.Unmarshal(b, &env) != nil {
			return
		}
		for _, f := range env.Names {
			k := DisplayName(f.Name)
			if k == "" {
				continue
			}
			fixBook[k] = f
		}
	})
}

func ResetFixturesForTest() {
	fixOnce = sync.Once{}
	fixBook = nil
}

// ResolveCovenant never calls KNS. Empty is correct. A hit is a local fixture only.
func ResolveCovenant(raw string) *Result {
	raw = strings.TrimSpace(raw)
	name := DisplayName(raw)
	qq, err := quote.Grams(ResolveGrams, "SIGN1")
	if err != nil {
		qq = quote.Quote{Grams: ResolveGrams, USD: "not quoted"}
	}
	out := &Result{
		Query:    raw,
		Name:     name,
		Evidence: "roadmap",
		Quote:    qq,
		Warning:  fmt.Sprintf("No kasdomain covenant for %s. Inscriptions and KNS are out of scope.", name),
	}
	if name == "" {
		out.Warning = "Type a name. This app does not register names yet."
		return out
	}
	loadFixtures()
	f, ok := fixBook[name]
	if !ok {
		return out
	}
	out.Hit = true
	out.Evidence = "local"
	out.Owner = f.OwnerPub
	out.PayURI = f.PayURI
	out.ScriptHash = f.ScriptHash
	out.Hex = f.Hex
	out.Layout = f.Layout
	out.Warning = "Local fixture. Operator-created name UTXO. Not a live unique name on L1. Tagged local, not indexer."
	rawJSON, _ := json.Marshal(f)
	out.Raw = rawJSON
	return out
}
