// Package market is the Kasdomain marketplace.
// Shelf is USD. Settlement is KAS on L1. Not grams — desk jobs keep grams.
package market

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
)

var path = appenv.File("market.json")

type Listing struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Seller   string `json:"seller"`
	USDCents uint64 `json:"usdCents"`
	USD      string `json:"usd"`
	Sompi    uint64 `json:"sompi"`
	KAS      string `json:"kas"`
	Rate     string `json:"rate"`
	Status   string `json:"status"`
	When     string `json:"when"`
	Buyer    string `json:"buyer,omitempty"`
	Sold     string `json:"sold,omitempty"`
	Tx       string `json:"tx,omitempty"`
	Note     string `json:"note"`
}

type Book struct {
	Listings []Listing `json:"listings"`
}

var (
	mu   sync.Mutex
	live *Book
)

func List() []Listing {
	mu.Lock()
	defer mu.Unlock()
	b := open()
	out := make([]Listing, len(b.Listings))
	copy(out, b.Listings)
	return out
}

func Open(name, seller string, usdCents uint64) (*Listing, error) {
	name = names.DisplayName(name)
	seller = strings.TrimSpace(seller)
	if usdCents == 0 {
		return nil, fmt.Errorf("price in USD")
	}
	if !strings.HasPrefix(seller, "kaspa:") {
		return nil, fmt.Errorf("seller must be a kaspa address")
	}
	if !names.Holds(seller, name) {
		return nil, fmt.Errorf("only a wallet that holds %s can list it", name)
	}
	ask := names.Settle(usdCents)
	mu.Lock()
	defer mu.Unlock()
	b := open()
	for _, L := range b.Listings {
		if L.Name == name && L.Status == "open" {
			return nil, fmt.Errorf("%s is already listed", name)
		}
	}
	L := Listing{
		ID:       fmt.Sprintf("%d", time.Now().UnixNano()),
		Name:     name,
		Seller:   seller,
		USDCents: usdCents,
		USD:      ask.USD,
		Sompi:    ask.PaySompi,
		KAS:      ask.PayKAS,
		Rate:     ask.Rate,
		Status:   "open",
		When:     time.Now().UTC().Format(time.RFC3339),
		Note:     "Shelf is USD. Buyer sends KAS on L1, not grams. Sale moves custody here. The name UTXO still needs a KasName transfer spend to finish on L1.",
	}
	b.Listings = append([]Listing{L}, b.Listings...)
	if err := save(b); err != nil {
		return nil, err
	}
	cp := L
	return &cp, nil
}

func Buy(id, buyer, tx string) (*Listing, error) {
	id = strings.TrimSpace(id)
	buyer = strings.TrimSpace(buyer)
	tx = strings.TrimSpace(tx)
	if !strings.HasPrefix(buyer, "kaspa:") {
		return nil, fmt.Errorf("buyer must be a kaspa address")
	}
	mu.Lock()
	defer mu.Unlock()
	b := open()
	i := -1
	for n := range b.Listings {
		if b.Listings[n].ID == id {
			i = n
			break
		}
	}
	if i < 0 {
		return nil, fmt.Errorf("unknown listing")
	}
	L := &b.Listings[i]
	if L.Status != "open" {
		return L, fmt.Errorf("not for sale")
	}
	if strings.EqualFold(L.Seller, buyer) {
		return L, fmt.Errorf("you already own this listing")
	}
	L.Status = "sold"
	L.Buyer = buyer
	L.Tx = tx
	L.Sold = time.Now().UTC().Format(time.RFC3339)
	names.Unhold(L.Seller, L.Name)
	if err := names.Receive(buyer, L.Name); err != nil {
		L.Note = L.Note + " Buyer drawer: " + err.Error()
	}
	if err := save(b); err != nil {
		return L, err
	}
	cp := *L
	return &cp, nil
}

func open() *Book {
	if live != nil {
		return live
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		live = &Book{}
		return live
	}
	var b Book
	if json.Unmarshal(raw, &b) != nil {
		live = &Book{}
		return live
	}
	live = &b
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
