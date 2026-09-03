// Package seq is the sequencer prepaid-gram ledger.
// Backing: 0.5 KAS sale + WorkCredit P2SH UTXO on L1.
// Burns here are operator accounting until consume() spends that UTXO.
package seq

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gramlane/internal/chain"
	"gramlane/internal/genesis"
)

var path = "data/ledger.json"

type Burn struct {
	When  string `json:"when"`
	Job   string `json:"job"`
	Grams uint64 `json:"grams"`
	Paid  string `json:"paid"`
	Left  uint64 `json:"remaining"`
}

type Mint struct {
	When  string `json:"when"`
	Tx    string `json:"tx"`
	Grams uint64 `json:"grams"`
	Sompi uint64 `json:"sompi"`
}

type Ledger struct {
	Holder     string `json:"holder"`
	Desk       string `json:"desk"`
	P2SH       string `json:"p2sh"`
	Lane       string `json:"lane"`
	Credits    uint64 `json:"credits"`
	Remaining  uint64 `json:"remaining"`
	SaleTx     string `json:"saleTx"`
	VoucherTx  string `json:"voucherTx,omitempty"`
	VoucherIdx int    `json:"voucherIndex"`
	OnChain    bool   `json:"voucherOnChain"`
	Note       string `json:"note"`
	Burns      []Burn `json:"burns"`
	Mints      []Mint `json:"mints,omitempty"`
}

var (
	mu      sync.Mutex
	live    *Ledger
	skipNet bool
)

func fresh() *Ledger {
	return &Ledger{
		Holder:    genesis.HolderAddress,
		Desk:      genesis.DeskAddress,
		Lane:      genesis.Lane,
		Credits:   genesis.Credits,
		Remaining: genesis.Credits,
		SaleTx:    genesis.SaleTx,
		Note:      "Prepaid 500000 grams from 0.5 KAS L1 sale. Burns are sequencer accounting until consume() spends the P2SH UTXO.",
	}
}

func openLocked() *Ledger {
	if live != nil {
		return live
	}
	live = load()
	if live == nil {
		live = fresh()
	}
	if !skipNet {
		refreshLocked(live)
	}
	_ = save(live)
	return live
}

func refreshLocked(l *Ledger) {
	addr, _, _, err := genesis.P2SH()
	if err != nil {
		return
	}
	l.P2SH = addr
	utxos, err := chain.AddressUTXOs(addr)
	if err != nil {
		return
	}
	l.OnChain = len(utxos) > 0
	if len(utxos) > 0 {
		l.VoucherTx = utxos[0].TxID
		l.VoucherIdx = utxos[0].Index
	}
}

func Open() *Ledger {
	mu.Lock()
	defer mu.Unlock()
	return openLocked()
}

func Snap() Ledger {
	mu.Lock()
	defer mu.Unlock()
	l := openLocked()
	cp := *l
	cp.Burns = append([]Burn(nil), l.Burns...)
	cp.Mints = append([]Mint(nil), l.Mints...)
	return cp
}

func CanBurn(grams uint64) bool {
	mu.Lock()
	defer mu.Unlock()
	l := openLocked()
	return grams > 0 && l.Remaining >= grams
}

func Accepts(paid string) bool {
	p := strings.ToLower(strings.TrimSpace(paid))
	switch p {
	case "grams", "prepaid", "work-credit", "kaspa-work-credit":
		return true
	}
	mu.Lock()
	defer mu.Unlock()
	l := openLocked()
	if p == strings.ToLower(l.SaleTx) {
		return true
	}
	if l.VoucherTx != "" && p == strings.ToLower(l.VoucherTx) {
		return true
	}
	return false
}

func BurnGrams(job string, grams uint64, paid string) (*Ledger, error) {
	mu.Lock()
	defer mu.Unlock()
	l := openLocked()
	if grams == 0 {
		return l, fmt.Errorf("grams")
	}
	if l.Remaining < grams {
		return l, fmt.Errorf("need %d grams, have %d", grams, l.Remaining)
	}
	l.Remaining -= grams
	l.Burns = append(l.Burns, Burn{
		When:  time.Now().UTC().Format(time.RFC3339),
		Job:   job,
		Grams: grams,
		Paid:  paid,
		Left:  l.Remaining,
	})
	if err := save(l); err != nil {
		return l, err
	}
	cp := *l
	cp.Burns = append([]Burn(nil), l.Burns...)
	cp.Mints = append([]Mint(nil), l.Mints...)
	return &cp, nil
}

func MintFromTx(tx string, grams, sompi uint64) (*Ledger, error) {
	tx = strings.ToLower(strings.TrimSpace(tx))
	if !chain.IsTxID(tx) {
		return nil, fmt.Errorf("txid")
	}
	if grams == 0 {
		return nil, fmt.Errorf("grams")
	}
	mu.Lock()
	defer mu.Unlock()
	l := openLocked()
	for _, m := range l.Mints {
		if m.Tx == tx {
			return l, fmt.Errorf("txid already minted")
		}
	}
	l.Credits += grams
	l.Remaining += grams
	l.Mints = append(l.Mints, Mint{
		When:  time.Now().UTC().Format(time.RFC3339),
		Tx:    tx,
		Grams: grams,
		Sompi: sompi,
	})
	if err := save(l); err != nil {
		return l, err
	}
	cp := *l
	cp.Burns = append([]Burn(nil), l.Burns...)
	cp.Mints = append([]Mint(nil), l.Mints...)
	return &cp, nil
}

func load() *Ledger {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var l Ledger
	if json.Unmarshal(b, &l) != nil || l.Credits == 0 {
		return nil
	}
	return &l
}

func save(l *Ledger) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}

// ResetForTest points the ledger at an isolated file and skips chain lookups.
func ResetForTest(dir string) {
	mu.Lock()
	defer mu.Unlock()
	path = filepath.Join(dir, "ledger.json")
	skipNet = true
	live = nil
}
