// Package quote prices sequenced L1 work in grams.
// 1 credit = 1 KIP-21 gram. Not USD. No L2.
package quote

import (
	"fmt"
	"math"
)

const (
	SompiPerGram uint64 = 100
	SompiPerKAS  uint64 = 100_000_000
	Unit         string = "gram"
)

type Quote struct {
	Grams        uint64  `json:"grams"`
	Credits      uint64  `json:"credits"`
	Unit         string  `json:"unit"`
	SompiPerGram uint64  `json:"sompiPerGram"`
	Sompi        uint64  `json:"sompi"`
	KAS          float64 `json:"kas"`
	Lane         string  `json:"lane,omitempty"`
	Scheme       string  `json:"scheme"`
	USD          string  `json:"usd"`
	Note         string  `json:"note"`
}

func Grams(n uint64, lane string) (Quote, error) {
	if n == 0 {
		return Quote{}, fmt.Errorf("grams must be > 0")
	}
	if n > math.MaxUint64/SompiPerGram {
		return Quote{}, fmt.Errorf("overflow")
	}
	sompi := n * SompiPerGram
	return Quote{
		Grams:        n,
		Credits:      n,
		Unit:         Unit,
		SompiPerGram: SompiPerGram,
		Sompi:        sompi,
		KAS:          float64(sompi) / float64(SompiPerKAS),
		Lane:         lane,
		Scheme:       "kaspa-work-credit",
		USD:          "not quoted",
		Note:         "Prepaid L1 grams. Covenant WorkCredit. Miners still take KAS. No L2. Not a stablecoin.",
	}, nil
}
