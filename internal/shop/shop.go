// Package shop hangs a page and a till on a kasdomain.
// The name is L1 (P2SH). Words and menu live on this desk until a content covenant exists.
// Till tickets settle in grams (recurrent sequencing). Acquiring the name stays KAS.
package shop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gramlane/internal/appenv"
	"gramlane/internal/names"
	"gramlane/internal/pos"
	"gramlane/internal/quote"
)

type Item struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	USDCents uint64 `json:"usdCents"`
	USD      string `json:"usd"`
	Grams    uint64 `json:"grams"`
	KAS      string `json:"kas,omitempty"`
	Sompi    uint64 `json:"sompi,omitempty"`
	PaySompi uint64 `json:"paySompi,omitempty"`
	Rate     string `json:"rate,omitempty"`
}

type Shop struct {
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	PayTo    string `json:"payTo,omitempty"`
	P2SH     string `json:"p2sh,omitempty"`
	Headline string `json:"headline"`
	About    string `json:"about"`
	PayNote  string `json:"payNote,omitempty"`
	Items    []Item `json:"items"`
	When     string `json:"when"`
}

type Storefront struct {
	Name       string         `json:"name"`
	P2SH       string         `json:"p2sh,omitempty"`
	Hit        bool           `json:"hit"`
	Evidence   string         `json:"evidence"`
	Warning    string         `json:"warning"`
	Shop       *Shop          `json:"shop,omitempty"`
	Rate       string         `json:"rate"`
	TillUnit   string         `json:"tillUnit"`
	NameSettle string         `json:"nameSettle"`
	Web4       map[string]any `json:"web4"`
}

var (
	mu   sync.Mutex
	live map[string]*Shop
	path = appenv.File("kasdomain-shops.json")
)

func ResetForTest(dir string) {
	mu.Lock()
	defer mu.Unlock()
	path = filepath.Join(dir, "kasdomain-shops.json")
	live = nil
}

func Hang(name, owner, payTo, headline, about string) (*Shop, error) {
	disp := names.DisplayName(name)
	owner = strings.TrimSpace(owner)
	payTo = strings.TrimSpace(payTo)
	if disp == "" {
		return nil, fmt.Errorf("name")
	}
	if !strings.HasPrefix(owner, "kaspa:") {
		return nil, fmt.Errorf("kaspa address")
	}
	if !names.Holds(owner, disp) {
		return nil, fmt.Errorf("only the wallet that holds %s can hang a shop on it", disp)
	}
	if payTo == "" {
		payTo = owner
	}
	p2sh, _, _, err := names.P2SHFor(disp)
	if err != nil {
		return nil, err
	}
	headline = strings.TrimSpace(headline)
	about = strings.TrimSpace(about)
	if headline == "" {
		headline = disp
	}
	mu.Lock()
	defer mu.Unlock()
	loadLocked()
	sh := live[disp]
	if sh == nil {
		sh = &Shop{Name: disp, When: time.Now().UTC().Format(time.RFC3339)}
		live[disp] = sh
	}
	sh.Owner = owner
	sh.PayTo = payTo
	sh.P2SH = p2sh
	sh.Headline = headline
	sh.About = about
	if sh.PayNote == "" {
		sh.PayNote = "USD on the shelf. Pay from the gram jar, or send KAS. Not a peg. Not KNS."
	}
	if err := saveLocked(); err != nil {
		return nil, err
	}
	cp := *sh
	cp.Items = append([]Item(nil), sh.Items...)
	return &cp, nil
}

func AddItem(name, owner, label string, cents uint64) (*Shop, error) {
	disp := names.DisplayName(name)
	owner = strings.TrimSpace(owner)
	label = strings.TrimSpace(label)
	if label == "" {
		return nil, fmt.Errorf("what is it?")
	}
	if cents == 0 {
		return nil, fmt.Errorf("price in USD")
	}
	if !names.Holds(owner, disp) {
		return nil, fmt.Errorf("only the wallet that holds %s can price it", disp)
	}
	mu.Lock()
	sh := peekLocked(disp)
	mu.Unlock()
	if sh == nil {
		if _, err := Hang(disp, owner, owner, disp, ""); err != nil {
			return nil, err
		}
	}
	it, err := priced(label, cents)
	if err != nil {
		return nil, err
	}
	mu.Lock()
	defer mu.Unlock()
	loadLocked()
	sh = live[disp]
	if sh == nil {
		return nil, fmt.Errorf("hang the shop first")
	}
	it.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	sh.Items = append(sh.Items, it)
	if err := saveLocked(); err != nil {
		return nil, err
	}
	cp := *sh
	cp.Items = append([]Item(nil), sh.Items...)
	return &cp, nil
}

