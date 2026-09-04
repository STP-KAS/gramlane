// Package kachat is Gramlane’s KaChat surface: resolve a .kas name to a contact.
// KaChat (kachat.org, @KaChat_) seals E2E as ciph_msg:1: in the wallet.
// This process does not encrypt. Grams pay discovery and postage.
package kachat

import (
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	Docs     = "https://www.kachat.org/"
	X        = "https://x.com/KaChat_"
	Linktree = "https://linktr.ee/KaChat_"
	IOS      = "https://apps.apple.com/us/app/kachat/id6759102359"
	Android  = "https://play.google.com/store/apps/details?id=com.kachat.app"
	GitHub   = "https://github.com/vsmirn0v/KaChat"
)

type Contact struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	PayURI   string `json:"payUri,omitempty"`
	Note     string `json:"note"`
	Initial  string `json:"initial"`
	DeepLink string `json:"deepLink"`
}

type Envelope struct {
	Kind     string `json:"kind"`
	Template string `json:"template"`
	Fill     string `json:"fill"`
}

func ContactFrom(name, addr string) Contact {
	pay := strings.TrimSpace(addr)
	if pay != "" && !strings.HasPrefix(pay, "kaspa:") {
		pay = "kaspa:" + strings.TrimPrefix(pay, "kaspa:")
	}
	return Contact{
		Name:     name,
		Address:  addr,
		PayURI:   pay,
		Note:     "The wallet is the person. Encryption stays in the KaChat app.",
		Initial:  Initial(name),
		DeepLink: DeepLink(addr),
	}
}

func Initial(name string) string {
	n := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".kas")
	n = strings.TrimPrefix(n, "#")
	if n == "" {
		return "K"
	}
	return strings.ToUpper(string([]rune(n)[0:1]))
}

func Envelopes(alias string) []Envelope {
	if alias == "" {
		alias = "alias"
	}
	b64 := base64.StdEncoding.EncodeToString([]byte("<ciphertext>"))
	return []Envelope{
		{Kind: "comm", Template: fmt.Sprintf("ciph_msg:1:comm:%s:%s", alias, b64), Fill: "ECDH secp256k1 + HKDF-SHA256 + ChaCha20-Poly1305 in the KaChat client. This string is a shape, not a sealed message."},
		{Kind: "pay", Template: "ciph_msg:1:pay:{encrypted_hex}", Fill: "Optional encrypted payment memo. Amount stays visible on-chain. Gramlane Pay is the till; this is the memo shape."},
		{Kind: "handshake", Template: "ciph_msg:1:handshake:{encrypted_bytes}", Fill: "Contact signaling. Routing aliases are derived, not trusted as identity."},
		{Kind: "self_stash", Template: "ciph_msg:1:self_stash:{scope}:{encrypted_hex}", Fill: "Self-stored handshake data."},
		{Kind: "group", Template: "kchat:1:gcomm / gctl", Fill: "Group chat payloads (KaChat 3+/4.0)."},
	}
}

func DeepLink(addr string) string {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return Linktree
	}
	return "https://www.kachat.org/?to=" + addr
}
