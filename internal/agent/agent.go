// Package agent is Gramlane's callable agent.
// Grams pay the call. Grok (xAI) if XAI_API_KEY is set; otherwise local tools.
package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gramlane/internal/framing"
	"gramlane/internal/names"
	"gramlane/internal/seq"
)

const (
	ModelGrok  = "grok-4.5"
	ModelLocal = "gramlane-local"
	xaiURL     = "https://api.x.ai/v1/chat/completions"
)

type Card struct {
	Type        string      `json:"type"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Image       string      `json:"image,omitempty"`
	Owner       string      `json:"owner,omitempty"`
	PayURI      string      `json:"payUri,omitempty"`
	DID         string      `json:"did,omitempty"`
	KasURI      string      `json:"kasUri,omitempty"`
	X402Note    string      `json:"x402Note"`
	Services    []Service   `json:"services"`
	Site        *names.Page `json:"site,omitempty"`
	Warning     string      `json:"warning"`
}

type Service struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
}

type Reply struct {
	Model   string `json:"model"`
	Text    string `json:"text"`
	Grok    bool   `json:"grok"`
	Name    string `json:"name,omitempty"`
	Warning string `json:"warning"`
}

func HasKey() bool {
	return strings.TrimSpace(os.Getenv("XAI_API_KEY")) != ""
}

func CardFor(raw string) (*Card, error) {
	p, err := names.Lookup(raw)
	if err != nil {
		return nil, err
	}
	desc := p.Bio
	if desc == "" {
		desc = p.Name + " is a Kaspa Name Service identity. Agents can resolve it and pay kaspa: to the owner. Uniqueness is indexer-backed, not consensus."
	}
	c := &Card{
		Type:        "https://eips.ethereum.org/EIPS/eip-8004#registration-v1",
		Name:        p.Name,
		Description: desc,
		Image:       p.Avatar,
		Owner:       p.Owner,
		PayURI:      p.PayURI,
		DID:         p.DID,
		KasURI:      p.KasURI,
		X402Note:    "Not Coinbase x402/USDC. HTTP 402 here asks for Work Credits (grams) or KAS fallback.",
		Site:        p,
		Warning:     "No on-chain agent NFT on Kaspa. This card is synthesized from the KNS indexer.",
		Services: []Service{
			{Name: "site", Endpoint: "/site/" + p.Name},
			{Name: "agent", Endpoint: "/agent?q=" + p.Name},
			{Name: "mcp", Endpoint: "/api/agent"},
			{Name: "pay", Endpoint: p.PayURI},
			{Name: "work-credits", Endpoint: "/api/seq"},
		},
	}
	return c, nil
}

func Ask(prompt, name string) (*Reply, error) {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		prompt = "What is Gramlane?"
	}
	if len(prompt) > 2000 {
		prompt = prompt[:2000]
	}
	name = strings.TrimSpace(name)
	if HasKey() {
		return grok(prompt, name)
	}
	return local(prompt, name), nil
}

func grok(prompt, name string) (*Reply, error) {
	sys := "You are Gramlane's agent on Kaspa L1. Shops take KAS. Grams pay apps (agent, postage, prior-art stamp, vault bump), not the shop drawer. Prior art is a hash timestamp, not a patent. Vaults lock KAS; grams pay the action. No dollar peg. No L2. Be concise and literal."
	if name != "" {
		if c, err := CardFor(name); err == nil {
			b, _ := json.Marshal(c)
			if len(b) > 1500 {
				b = b[:1500]
			}
			sys += "\nCurrent name card JSON: " + string(b)
		}
	}
	led := seq.Snap()
	sys += fmt.Sprintf("\nPrepaid grams remaining: %d / %d.", led.Remaining, led.Credits)
	body, _ := json.Marshal(map[string]any{
		"model": ModelGrok,
		"messages": []map[string]string{
			{"role": "system", "content": sys},
			{"role": "user", "content": prompt},
		},
		"max_tokens": 600,
	})
	req, err := http.NewRequest(http.MethodPost, xaiURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("XAI_API_KEY"))
	cli := &http.Client{Timeout: 45 * time.Second}
	res, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("xAI %s: %s", res.Status, truncate(string(raw), 180))
	}
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	text := ""
	if len(out.Choices) > 0 {
		text = out.Choices[0].Message.Content
	}
	if text == "" {
		return nil, fmt.Errorf("empty grok reply")
	}
	return &Reply{Model: ModelGrok, Text: text, Grok: true, Name: name, Warning: "Grok via api.x.ai. Grams paid the sequenced call, not the model weights."}, nil
}

func local(prompt, name string) *Reply {
	low := strings.ToLower(prompt)
	var b strings.Builder
	b.WriteString("local tools agent (set XAI_API_KEY for Grok on api.x.ai).\n\n")
	if name == "" {
		if i := strings.Index(low, ".kas"); i > 0 {
			start := i
			for start > 0 && (low[start-1] == '.' || low[start-1] == '-' || (low[start-1] >= 'a' && low[start-1] <= 'z') || (low[start-1] >= '0' && low[start-1] <= '9')) {
				start--
			}
			name = prompt[start : i+4]
		}
	}
	if name != "" {
		if c, err := CardFor(name); err == nil {
			fmt.Fprintf(&b, "%s owner %s\n", c.Name, c.Owner)
			if c.Site != nil && c.Site.Bio != "" {
				fmt.Fprintf(&b, "bio: %s\n", c.Site.Bio)
			}
			fmt.Fprintf(&b, "pay: %s\n\n", c.PayURI)
		} else {
			fmt.Fprintf(&b, "lookup %s: %s\n\n", name, err.Error())
		}
	}
	switch {
	case strings.Contains(low, "prior") || strings.Contains(low, "patent") || strings.Contains(low, "copyright"):
		fmt.Fprintf(&b, "Prior art: hash the work (sha256), stamp on /prior with grams, optionally fund the kaspa:p lock with 0.5 KAS. First UTXO is first-to-file on this template. Not a USPTO filing. Not Instant IP.\n")
	case strings.Contains(low, "vault") || strings.Contains(low, "234"):
		v := framing.Demo()
		fmt.Fprintf(&b, "Vault bump: real amount %d, hostile packing reads %d. Same 42 bytes. Vaults still lock KAS. Grams pay this framing action.\n", v.Canonical.RealAmount, v.Attack.VaultAmount)
	case strings.Contains(low, "postage") || strings.Contains(low, "stamp") || strings.Contains(low, "message"):
		b.WriteString("Postage bills grams for a sequenced envelope. Not KaChat. Not broadcast. Encryption stays with the messenger.\n")
	default:
		led := seq.Snap()
		fmt.Fprintf(&b, "Gramlane is an L1 work-credit lane. Remaining %d / %d grams. Jobs: vault bump, postage, this agent, .kas site builder. Not a dollar. USDT can sit off to the side.\n", led.Remaining, led.Credits)
	}
	return &Reply{
		Model:   ModelLocal,
		Text:    strings.TrimSpace(b.String()),
		Grok:    false,
		Name:    name,
		Warning: "No XAI_API_KEY in this process. Local tools only. Not a fake LLM.",
	}
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