func Ticket(name, itemID, place string) (*pos.Invoice, error) {
	sh := Get(name)
	if sh == nil {
		return nil, fmt.Errorf("no shop on %s yet", names.DisplayName(name))
	}
	itemID = strings.TrimSpace(itemID)
	var it *Item
	for i := range sh.Items {
		if sh.Items[i].ID == itemID {
			it = &sh.Items[i]
			break
		}
	}
	if it == nil {
		return nil, fmt.Errorf("unknown item")
	}
	inv, err := pos.Create(it.Label, it.Grams, sh.Name, sh.PayTo, place)
	if err != nil {
		return nil, err
	}
	return pos.SetShelf(inv.ID, quote.FormatUSDAmount(it.USDCents), "USD", it.Rate, it.Label+" "+it.USD), nil
}

func Get(name string) *Shop {
	disp := names.DisplayName(name)
	mu.Lock()
	defer mu.Unlock()
	loadLocked()
	sh := live[disp]
	if sh == nil {
		return nil
	}
	cp := *sh
	cp.Items = append([]Item(nil), sh.Items...)
	return &cp
}

func List() []Shop {
	mu.Lock()
	defer mu.Unlock()
	loadLocked()
	out := make([]Shop, 0, len(live))
	for _, sh := range live {
		cp := *sh
		cp.Items = append([]Item(nil), sh.Items...)
		out = append(out, cp)
	}
	return out
}

func View(raw string) *Storefront {
	disp := names.DisplayName(raw)
	rate := names.LoadSign().KasInFiat
	v := &Storefront{
		Name:       disp,
		Evidence:   "compiled",
		Rate:       rate,
		TillUnit:   "gram",
		NameSettle: "kas",
		Web4: map[string]any{
			"named":     true,
			"loginless": true,
			"l2":        false,
			"dns":       false,
			"kns":       false,
			"vprogs":    "roadmap",
			"toccata":   "live",
			"note":      "The name is a covenant output. The page and menu are this desk. Agents pay without an account.",
		},
	}
	if disp == "" {
		v.Evidence = "roadmap"
		v.Warning = "Type a kasdomain. A shop is a name people say, a page they open, a till they pay."
		return v
	}
	addr, _, _, err := names.P2SHFor(disp)
	if err != nil {
		v.Evidence = "roadmap"
		v.Warning = "KasName.sil is not compiled here. " + err.Error()
		return v
	}
	v.P2SH = addr
	res := names.ResolveCovenant(disp)
	if res != nil {
		v.Hit = res.Hit
		v.P2SH = res.PayURI
	}
	if sh := Get(disp); sh != nil {
		v.Shop = sh
		v.Warning = "Name is L1. Words and prices are this desk — not DNS, not IPFS, not a vProg. Till is grams. Acquiring the name is KAS. USD is the board sign. Not KNS."
		return v
	}
	if v.Hit {
		v.Warning = disp + " is funded on L1. No shop hung on this desk yet. The holder opens one from Names."
	} else {
		v.Warning = "No kasdomain output yet. Fund " + disp + " with KAS. Then hang a shop. Not KNS."
	}
	return v
}

func peekLocked(disp string) *Shop {
	loadLocked()
	return live[disp]
}

func priced(label string, cents uint64) (Item, error) {
	sign := names.LoadSign()
	amount := quote.FormatUSDAmount(cents)
	bill, err := quote.FromFiat(amount, "USD", sign.KasInFiat)
	if err != nil {
		return Item{}, err
	}
	if bill.Grams == 0 {
		return Item{}, fmt.Errorf("grams")
	}
	pay := quote.FloorPay(bill.Sompi)
	kas := bill.KASText
	if pay != bill.Sompi {
		if c, e := quote.FromSompi(pay); e == nil {
			kas = c.KASText
		}
	}
	return Item{
		Label:    label,
		USDCents: cents,
		USD:      quote.FormatUSD(cents),
		Grams:    bill.Grams,
		KAS:      kas,
		Sompi:    bill.Sompi,
		PaySompi: pay,
		Rate:     sign.KasInFiat,
	}, nil
}

func loadLocked() {
	if live != nil {
		return
	}
	live = map[string]*Shop{}
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
