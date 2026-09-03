// Package post is local message postage.
// Grams pay the stamp. This is not KaChat, not broadcast, not encrypted.
package post

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const path = "data/postage.json"

type Stamp struct {
	When   string `json:"when"`
	SHA256 string `json:"sha256"`
	From   string `json:"from,omitempty"`
	Text   string `json:"text"`
	Bytes  int    `json:"bytes"`
	Grams  uint64 `json:"grams"`
	Lane   string `json:"lane"`
	Note   string `json:"note"`
}

var (
	mu    sync.Mutex
	cache []Stamp
)

func StampMsg(from, text string, grams uint64) Stamp {
	text = strings.TrimSpace(text)
	if text == "" {
		text = "hello.kas"
	}
	if utf8.RuneCountInString(text) > 280 {
		r := []rune(text)
		text = string(r[:280])
	}
	sum := sha256.Sum256([]byte(text))
	s := Stamp{
		When:   time.Now().UTC().Format(time.RFC3339),
		SHA256: hex.EncodeToString(sum[:]),
		From:   strings.TrimSpace(from),
		Text:   text,
		Bytes:  len(text),
		Grams:  grams,
		Lane:   "MSG1",
		Note:   "Postage is sequenced inclusion of this envelope. Not KaChat. Not broadcast. Encryption stays with the messenger.",
	}
	mu.Lock()
	defer mu.Unlock()
	all := loadLocked()
	all = append([]Stamp{s}, all...)
	if len(all) > 50 {
		all = all[:50]
	}
	cache = all
	_ = saveLocked(all)
	return s
}

func List() []Stamp {
	mu.Lock()
	defer mu.Unlock()
	all := loadLocked()
	out := make([]Stamp, len(all))
	copy(out, all)
	return out
}

func loadLocked() []Stamp {
	if cache != nil {
		return cache
	}
	b, err := os.ReadFile(path)
	if err != nil {
		cache = []Stamp{}
		return cache
	}
	var all []Stamp
	if json.Unmarshal(b, &all) != nil {
		cache = []Stamp{}
		return cache
	}
	cache = all
	return cache
}

func saveLocked(all []Stamp) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
