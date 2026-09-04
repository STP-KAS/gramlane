package names

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gramlane/internal/appenv"
)

// Held is a kasdomain this wallet funded. Only Primary is the face (replaces kaspa:q…).
// The rest are in custody: you keep them, you can sell them, they do not log you in.
type Held struct {
	Name     string `json:"name"`
	P2SH     string `json:"p2sh"`
	When     string `json:"when"`
	Tx       string `json:"tx,omitempty"`
	USDCents uint64 `json:"usdCents"`
	USD      string `json:"usd"`
	KAS      string `json:"kas"`
	Sompi    uint64 `json:"sompi"`
	Face     bool   `json:"face"`
}

type WalletBook struct {
	Address string `json:"address"`
	First   string `json:"first"` // first registered; default face
	Primary string `json:"primary"`
	Names   []Held `json:"names"`
}

var (
	bookMu   sync.Mutex
	bookLive map[string]*WalletBook
	bookPath = appenv.File("kasdomain-held.json")
)

func loadBook() {
	if bookLive != nil {
		return
	}
	bookLive = map[string]*WalletBook{}
	b, err := os.ReadFile(bookPath)
	if err != nil {
		return
	}
	_ = json.Unmarshal(b, &bookLive)
}

func saveBook() error {
	if err := os.MkdirAll(filepath.Dir(bookPath), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(bookLive, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(bookPath, raw, 0o644)
}

func keyAddr(addr string) string {
	return strings.ToLower(strings.TrimSpace(addr))
}

func Holds(addr, name string) bool {
	b := Book(addr)
	if b == nil {
		return false
	}
	n := DisplayName(name)
	for _, h := range b.Names {
		if h.Name == n {
			return true
		}
	}
	return false
}

func Primary(addr string) string {
	b := Book(addr)
	if b == nil {
		return ""
	}
	return b.Primary
}

func Book(addr string) *WalletBook {
	bookMu.Lock()
	defer bookMu.Unlock()
	loadBook()
	k := keyAddr(addr)
	src := bookLive[k]
	if src == nil {
		return nil
	}
	cp := *src
	cp.Names = append([]Held(nil), src.Names...)
	for i := range cp.Names {
		paintHeld(&cp.Names[i], cp.Primary)
	}
	return &cp
}

// Record notes a funded kasdomain. First name on this wallet becomes the face.
func Record(addr, name, tx string) (*WalletBook, error) {
	if !strings.HasPrefix(strings.TrimSpace(addr), "kaspa:") {
		return nil, fmt.Errorf("kaspa address")
	}
	disp := DisplayName(name)
	if disp == "" {
		return nil, fmt.Errorf("name")
	}
	if h := LookupHold(disp); h != nil && h.Kind != HoldSold {
		if h.Kind == HoldRetailer {
			return nil, fmt.Errorf("not for sale — reserved for the real %s", disp)
		}
		if keyAddr(addr) != keyAddr(h.Owner) {
			return nil, fmt.Errorf("reserved. Buy %s from the growth vault", disp)
		}
	}
	p2sh, _, _, err := P2SHFor(disp)
	if err != nil {
		return nil, err
	}
	bookMu.Lock()
	defer bookMu.Unlock()
	loadBook()
	k := keyAddr(addr)
	b := bookLive[k]
	if b == nil {
		b = &WalletBook{Address: strings.TrimSpace(addr)}
		bookLive[k] = b
	}
	for _, h := range b.Names {
		if h.Name == disp {
			return snapshotLocked(b), nil
		}
	}
	h := Held{Name: disp, P2SH: p2sh, When: time.Now().UTC().Format(time.RFC3339), Tx: strings.TrimSpace(tx)}
	b.Names = append(b.Names, h)
	if b.First == "" {
		b.First = disp
		b.Primary = disp
		setFaceLocked(k, disp)
	}
	if err := saveBook(); err != nil {
		return nil, err
	}
	return snapshotLocked(b), nil
}

func paintHeld(h *Held, primary string) {
	a := AskFor(h.Name)
	h.USDCents = a.USDCents
	h.USD = a.USD
	h.KAS = a.PayKAS
	h.Sompi = a.PaySompi
	h.Face = h.Name == primary
}

func snapshotLocked(b *WalletBook) *WalletBook {
	cp := *b
	cp.Names = append([]Held(nil), b.Names...)
	for i := range cp.Names {
		paintHeld(&cp.Names[i], cp.Primary)
	}
	return &cp
}

func setFaceLocked(addrKey, name string) {
	if links == nil {
		loadLinks()
	}
	linkMu.Lock()
	if links == nil {
		links = map[string]string{}
	}
	if name == "" {
		delete(links, addrKey)
	} else {
		links[addrKey] = name
	}
	_ = saveLinksLocked()
	linkMu.Unlock()
}

// SetPrimary picks which held name replaces kaspa:q…  Others stay in custody.
func SetPrimary(addr, name string) error {
	disp := DisplayName(name)
	k := keyAddr(addr)
	bookMu.Lock()
	defer bookMu.Unlock()
	loadBook()
	b := bookLive[k]
	if b == nil {
		return fmt.Errorf("no kasdomains in this wallet yet")
	}
	ok := false
	for _, h := range b.Names {
		if h.Name == disp {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("%s is not in this wallet's drawer", disp)
	}
	b.Primary = disp
	setFaceLocked(k, disp)
	return saveBook()
}

func Unhold(addr, name string) {
	disp := DisplayName(name)
	k := keyAddr(addr)
	bookMu.Lock()
	defer bookMu.Unlock()
	loadBook()
	b := bookLive[k]
	if b == nil {
		return
	}
	out := b.Names[:0]
	for _, h := range b.Names {
		if h.Name != disp {
			out = append(out, h)
		}
	}
	b.Names = out
	if b.First == disp {
		if len(b.Names) > 0 {
			b.First = b.Names[0].Name
		} else {
			b.First = ""
		}
	}
	if b.Primary == disp {
		if b.First != "" {
			b.Primary = b.First
		} else if len(b.Names) > 0 {
			b.Primary = b.Names[0].Name
		} else {
			b.Primary = ""
		}
		setFaceLocked(k, b.Primary)
	}
	_ = saveBook()
}

func HoldsUnlocked(b *WalletBook, name string) bool {
	for _, h := range b.Names {
		if h.Name == name {
			return true
		}
	}
	return false
}

// Receive puts a bought name in custody. It becomes the face only if this wallet has none.
func Receive(addr, name string) error {
	b, err := Record(addr, name, "market")
	if err != nil {
		return err
	}
	if b.First != "" && b.Primary == DisplayName(name) && len(b.Names) > 1 {
		// Record already set primary if first; fine
	}
	return nil
}

func ResetBookForTest(dir string) {
	bookMu.Lock()
	defer bookMu.Unlock()
	bookPath = filepath.Join(dir, "kasdomain-held.json")
	bookLive = nil
	linksPath = filepath.Join(dir, "id-links.json")
	links = nil
}
