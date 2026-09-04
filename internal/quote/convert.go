package quote

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Convert is the KIP-21 policy meter: KAS ↔ sompi ↔ grams.
// Not a market, not a peg, not a KCC-20.
type Convert struct {
	Source       string  `json:"source"`
	Grams        uint64  `json:"grams"`
	Sompi        uint64  `json:"sompi"`
	Dust         uint64  `json:"dustSompi"`
	KAS          float64 `json:"kas"`
	KASText      string  `json:"kasText"`
	KAS1         string  `json:"kas1"`
	SompiPerGram uint64  `json:"sompiPerGram"`
	SompiPerKAS  uint64  `json:"sompiPerKas"`
	USD          string  `json:"usd"`
	KCC20        bool    `json:"kcc20"`
	Note         string  `json:"note"`
}

func FromGrams(n uint64) (Convert, error) {
	if n > math.MaxUint64/SompiPerGram {
		return Convert{}, fmt.Errorf("overflow")
	}
	return pack("grams", n, n*SompiPerGram, 0)
}

func FromSompi(n uint64) (Convert, error) {
	return pack("sompi", n/SompiPerGram, n, n%SompiPerGram)
}

func FromKAS(s string) (Convert, error) {
	sompi, err := ParseKAS(s)
	if err != nil {
		return Convert{}, err
	}
	return FromSompi(sompi)
}

func Parse(kind, amount string) (Convert, error) {
	kind = strings.ToLower(strings.TrimSpace(kind))
	amount = strings.TrimSpace(amount)
	if amount == "" {
		return Convert{}, fmt.Errorf("amount")
	}
	switch kind {
	case "", "kas", "kaspa":
		c, err := FromKAS(amount)
		if err != nil {
			return Convert{}, err
		}
		c.Source = "kas"
		return c, nil
	case "gram", "grams", "credit", "credits":
		n, err := strconv.ParseUint(strings.ReplaceAll(amount, ",", ""), 10, 64)
		if err != nil {
			return Convert{}, fmt.Errorf("grams")
		}
		return FromGrams(n)
	case "sompi":
		n, err := strconv.ParseUint(strings.ReplaceAll(amount, ",", ""), 10, 64)
		if err != nil {
			return Convert{}, fmt.Errorf("sompi")
		}
		return FromSompi(n)
	default:
		return Convert{}, fmt.Errorf("unit: kas, grams, or sompi")
	}
}

func ParseKAS(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "KAS")
	s = strings.TrimSuffix(s, "kas")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	if s == "" || strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("kas")
	}
	parts := strings.SplitN(s, ".", 2)
	whole := parts[0]
	if whole == "" {
		whole = "0"
	}
	w, err := strconv.ParseUint(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("kas")
	}
	var frac uint64
	if len(parts) == 2 {
		f := parts[1]
		if len(f) > 8 {
			f = f[:8]
		}
		for len(f) < 8 {
			f += "0"
		}
		frac, err = strconv.ParseUint(f, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("kas")
		}
	}
	if w > math.MaxUint64/SompiPerKAS {
		return 0, fmt.Errorf("overflow")
	}
	out := w * SompiPerKAS
	if out+frac < out {
		return 0, fmt.Errorf("overflow")
	}
	return out + frac, nil
}

// Aside splits a KAS pile: hold (volatile) vs spend (prepaid grams).
type Aside struct {
	Percent    uint64  `json:"percent"`
	TotalSompi uint64  `json:"totalSompi"`
	Hold       Convert `json:"hold"`
	Spend      Convert `json:"spend"`
	PaySompi   uint64  `json:"paySompi"`
	PayKASText string  `json:"payKasText"`
	FloorNote  string  `json:"floorNote,omitempty"`
}

func SetAside(totalKAS string, pct uint64) (Aside, error) {
	if pct == 0 || pct > 100 {
		pct = 50
	}
	tot, err := ParseKAS(totalKAS)
	if err != nil {
		return Aside{}, err
	}
	spendS := tot * pct / 100
	holdS := tot - spendS
	spend, err := FromSompi(spendS)
	if err != nil {
		return Aside{}, err
	}
	hold, err := FromSompi(holdS)
	if err != nil {
		return Aside{}, err
	}
	pay := spendS
	note := ""
	if spendS > 0 && spendS < KaswareMinSompi {
		pay = KaswareMinSompi
		note = "Kasware min send is 0.5 KAS. Your share at policy is smaller; sending 0.5 KAS mints 500000 grams into the spend pile."
	}
	return Aside{
		Percent:    pct,
		TotalSompi: tot,
		Hold:       hold,
		Spend:      spend,
		PaySompi:   pay,
		PayKASText: kasText(pay),
		FloorNote:  note,
	}, nil
}

