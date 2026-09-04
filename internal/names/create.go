package names

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Official KNS reveal fee address (mainnet). We do not collect this.
const InscribeFeeAddress = "kaspa:qyp4nvaq3pdq7609z09fvdgwtc9c7rg07fuw5zgeee7xpr085de59eseqfcmynn"

// Official inscriber. Gramlane checks and explains; it does not broadcast.
const InscribeURL = "https://app.knsdomains.org/"

var labelRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// Want is a name someone is trying to get. Plain fields for the Kasdomain tab.
type Want struct {
	Query       string `json:"query"`
	Label       string `json:"label"`
	Name        string `json:"name"`
	Valid       bool   `json:"valid"`
	Why         string `json:"why,omitempty"`
	Available   bool   `json:"available"`
	Taken       bool   `json:"taken"`
	Owner       string `json:"owner,omitempty"`
	Mine        bool   `json:"mine"`
	PriceKAS    int    `json:"priceKas,omitempty"`
	PriceNote   string `json:"priceNote,omitempty"`
	Payload     string `json:"payload,omitempty"`
	FeeAddress  string `json:"feeAddress,omitempty"`
	InscribeURL string `json:"inscribeUrl,omitempty"`
	Note        string `json:"note"`
}

func labelOf(raw string) string {
	n := strings.ToLower(strings.TrimSpace(raw))
	n = strings.TrimPrefix(n, "kas://")
	n = strings.TrimPrefix(n, "did:kas:")
	n = strings.TrimSuffix(n, "/")
	n = strings.TrimSuffix(n, ".kas")
	if i := strings.IndexByte(n, '.'); i >= 0 {
		n = n[:i]
	}
	return n
}

func ValidLabel(label string) (bool, string) {
	if label == "" {
		return false, "Type a word. Example: bakery"
	}
	if strings.ContainsAny(label, " .") {
		return false, "One word only. No spaces and no extra dots. Try bakery, not pay.bakery"
	}
	if utf8.RuneCountInString(label) > 63 {
		return false, "That word is too long. Keep it under 64 letters."
	}
	if !labelRe.MatchString(label) {
		return false, "Use letters and numbers. You may put a hyphen in the middle. No emoji here — keep it easy to type."
	}
	return true, ""
}

func PriceKAS(label string) int {
	n := utf8.RuneCountInString(label)
	switch {
	case n <= 2:
		return 4200
	case n == 3:
		return 2100
	case n == 4:
		return 525
	default:
		return 35
	}
}

func priceNote(label string) string {
	n := utf8.RuneCountInString(label)
	switch {
	case n <= 2:
		return "Very short names cost more — like a short car plate. Paid once. No yearly bill."
	case n == 3:
		return "Three-letter names cost more. Paid once. No yearly bill."
	case n == 4:
		return "Four-letter names are in the middle. Paid once. No yearly bill."
	default:
		return "Five letters or more: 35 KAS, once. No yearly bill."
	}
}

func ParseWant(raw string) *Want {
	w := &Want{
		Query:       strings.TrimSpace(raw),
		InscribeURL: InscribeURL,
		FeeAddress:  InscribeFeeAddress,
		Note:        "Gramlane does not mint the name. The official name shop writes it to Kaspa. First person to pay wins. Not a unique lock in consensus.",
	}
	w.Label = labelOf(raw)
	core := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(raw)), ".kas")
	if strings.Contains(core, ".") {
		w.Valid = false
		w.Why = "One word only. No spaces and no extra dots. Try bakery, not pay.bakery"
		return w
	}
	ok, why := ValidLabel(w.Label)
	w.Valid = ok
	w.Why = why
	if !ok {
		return w
	}
	w.Name = w.Label + ".kas"
	w.PriceKAS = PriceKAS(w.Label)
	w.PriceNote = priceNote(w.Label)
	b, _ := json.Marshal(map[string]string{"op": "create", "p": "domain", "v": w.Label})
	w.Payload = string(b)
	return w
}

// Check asks the public name index if the word is free.
func Check(raw, myAddr string) (*Want, error) {
	w := ParseWant(raw)
	if !w.Valid {
		return w, nil
	}
	p, err := Lookup(w.Name)
	if err != nil {
		if isFreeOnIndex(err) {
			w.Available = true
			return w, nil
		}
		return w, fmt.Errorf("could not ask the name list right now. Try again in a moment")
	}
	if p == nil || p.Owner == "" {
		w.Available = true
		return w, nil
	}
	w.Taken = true
	w.Owner = p.Owner
	w.Mine = strings.EqualFold(strings.TrimSpace(myAddr), p.Owner)
	if w.Mine {
		w.Note = "You already own this name. Hang it as your Gramlane sign."
	} else {
		w.Note = "Someone else has this sign. Try another word, or buy one on the market for grams."
	}
	return w, nil
}
