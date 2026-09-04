// Package feedback writes user notes to a folder next to the Kaspa apps.
// It never asks for seeds. Messages that look like recovery phrases are rejected.
package feedback

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const DefaultDir = `C:\Users\Remco\Documents\kaspa\feedback`

var wordRe = regexp.MustCompile(`[a-z]+`)

type Note struct {
	ID      string `json:"id"`
	App     string `json:"app"`
	At      string `json:"at"`
	Text    string `json:"text"`
	Contact string `json:"contact,omitempty"`
	Note    string `json:"note"`
}

func Dir() string {
	if d := strings.TrimSpace(os.Getenv("KASPA_FEEDBACK_DIR")); d != "" {
		return d
	}
	if d := strings.TrimSpace(os.Getenv("DATA_DIR")); d != "" {
		return filepath.Join(d, "feedback")
	}
	return DefaultDir
}

func LooksLikeSeed(s string) bool {
	low := strings.ToLower(s)
	if strings.Contains(low, "seed phrase") || strings.Contains(low, "recovery phrase") || strings.Contains(low, "private key") {
		// mentioning the topic is ok; a dump of 12/24 words is not
	}
	words := wordRe.FindAllString(low, -1)
	n := 0
	for _, w := range words {
		if len(w) >= 3 && len(w) <= 8 {
			n++
		}
	}
	// 12 or 24 consecutive-ish mnemonic dumps
	if n >= 12 && n <= 15 && len(words) <= 20 {
		return true
	}
	if n >= 23 && n <= 26 && len(words) <= 32 {
		return true
	}
	hex := 0
	for _, r := range strings.ReplaceAll(low, "0x", "") {
		if unicode.Is(unicode.Hex_Digit, r) {
			hex++
		}
	}
	if hex >= 64 && len(strings.TrimSpace(s)) < 200 {
		return true
	}
	return false
}

func Save(app, text, contact string) (Note, error) {
	text = strings.TrimSpace(text)
	contact = strings.TrimSpace(contact)
	if text == "" {
		return Note{}, fmt.Errorf("write something")
	}
	if len(text) > 8000 {
		return Note{}, fmt.Errorf("too long")
	}
	if LooksLikeSeed(text) || LooksLikeSeed(contact) {
		return Note{}, fmt.Errorf("do not paste a seed, private key, or password. we never ask for those")
	}
	app = strings.TrimSpace(app)
	if app == "" {
		app = "unknown"
	}
	now := time.Now().UTC()
	id := now.Format("20060102T150405Z")
	n := Note{
		ID:      id,
		App:     app,
		At:      now.Format(time.RFC3339),
		Text:    text,
		Contact: contact,
		Note:    "Stored locally. This site never DMs you. Optional contact is not a request for us to message first.",
	}
	root := filepath.Join(Dir(), app)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return Note{}, err
	}
	raw, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return Note{}, err
	}
	path := filepath.Join(root, id+".json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return Note{}, err
	}
	return n, nil
}