// FiatBill is a merchant shelf price → grams.
// The EUR/USD rate is the sign on the counter, not an oracle and not a peg.
type FiatBill struct {
	Amount    string `json:"amount"`
	Ccy       string `json:"ccy"`
	KasInFiat string `json:"kasInFiat"`
	Grams     uint64 `json:"grams"`
	KASText   string `json:"kasText"`
	Sompi     uint64 `json:"sompi"`
	Label     string `json:"label"`
	Note      string `json:"note"`
}

func FromFiat(amount, ccy, kasInFiat string) (FiatBill, error) {
	ccy = strings.ToUpper(strings.TrimSpace(ccy))
	if ccy != "EUR" && ccy != "USD" {
		return FiatBill{}, fmt.Errorf("currency: EUR or USD")
	}
	amt, err := parseDec(amount, 6)
	if err != nil || amt == 0 {
		return FiatBill{}, fmt.Errorf("price")
	}
	rate, err := parseDec(kasInFiat, 6)
	if err != nil || rate == 0 {
		return FiatBill{}, fmt.Errorf("rate: 1 KAS in %s", ccy)
	}
	if amt > math.MaxUint64/SompiPerKAS {
		return FiatBill{}, fmt.Errorf("overflow")
	}
	sompi := amt * SompiPerKAS / rate
	c, err := FromSompi(sompi)
	if err != nil {
		return FiatBill{}, err
	}
	label := strings.TrimSpace(amount) + " " + ccy
	return FiatBill{
		Amount:    strings.TrimSpace(amount),
		Ccy:       ccy,
		KasInFiat: strings.TrimSpace(kasInFiat),
		Grams:     c.Grams,
		KASText:   c.KASText,
		Sompi:     c.Sompi,
		Label:     label,
		Note:      "Merchant sign, not an oracle. 1 KAS = " + strings.TrimSpace(kasInFiat) + " " + ccy + " on this till. Settlement is grams at 100 sompi/gram. Not a dollar peg.",
	}, nil
}

// FromUSD is a board shelf in dollars that settles in KAS on L1.
// Not a peg. Not grams. Desk jobs that sequence work stay in grams.
func FromUSD(amount, kasInUSD string) (FiatBill, error) {
	b, err := FromFiat(amount, "USD", kasInUSD)
	if err != nil {
		return FiatBill{}, err
	}
	b.Note = "Shelf in USD. Settlement is KAS on L1. 1 KAS = " + strings.TrimSpace(kasInUSD) + " USD on this board. Sign, not an oracle. Not a dollar peg. Desk jobs stay in grams."
	return b, nil
}

// USDCents parses a dollar string ("1", "1.00", "$0.50") into cents.
func USDCents(amount string) (uint64, error) {
	amount = strings.TrimSpace(amount)
	amount = strings.TrimPrefix(amount, "$")
	n, err := parseDec(amount, 2)
	if err != nil || n == 0 {
		return 0, fmt.Errorf("usd")
	}
	return n, nil
}

func FormatUSD(cents uint64) string {
	return fmt.Sprintf("$%d.%02d", cents/100, cents%100)
}

func FormatUSDAmount(cents uint64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

func FloorPay(sompi uint64) uint64 {
	if sompi < KaswareMinSompi {
		return KaswareMinSompi
	}
	return sompi
}

func parseDec(s string, scale int) (uint64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, " ", "")
	if s == "" || strings.HasPrefix(s, "-") {
		return 0, fmt.Errorf("number")
	}
	parts := strings.SplitN(s, ".", 2)
	whole := parts[0]
	if whole == "" {
		whole = "0"
	}
	w, err := strconv.ParseUint(whole, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("number")
	}
	var frac uint64
	if len(parts) == 2 {
		f := parts[1]
		if len(f) > scale {
			f = f[:scale]
		}
		for len(f) < scale {
			f += "0"
		}
		frac, err = strconv.ParseUint(f, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("number")
		}
	}
	base := uint64(1)
	for i := 0; i < scale; i++ {
		if base > math.MaxUint64/10 {
			return 0, fmt.Errorf("overflow")
		}
		base *= 10
	}
	if w > math.MaxUint64/base {
		return 0, fmt.Errorf("overflow")
	}
	out := w * base
	if out+frac < out {
		return 0, fmt.Errorf("overflow")
	}
	return out + frac, nil
}

func pack(src string, grams, sompi, dust uint64) (Convert, error) {
	return Convert{
		Source:       src,
		Grams:        grams,
		Sompi:        sompi,
		Dust:         dust,
		KAS:          float64(sompi) / float64(SompiPerKAS),
		KASText:      kasText(sompi),
		KAS1:         FormatKAS1(sompi),
		SompiPerGram: SompiPerGram,
		SompiPerKAS:  SompiPerKAS,
		USD:          "not quoted",
		KCC20:        false,
		Note:         "Policy rate 100 sompi/gram (KIP-21 min-relay, not consensus). Not a market. Not a KCC-20. Not USD. WorkCredit already is the voucher; a token would be the wrong object.",
	}, nil
}
