package names

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gramlane/internal/appenv"
	"gramlane/internal/quote"
)

const (
	HoldSale     = "sale"
	HoldRetailer = "retailer"
	HoldSold     = "sold"
)

// Hold is a desk reservation for the Kaspa growth vault.
// Not an L1 lock. First funded P2SH still wins on chain. We do not mass-mint UTXOs.
type Hold struct {
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	KAS   uint64 `json:"kas"`
	Sompi uint64 `json:"sompi"`
	Owner string `json:"owner"`
	Note  string `json:"note"`
	When  string `json:"when,omitempty"`
	Buyer string `json:"buyer,omitempty"`
	Sold  string `json:"sold,omitempty"`
	Tx    string `json:"tx,omitempty"`
}

type reserveBook struct {
	Seeded bool            `json:"seeded"`
	Owner  string          `json:"owner"`
	Holds  map[string]Hold `json:"holds"`
}

var (
	resMu   sync.Mutex
	resLive *reserveBook
	resPath = appenv.File("kasdomain-reserve.json")
)

func ResetReserveForTest(dir string) {
	resMu.Lock()
	defer resMu.Unlock()
	resPath = filepath.Join(dir, "kasdomain-reserve.json")
	resLive = nil
}

func SalePriceKAS(name string) uint64 {
	n := apex(name)
	L := utf8.RuneCountInString(n)
	switch {
	case L <= 1:
		return 1000
	case L == 2:
		return 500
	case L == 3:
		return 250
	default:
		return 30
	}
}

func LookupHold(name string) *Hold {
	disp := DisplayName(name)
	resMu.Lock()
	defer resMu.Unlock()
	loadResLocked()
	if resLive == nil || resLive.Holds == nil {
		return nil
	}
	h, ok := resLive.Holds[disp]
	if !ok {
		return nil
	}
	cp := h
	return &cp
}

func SeedVault() error {
	resMu.Lock()
	defer resMu.Unlock()
	loadResLocked()
	if resLive != nil && resLive.Seeded && len(resLive.Holds) > 0 {
		return nil
	}
	owner := VaultAddress()
	b := &reserveBook{
		Seeded: true,
		Owner:  owner,
		Holds:  map[string]Hold{},
	}
	now := time.Now().UTC().Format(time.RFC3339)
	put := func(raw, kind string) {
		disp := DisplayName(raw)
		if disp == "" {
			return
		}
		kas := SalePriceKAS(disp)
		note := "Reserved for the Kaspa growth vault. Buy for " + fmt.Sprintf("%d", kas) + " KAS. Desk hold — first L1 UTXO still wins if someone funds the P2SH. We do not mass-mint outputs."
		if kind == HoldRetailer {
			note = "Not for sale. Held for the real " + disp + " — giant industry or company, not a squatter. Contact the growth vault."
		}
		b.Holds[disp] = Hold{
			Name:  disp,
			Kind:  kind,
			KAS:   kas,
			Sompi: kas * quote.SompiPerKAS,
			Owner: owner,
			Note:  note,
			When:  now,
		}
	}
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	for i := 0; i < len(alphabet); i++ {
		put(alphabet[i:i+1], HoldSale)
		for j := 0; j < len(alphabet); j++ {
			put(string([]byte{alphabet[i], alphabet[j]}), HoldSale)
		}
	}
	for _, w := range retailerIndustries {
		put(w, HoldRetailer)
	}
	for _, w := range retailerCompanies {
		put(w, HoldRetailer)
	}
	resLive = b
	return saveResLocked()
}

func BuyReserved(name, buyer, tx string) (*Hold, error) {
	disp := DisplayName(name)
	buyer = strings.TrimSpace(buyer)
	tx = strings.TrimSpace(tx)
	if !strings.HasPrefix(buyer, "kaspa:") {
		return nil, fmt.Errorf("buyer must be a kaspa address")
	}
	resMu.Lock()
	loadResLocked()
	if resLive == nil || resLive.Holds == nil {
		resMu.Unlock()
		return nil, fmt.Errorf("not reserved")
	}
	h, ok := resLive.Holds[disp]
	if !ok {
		resMu.Unlock()
		return nil, fmt.Errorf("not reserved")
	}
	if h.Kind == HoldRetailer {
		resMu.Unlock()
		return &h, fmt.Errorf("not for sale — reserved for the real %s", disp)
	}
	if h.Kind != HoldSale {
		resMu.Unlock()
		return &h, fmt.Errorf("not for sale")
	}
	h.Kind = HoldSold
	h.Buyer = buyer
	h.Tx = tx
	h.Sold = time.Now().UTC().Format(time.RFC3339)
	h.Note = "Sold from the growth vault. Custody moved on this desk. L1 name UTXO still needs a KasName spend if it was funded."
	resLive.Holds[disp] = h
	owner := h.Owner
	err := saveResLocked()
	resMu.Unlock()
	if err != nil {
		return &h, err
	}
	Unhold(owner, disp)
	if err := Receive(buyer, disp); err != nil {
		h.Note += " " + err.Error()
		return &h, err
	}
	cp := h
	return &cp, nil
}

func loadResLocked() {
	if resLive != nil {
		return
	}
	raw, err := os.ReadFile(resPath)
	if err != nil {
		resLive = &reserveBook{Owner: VaultAddress(), Holds: map[string]Hold{}}
		return
	}
	var b reserveBook
	if json.Unmarshal(raw, &b) != nil || b.Holds == nil {
		resLive = &reserveBook{Owner: VaultAddress(), Holds: map[string]Hold{}}
		return
	}
	resLive = &b
}

func saveResLocked() error {
	if err := os.MkdirAll(filepath.Dir(resPath), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(resLive, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(resPath, raw, 0o644)
}
