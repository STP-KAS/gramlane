package names

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"gramlane/internal/appenv"
	"gramlane/internal/quote"
)

// Sign is 1 KAS in USD on this name board. Operator sign, not an oracle, not a peg.
type Sign struct {
	Ccy       string `json:"ccy"`
	KasInFiat string `json:"kasInFiat"`
}

// Ask is a kasdomain shelf in USD that settles in KAS on L1.
// Recurrent desk jobs stay in grams. Acquiring or buying a name is a Kaspa transaction.
type Ask struct {
	Name     string `json:"name,omitempty"`
	USDCents uint64 `json:"usdCents"`
	USD      string `json:"usd"`
	Amount   string `json:"amount"`
	Rate     string `json:"rate"`
	Sompi    uint64 `json:"sompi"`
	KASText  string `json:"kasText"`
	PaySompi uint64 `json:"paySompi"`
	PayKAS   string `json:"payKas"`
	Note     string `json:"note"`
}

var signPath = appenv.File("name-rate.json")

func ResetSignForTest(dir string) {
	signPath = filepath.Join(dir, "name-rate.json")
}

func LoadSign() Sign {
	b, err := os.ReadFile(signPath)
	if err != nil {
		return Sign{Ccy: "USD", KasInFiat: "0.10"}
	}
	var s Sign
	if json.Unmarshal(b, &s) != nil || s.Ccy == "" || s.KasInFiat == "" {
		return Sign{Ccy: "USD", KasInFiat: "0.10"}
	}
	s.Ccy = strings.ToUpper(strings.TrimSpace(s.Ccy))
	if s.Ccy != "USD" {
		s.Ccy = "USD"
	}
	return s
}

// SuggestUSDCents is a fair, affordable list price in USD cents.
// Short names cost more. Long everyday names stay cheap so extras can sit on the market.
func SuggestUSDCents(name string) uint64 {
	n := apex(name)
	if n == "" {
		return 15
	}
	L := utf8.RuneCountInString(n)
	var base uint64
	switch {
	case L <= 1:
		base = 800 // $8.00 — rare, still far under KNS 4200 KAS
	case L == 2:
		base = 250
	case L == 3:
		base = 100
	case L == 4:
		base = 50
	case L <= 6:
		base = 25
	default:
		base = 15
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
	if g < 10 {
		g = 10
	}
	return g
}

func Settle(cents uint64) Ask {
	sign := LoadSign()
	amount := quote.FormatUSDAmount(cents)
	bill, err := quote.FromUSD(amount, sign.KasInFiat)
	if err != nil {
		pay := quote.KaswareMinSompi
		return Ask{
			USDCents: cents,
			USD:      quote.FormatUSD(cents),
			Amount:   amount,
			Rate:     sign.KasInFiat,
			Sompi:    pay,
			KASText:  "0.5",
			PaySompi: pay,
			PayKAS:   "0.5",
			Note:     "Could not convert USD to KAS at this sign. Wallet floor 0.5 KAS.",
		}
	}
	pay := quote.FloorPay(bill.Sompi)
	kas := bill.KASText
	if pay != bill.Sompi {
		if c, e := quote.FromSompi(pay); e == nil {
			kas = c.KASText
		}
	}
	return Ask{
		USDCents: cents,
		USD:      quote.FormatUSD(cents),
		Amount:   amount,
		Rate:     sign.KasInFiat,
		Sompi:    bill.Sompi,
		KASText:  bill.KASText,
		PaySompi: pay,
		PayKAS:   kas,
		Note:     bill.Note,
	}
}

func AskFor(name string) Ask {
	a := Settle(SuggestUSDCents(name))
	a.Name = DisplayName(name)
	return a
}

func FundQuote() quote.Quote {
	m := FundMint()
	return quote.Quote{
		Unit:       "KAS",
		Sompi:      m.NameSompi,
		KAS:        100,
		PaySompi:   m.NameSompi,
		PayKAS:     100,
		PayKASText: m.TotalKAS,
		Lane:       "kasdomain",
		Scheme:     "kaspa-l1",
		USD:        "shelf",
		Note:       m.Note,
	}
}

// AdoptionVault receives half of every name mint. Ideas, dApps, Kaspa growth.
// Not Gramlane's till. Not a protocol tax. Override with ADOPTION_VAULT.
const AdoptionVault = "kaspa:qzpvdakagvwfm95g8pv9ndpupjtndgjfhmve08cg3tv5wgfytjzf7e6t4puat"

type Mint struct {
	NameSompi  uint64 `json:"nameSompi"`
	VaultSompi uint64 `json:"vaultSompi"`
	TotalSompi uint64 `json:"totalSompi"`
	NameKAS    string `json:"nameKas"`
	VaultKAS   string `json:"vaultKas"`
	TotalKAS   string `json:"totalKas"`
	Vault      string `json:"vault"`
	SharePct   uint64 `json:"sharePct"`
	Note       string `json:"note"`
}

func VaultAddress() string {
	if v := strings.TrimSpace(os.Getenv("ADOPTION_VAULT")); strings.HasPrefix(v, "kaspa:") {
		return v
	}
	return AdoptionVault
}

const (
	MintKAS      uint64 = 200
	MintSharePct uint64 = 50
)

func FundMint() Mint {
	total := MintKAS * quote.SompiPerKAS
	half := total / 2
	return Mint{
		NameSompi:  half,
		VaultSompi: half,
		TotalSompi: total,
		NameKAS:    "100",
		VaultKAS:   "100",
		TotalKAS:   "200",
		Vault:      VaultAddress(),
		SharePct:   MintSharePct,
		Note:       "Standard mint is 200 KAS. Half funds the name output. Half goes to a Kaspa address for ideas and dApps that grow Kaspa. Not a protocol tax. Not grams.",
	}
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
