// Package pos is Gramlane point-of-sale.
// Customer spend pile is prepaid grams. Merchants invoice in grams.
// Digital = share a URL. On-location = scan the same URL as a QR.
// Not a dollar. Operator ledger until WorkCredit consume/transfer.
package pos

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gramlane/internal/seq"
)

var path = "data/pos.json"

type Invoice struct {
	ID       string `json:"id"`
	Item     string `json:"item"`
	Grams    uint64 `json:"grams"`
	Merchant string `json:"merchant"`
	PayTo    string `json:"payTo,omitempty"`
	Place    string `json:"place"`
	Status   string `json:"status"`
	When     string `json:"when"`
	PaidWhen string `json:"paidWhen,omitempty"`
	Paid     string `json:"paid,omitempty"`
	Payer    string `json:"payer,omitempty"`
}

type Merchant struct {
	Name  string `json:"name"`
	PayTo string `json:"payTo,omitempty"`
	Owed  uint64 `json:"owedGrams"`
}

type Book struct {
	Invoices  []Invoice  `json:"invoices"`
	Merchants []Merchant `json:"merchants"`
}

var (
	mu   sync.Mutex
	live *Book
)

func Create(item string, grams uint64, merchant, payTo, place string) (*Invoice, error) {
	item = strings.TrimSpace(item)
	merchant = strings.TrimSpace(merchant)
	payTo = strings.TrimSpace(payTo)
	place = strings.ToLower(strings.TrimSpace(place))
	if item == "" {
		item = "sale"
	}
	if merchant == "" {
		merchant = "till"
	}
	if grams == 0 {
		return nil, fmt.Errorf("grams")
	}
	switch place {
	case "digital", "location", "both":
	default:
		place = "both"
	}
	inv := Invoice{
		ID:       newID(),
		Item:     item,
		Grams:    grams,
		Merchant: merchant,
		PayTo:    payTo,
		Place:    place,
		Status:   "open",
		When:     time.Now().UTC().Format(time.RFC3339),
	}
	mu.Lock()
	defer mu.Unlock()
	b := openLocked()
	b.Invoices = append([]Invoice{inv}, b.Invoices...)
	if len(b.Invoices) > 200 {
		b.Invoices = b.Invoices[:200]
	}
	touchMerchant(b, merchant, payTo, 0)
	if err := save(b); err != nil {
		return nil, err
	}
	cp := inv
	return &cp, nil
}

func Get(id string) (*Invoice, bool) {
	id = strings.TrimSpace(id)
	mu.Lock()
	defer mu.Unlock()
	b := openLocked()
	for i := range b.Invoices {
		if b.Invoices[i].ID == id {
			cp := b.Invoices[i]
			return &cp, true
		}
	}
	return nil, false
}

func List() []Invoice {
	mu.Lock()
	defer mu.Unlock()
	b := openLocked()
	out := make([]Invoice, len(b.Invoices))
	copy(out, b.Invoices)
	return out
}

func Merchants() []Merchant {
	mu.Lock()
	defer mu.Unlock()
	b := openLocked()
	out := make([]Merchant, len(b.Merchants))
	copy(out, b.Merchants)
	return out
}

func Pay(id, payer string) (*Invoice, error) {
	id = strings.TrimSpace(id)
	mu.Lock()
	defer mu.Unlock()
	b := openLocked()
	i := -1
	for n := range b.Invoices {
		if b.Invoices[n].ID == id {
			i = n
			break
		}
	}
	if i < 0 {
		return nil, fmt.Errorf("unknown invoice")
	}
	inv := &b.Invoices[i]
	if inv.Status == "paid" {
		cp := *inv
		return &cp, fmt.Errorf("already paid")
	}
	if _, err := seq.BurnGrams("pos:"+inv.ID, inv.Grams, "grams"); err != nil {
		return inv, err
	}
	inv.Status = "paid"
	inv.Paid = "grams"
	inv.Payer = strings.TrimSpace(payer)
	inv.PaidWhen = time.Now().UTC().Format(time.RFC3339)
	touchMerchant(b, inv.Merchant, inv.PayTo, inv.Grams)
	if err := save(b); err != nil {
		return inv, err
	}
	cp := *inv
	return &cp, nil
}

func PayURL(origin, id string) string {
	origin = strings.TrimRight(origin, "/")
	if origin == "" {
		origin = "http://127.0.0.1:8081"
	}
	return origin + "/pay/" + id
}

func openLocked() *Book {
	if live != nil {
		return live
	}
	b, err := os.ReadFile(path)
	if err != nil {
		live = &Book{}
		return live
	}
	var book Book
	if json.Unmarshal(b, &book) != nil {
		live = &Book{}
		return live
	}
	live = &book
	return live
}

func save(b *Book) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func touchMerchant(b *Book, name, payTo string, add uint64) {
	for i := range b.Merchants {
		if strings.EqualFold(b.Merchants[i].Name, name) {
			b.Merchants[i].Owed += add
			if payTo != "" {
				b.Merchants[i].PayTo = payTo
			}
			return
		}
	}
	b.Merchants = append(b.Merchants, Merchant{Name: name, PayTo: payTo, Owed: add})
}

func newID() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func ResetForTest(dir string) {
	mu.Lock()
	defer mu.Unlock()
	path = filepath.Join(dir, "pos.json")
	live = nil
}
