// Package genesis is the WorkCredit instance tied to the 0.5 KAS L1 sale.
// Template hash is the code. Constructor state is this issuer/holder/credits/lane.
// This process does not broadcast the covenant UTXO. Kasware sendKaspa cannot lock a script.
package genesis

import (
	"encoding/hex"
	"os"
)

const (
	SaleTx          = "c1799b0de40f71cfd7a153684ef22326ad920d0dca2a8b519ce2c8379c4f7bc2"
	Explorer        = "https://explorer.kaspa.org/txs/" + SaleTx
	DeskAddress     = "kaspa:qpjm8kzpcj5he3hg9msrdc78a3k46zda866pucwetprgtgc7s3ry2kq38atpq"
	HolderAddress   = "kaspa:qrw0tuxnjukkrk7lm5v7chx4qpu8f8q0janp7jdk80z4tc8anu62g62s4zsj0"
	IssuerPubHex    = "65b3d841c4a97cc6e82ee036e3c7ec6d5d09bd3eb41e61d9584685a31e844645"
	HolderPubHex    = "dcf5f0d3972d61dbdfdd19ec5cd50078749c0f97661f49b63bc555e0fd9f34a4"
	Credits         = 500_000
	Lane            = "SEQ1"
	SaleSompi       = 50_000_000
	TemplateHashHex = "c61458da2134faad500c05cc35fb5dfd2db7e44538782703207f14926a11521e"
	Artifact        = "contracts/v1/WorkCredit-live.json"
)

type Plan struct {
	SaleTx         string `json:"saleTx"`
	Explorer       string `json:"explorer"`
	Desk           string `json:"desk"`
	Holder         string `json:"holder"`
	IssuerPub      string `json:"issuerPubXOnly"`
	HolderPub      string `json:"holderPubXOnly"`
	Credits        int    `json:"credits"`
	Lane           string `json:"lane"`
	SaleKAS        string `json:"saleKas"`
	TemplateHash   string `json:"templateHash"`
	Artifact       string `json:"artifact"`
	OnChainSale    bool   `json:"onChainSale"`
	VoucherOnChain bool   `json:"voucherOnChain"`
	Note           string `json:"note"`
	CtorPath       string `json:"ctorPath"`
}

func Live() Plan {
	art := Artifact
	if _, err := os.Stat(art); err != nil {
		art = ""
	}
	return Plan{
		SaleTx:         SaleTx,
		Explorer:       Explorer,
		Desk:           DeskAddress,
		Holder:         HolderAddress,
		IssuerPub:      IssuerPubHex,
		HolderPub:      HolderPubHex,
		Credits:        Credits,
		Lane:           Lane,
		SaleKAS:        "0.5",
		TemplateHash:   TemplateHashHex,
		Artifact:       art,
		OnChainSale:    true,
		VoucherOnChain: false,
		CtorPath:       "contracts/v1/ctor-WorkCredit-live.json",
		Note:           "0.5 KAS sale is on L1. WorkCredit.sil compiled with those two x-only pubkeys and 500000 grams on lane SEQ1. Template hash is the code (same for every WorkCredit). Instance state is in the artifact. Broadcasting the covenant UTXO is not sendKaspa — it is a Toccata P2SH deploy. Not done.",
	}
}

func MustHex(h string) []byte {
	b, err := hex.DecodeString(h)
	if err != nil {
		panic(err)
	}
	return b
}
