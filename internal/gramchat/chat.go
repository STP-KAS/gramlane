// Package gramchat is encrypted notes paid in grams.
// Ciphertext only. The passphrase never leaves the browser.
// Not Telegram Messenger. Not KaChat E2E.
package gramchat

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gramlane/internal/appenv"
)

type Note struct {
	Room  string `json:"room"`
	When  string `json:"when"`
	From  string `json:"from,omitempty"`
	Nonce string `json:"nonce"`
	Box   string `json:"box"`
	Grams uint64 `json:"grams"`
}

var (
	mu   sync.Mutex
	live []Note
	path = appenv.File("telegram.json")
)

func ResetForTest(dir string) {
	mu.Lock()
	defer mu.Unlock()
	path = filepath.Join(dir, "telegram.json")
	live = nil
}

func Room(raw string) string {
	n := strings.ToLower(strings.TrimSpace(raw))
	n = strings.TrimPrefix(n, "#")
	if n == "" || n == "board" {
		return "board"
	}
	n = strings.TrimSuffix(n, ".kas")
	if n == "" {
		return "board"
	}
	return n + ".kas"
}

func Put(room, from, nonce, box string, grams uint64) (*Note, error) {
	room = Room(room)
	nonce = strings.ToLower(strings.TrimSpace(nonce))
	box = strings.ToLower(strings.TrimSpace(box))
	if _, err := hex.DecodeString(nonce); err != nil || len(nonce) != 24 {
		return nil, fmt.Errorf("nonce")
	}
	raw, err := hex.DecodeString(box)
	if err != nil || len(raw) < 16 {
		return nil, fmt.Errorf("box")
	}
	if utf8.RuneCountInString(box) > 8000 {
		return nil, fmt.Errorf("too long")
	}
	n := Note{
		Room:  room,
		When:  time.Now().UTC().Format(time.RFC3339),
		From:  strings.TrimSpace(from),
		Nonce: nonce,
		Box:   box,
		Grams: grams,
	}
	mu.Lock()
	defer mu.Unlock()
	loadLocked()
	live = append([]Note{n}, live...)
	if len(live) > 400 {
		live = live[:400]
	}
	if err := saveLocked(); err != nil {
		return nil, err
	}
	cp := n
	return &cp, nil
}

func List(room string) []Note {
	room = Room(room)
	mu.Lock()
	defer mu.Unlock()
	loadLocked()
	var out []Note
	for _, n := range live {
		if n.Room == room {
			out = append(out, n)
		}
	}
	return out
}

func loadLocked() {
	if live != nil {
		return
	}
	live = []Note{}
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
