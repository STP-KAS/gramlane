package names

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/blake2b"

	"gramlane/internal/chain"
	"gramlane/internal/kasaddr"
	"gramlane/internal/quote"
)

const (
	labelOffset = 2
	labelLen    = 32
	artifactRel = "contracts/v1/KasName.json"
	labelDomain = "kasdomain:v1:"
)

var (
	skipChain bool
	tplMu     sync.Mutex
	tplCached []byte
)

// Result is a kasdomain hit. Evidence is live when KasName.sil derives a P2SH.
// Never an indexer record. Not KNS.
type Result struct {
	Query      string          `json:"query"`
	Name       string          `json:"name"`
	Owner      string          `json:"owner,omitempty"`
	PayURI     string          `json:"payUri,omitempty"`
	Evidence   string          `json:"evidence"`
	Warning    string          `json:"warning"`
	Quote      quote.Quote     `json:"quote"`
	Mint       Mint            `json:"mint"`
	Shelf      Ask             `json:"shelf"`
	Raw        json.RawMessage `json:"raw,omitempty"`
	ScriptHash string          `json:"scriptHash,omitempty"`
	Hex        string          `json:"hex,omitempty"`
	Layout     string          `json:"layout,omitempty"`
	Hit        bool            `json:"hit"`
	Hold       *Hold           `json:"hold,omitempty"`
	Outpoint   string          `json:"outpoint,omitempty"`
}

func LabelBytes(name string) []byte {
	n := apex(name)
	sum := sha256.Sum256([]byte(labelDomain + n))
	return sum[:]
}

func apex(raw string) string {
	n := strings.ToLower(strings.TrimSpace(raw))
	n = strings.TrimPrefix(n, "kas://")
	n = strings.TrimSuffix(n, "/")
	n = strings.TrimSuffix(n, ".kas")
	if i := strings.IndexByte(n, '.'); i >= 0 {
		n = n[:i]
	}
	return n
}

func DisplayName(raw string) string {
	n := apex(raw)
	if n == "" {
		return ""
	}
	return n + ".kas"
}

func ResetCovenantForTest() {
	skipChain = true
	tplMu.Lock()
	tplCached = nil
	tplMu.Unlock()
}

func kasNameArtifact() string {
	cands := []string{artifactRel}
	if wd, err := os.Getwd(); err == nil {
		cands = append(cands, filepath.Join(wd, artifactRel), filepath.Join(wd, "..", "..", artifactRel))
	}
	if exe, err := os.Executable(); err == nil {
		cands = append(cands, filepath.Join(filepath.Dir(exe), artifactRel))
	}
	for _, p := range cands {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return artifactRel
}

func templateScript() ([]byte, error) {
	tplMu.Lock()
	if tplCached != nil {
		out := append([]byte(nil), tplCached...)
		tplMu.Unlock()
		return out, nil
	}
	tplMu.Unlock()
	b, err := os.ReadFile(kasNameArtifact())
	if err != nil {
		return nil, err
	}
	var art struct {
		Contracts map[string]struct {
			Compiled struct {
				Bytecode []byte `json:"bytecode"`
			} `json:"compiled"`
		} `json:"contracts"`
	}
	if json.Unmarshal(b, &art) != nil {
		return nil, fmt.Errorf("KasName artifact")
	}
	c, ok := art.Contracts["KasName"]
	if !ok || len(c.Compiled.Bytecode) < labelOffset+labelLen {
		return nil, fmt.Errorf("KasName bytecode")
	}
	tplMu.Lock()
	tplCached = append([]byte(nil), c.Compiled.Bytecode...)
	tplMu.Unlock()
	return append([]byte(nil), c.Compiled.Bytecode...), nil
}

func RedeemFor(name string) ([]byte, error) {
	tpl, err := templateScript()
	if err != nil {
		return nil, err
	}
	out := append([]byte(nil), tpl...)
	lab := LabelBytes(name)
	copy(out[labelOffset:labelOffset+labelLen], lab)
	return out, nil
}

func P2SHFor(name string) (addr, hashHex string, redeem []byte, err error) {
	redeem, err = RedeemFor(name)
	if err != nil {
		return "", "", nil, err
	}
	sum := blake2b.Sum256(redeem)
	addr = kasaddr.Encode("kaspa", sum[:], kasaddr.VersionScriptHash)
	return addr, hex.EncodeToString(sum[:]), redeem, nil
}

// ResolveCovenant never calls KNS. The name is a P2SH from KasName.sil + label.
// First UTXO at that address is the name. api.kaspa.org is L1, not an inscription index.
func ResolveCovenant(raw string) *Result {
	raw = strings.TrimSpace(raw)
	disp := DisplayName(raw)
	out := &Result{
		Query:    raw,
		Name:     disp,
		Evidence: "live",
		Quote:    FundQuote(),
		Mint:     FundMint(),
		Shelf:    AskFor(disp),
		Layout:   "KasName { label, claimed, owner } — own UTXO only",
	}
	if disp == "" {
		out.Evidence = "roadmap"
		out.Warning = "Type a name. A kasdomain is a covenant output, not a KNS inscription."
		return out
	}
	addr, hashHex, redeem, err := P2SHFor(disp)
	if err != nil {
		out.Evidence = "roadmap"
		out.Warning = "KasName.sil is not compiled in this tree. " + err.Error()
		return out
	}
	out.PayURI = addr
	out.ScriptHash = hashHex
	out.Hex = hex.EncodeToString(redeem)
	out.Hold = LookupHold(disp)
	out.Warning = "kasdomain covenant is live on L1. This P2SH is KasName.sil plus this label. First funded output is the name. No yearly bill. Not KNS. Anyone can recompute the script."
	if out.Hold != nil {
		out.Warning = out.Hold.Note
	}
	if skipChain {
		out.Hit = false
		return out
	}
	utxos, err := chain.AddressUTXOs(addr)
	if err != nil {
		out.Warning += " Could not ask api.kaspa.org: " + err.Error()
		return out
	}
	if len(utxos) == 0 {
		out.Hit = false
		out.Warning = fmt.Sprintf("No kasdomain covenant output for %s yet. Fund the P2SH once (200 KAS mint: 100 name, 100 growth). First UTXO wins. No yearly rent. Inscriptions and KNS are out of scope.", disp)
		return out
	}
	out.Hit = true
	out.Outpoint = utxos[0].TxID + ":" + fmt.Sprint(utxos[0].Index)
	if who := HolderOf(disp); who != "" {
		out.Owner = who
	}
	rawJSON, _ := json.Marshal(utxos)
	out.Raw = rawJSON
	out.Warning = fmt.Sprintf("%s has a UTXO at this exact script. That is L1, not a KNS inscription. The kaspa:p… is the name lock, not the minter wallet.", disp)
	return out
}
