package names

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SuggestGrams is a fair, affordable list price in grams (not USD, not KNS KAS).
// Short names cost more. Long everyday names stay cheap so extras can sit on the market.
func SuggestGrams(name string) uint64 {
	n := apex(name)
	if n == "" {
		return 25_000
	}
	L := utf8.RuneCountInString(n)
	var base uint64
	switch {
	case L <= 1:
		base = 8_000_000 // ~8 KAS at policy — rare, still far under KNS 4200 KAS
	case L == 2:
		base = 2_000_000
	case L == 3:
		base = 500_000
	case L == 4:
		base = 150_000
	case L <= 6:
		base = 50_000
	default:
		base = 25_000
	}
	mult := 1.0
	if allDigits(n) {
		mult *= 0.5
	}
	if strings.Contains(n, "-") {
		mult *= 0.8
	}
	if repeating(n) {
		mult *= 0.7
	}
	if hasVowel(n) && L >= 3 && L <= 8 {
		mult *= 1.15
	}
	if commonWord(n) {
		mult *= 1.4
	}
	g := uint64(float64(base)*mult + 0.5)
	if g < 10_000 {
		g = 10_000
	}
	return g
}

func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return len(s) > 0
}

func repeating(s string) bool {
	if len(s) < 2 {
		return false
	}
	c := s[0]
	for i := 1; i < len(s); i++ {
		if s[i] != c {
			return false
		}
	}
	return true
}

func hasVowel(s string) bool {
	return strings.ContainsAny(s, "aeiouy")
}

func commonWord(s string) bool {
	switch s {
	case "shop", "pay", "cafe", "bakery", "city", "home", "work", "mail",
		"news", "bank", "gold", "food", "book", "game", "play", "love",
		"kaspa", "gram", "node", "mine":
		return true
	}
	return false
}
