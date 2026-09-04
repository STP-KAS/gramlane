// Package prior is first-use evidence for a work, not a patent office.
// Idea from Kary Oberbrunner's TEDx: timestamp a hash on a public ledger.
// Honest: this proves a hash existed at a time. It is not a copyright, not a
// USPTO filing, not Instant IP. First funded P2SH of the hash is first-to-file
// on this template. Grams pay the desk stamp. KAS to the lock is the L1 time.
package prior

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gramlane/internal/appenv"
	"gramlane/internal/chain"
	"gramlane/internal/names"
	"gramlane/internal/quote"
)

const Tag = "kasprior:v1:"

type Record struct {
	Hash    string `json:"hash"`
	Title   string `json:"title,omitempty"`
	Owner   string `json:"owner,omitempty"`
	Name    string `json:"name,omitempty"`
	When    string `json:"when"`
	Tx      string `json:"tx,omitempty"`
	Lock    string `json:"lock"`
	Grams   uint64 `json:"grams,omitempty"`
	Hit     bool   `json:"hit"`
	Note    string `json:"note"`
	Excerpt string `json:"excerpt,omitempty"`
}

var (
	mu      sync.Mutex
	live    []Record
	path    = appenv.File("prior-art.json")
	skipNet bool
)

func ResetForTest(dir string) {
	mu.Lock()
	defer mu.Unlock()
	path = filepath.Join(dir, "prior-art.json")
	live = nil
	skipNet = true
}

func HashOf(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}

func NormalizeHash(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	h = strings.TrimPrefix(h, "0x")
	if len(h) != 64 {
		return ""
	}
	for _, c := range h {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return ""
		}
	}
	return h
}

func LockFor(hash string) (addr, scriptHash string, err error) {
	h := NormalizeHash(hash)
	if h == "" {
		return "", "", fmt.Errorf("sha256 hex")
	}
	addr, scriptHash, _, err = names.P2SHForTagged(Tag, h)
	return addr, scriptHash, err
}

func Lookup(hash string) *Record {
	h := NormalizeHash(hash)
	if h == "" {
		return nil
	}
	mu.Lock()
	defer mu.Unlock()
	loadLocked()
	for i := range live {
		if live[i].Hash == h {
			cp := live[i]
			paint(&cp)
			return &cp
		}
	}
	r := &Record{
		Hash: h,
		Note: "No desk stamp yet. The lock below is still the L1 first-to-file slot.",
	}
	paint(r)
	return r
}

func List() []Record {
	mu.Lock()
	defer mu.Unlock()
	loadLocked()
	out := make([]Record, len(live))
	copy(out, live)
	for i := range out {
		paint(&out[i])
	}
	return out
}

func File(hash, title, owner, name, tx, excerpt string, grams uint64) (*Record, error) {
	h := NormalizeHash(hash)
	if h == "" {
		return nil, fmt.Errorf("hash the work (sha256 hex)")
	}
	owner = strings.TrimSpace(owner)
	if owner != "" && !strings.HasPrefix(owner, "kaspa:") {
		return nil, fmt.Errorf("kaspa address")
	}
	title = strings.TrimSpace(title)
	if utf8len(title) > 80 {
		title = string([]rune(title)[:80])
	}
	excerpt = strings.TrimSpace(excerpt)
	if utf8len(excerpt) > 280 {
		excerpt = string([]rune(excerpt)[:280])
	}
	name = names.DisplayName(name)
	lock, _, err := LockFor(h)
	if err != nil {
		return nil, err
	}
	r := Record{
		Hash:    h,
		Title:   title,
		Owner:   owner,
		Name:    name,
		When:    time.Now().UTC().Format(time.RFC3339),
		Tx:      strings.TrimSpace(tx),
		Lock:    lock,
		Grams:   grams,
		Excerpt: excerpt,
		Note:    "Desk stamp. Not a patent. Not a copyright registration. The hash existed here at this time. Fund the lock on L1 for first UTXO.",
	}
	mu.Lock()
	defer mu.Unlock()
	loadLocked()
	for i := range live {
		if live[i].Hash == h {
			if live[i].Owner != "" && owner != "" && !strings.EqualFold(live[i].Owner, owner) {
				return nil, fmt.Errorf("this hash is already stamped by another wallet on this desk")
			}
			if title != "" {
				live[i].Title = title
			}
			if owner != "" {
				live[i].Owner = owner
			}
			if name != "" {
				live[i].Name = name
			}
			if r.Tx != "" {
				live[i].Tx = r.Tx
			}
			if excerpt != "" {
				live[i].Excerpt = excerpt
			}
			live[i].Grams += grams
			paint(&live[i])
			_ = saveLocked()
			cp := live[i]
			return &cp, nil
		}
	}
	live = append([]Record{r}, live...)
	if len(live) > 200 {
		live = live[:200]
	}
	paint(&live[0])
	if err := saveLocked(); err != nil {
		return nil, err
	}
	cp := live[0]
	return &cp, nil
}

func paint(r *Record) {
	if r.Lock == "" {
		if a, _, err := LockFor(r.Hash); err == nil {
			r.Lock = a
		}
	}
	if skipNet || r.Lock == "" {
		return
	}
	utxos, err := chain.AddressUTXOs(r.Lock)
	if err != nil || len(utxos) == 0 {
		r.Hit = false
		return
	}
	r.Hit = true
	if r.Tx == "" {
		r.Tx = utxos[0].TxID
	}
}

func loadLocked() {
	if live != nil {
		return
	}
	live = []Record{}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(b, &live)
}

func saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(live, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func utf8len(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func DustSompi() uint64 {
	return quote.KaswareMinSompi
}
