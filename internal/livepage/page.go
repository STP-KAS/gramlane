// Package livepage stores the words on a Kasdomain’s Gramlane page.
package livepage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gramlane/internal/names"
)

var path = "data/live-pages.json"

type Page struct {
	Name     string `json:"name"`
	Headline string `json:"headline"`
	About    string `json:"about"`
	PayNote  string `json:"payNote,omitempty"`
}

var (
	mu   sync.Mutex
	book map[string]Page
)

func Get(name string) *Page {
	name = names.Normalize(name)
	load()
	mu.Lock()
	defer mu.Unlock()
	p, ok := book[name]
	if !ok {
		return nil
	}
	cp := p
	return &cp
}

// Ensure writes a first living page only when the name has none.
func Ensure(name, headline, about, payNote string) Page {
	if p := Get(name); p != nil && (p.Headline != "" || p.About != "") {
		return *p
	}
	return Save(name, headline, about, payNote)
}

func Save(name, headline, about, payNote string) Page {
	name = names.Normalize(name)
	p := Page{
		Name:     name,
		Headline: strings.TrimSpace(headline),
		About:    strings.TrimSpace(about),
		PayNote:  strings.TrimSpace(payNote),
	}
	load()
	mu.Lock()
	defer mu.Unlock()
	if book == nil {
		book = map[string]Page{}
	}
	book[name] = p
	_ = persist()
	return p
}

func load() {
	mu.Lock()
	defer mu.Unlock()
	if book != nil {
		return
	}
	book = map[string]Page{}
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	_ = json.Unmarshal(b, &book)
}

func persist() error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(book, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
